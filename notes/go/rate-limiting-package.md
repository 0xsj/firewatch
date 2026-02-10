# golang.org/x/time/rate Package

## Overview

The `golang.org/x/time/rate` package provides efficient token bucket
rate limiting. It's part of the official Go extended libraries and is
production-ready for high-concurrency scenarios.

Package: `golang.org/x/time/rate`

---

## Core Type: rate.Limiter

```go
import "golang.org/x/time/rate"

// Create a limiter: 10 requests/second, burst of 20
limiter := rate.NewLimiter(10, 20)
//                         ^rate  ^burst

// Check if request is allowed
if limiter.Allow() {
    // Token available, request proceeds
} else {
    // No tokens, request rejected
}
```

### Parameters

**rate.Limit** (requests per second):
- Type: `float64` (can be fractional)
- `10` = 10 requests/second
- `0.5` = 1 request every 2 seconds
- `rate.Inf` = unlimited

**burst** (bucket capacity):
- Type: `int`
- Maximum tokens that can accumulate
- Allows short traffic spikes

---

## Token Bucket Mechanics

### Refill Behavior

Tokens refill continuously, not in discrete steps:

```go
limiter := rate.NewLimiter(10, 5)  // 10/sec, burst 5

// t=0.0s: 5 tokens (full bucket)
limiter.Allow()  // t=0.0s: true, 4 tokens remain

// t=0.1s: 4 + (0.1s × 10/s) = 5 tokens (capped at burst)
limiter.Allow()  // t=0.1s: true, 4 tokens remain

// t=0.5s: 4 + (0.5s × 10/s) = 5 tokens (capped)
```

**Key insight:** Refill is continuous, not periodic. You don't wait
for the next "second" to tick over.

### Burst Capacity

Burst allows temporary spikes above the sustained rate:

```go
limiter := rate.NewLimiter(1, 10)  // 1/sec sustained, burst 10

// Send 10 instant requests (burst capacity)
for i := 0; i < 10; i++ {
    limiter.Allow()  // All return true
}

limiter.Allow()  // False (bucket empty)

time.Sleep(5 * time.Second)  // Refill 5 tokens

for i := 0; i < 5; i++ {
    limiter.Allow()  // True (5 tokens available)
}
```

---

## API Methods

### Allow() — Non-Blocking Check

```go
if limiter.Allow() {
    // Consume 1 token, proceed immediately
} else {
    // No tokens available, reject
}
```

**Use case:** HTTP servers, where you want instant allow/deny.

### Wait(ctx) — Blocking Wait

```go
err := limiter.Wait(ctx)
if err != nil {
    // Context canceled or deadline exceeded
}
// Token consumed, proceed
```

**Use case:** Background workers that can afford to wait.

**Behavior:**
- Blocks until a token is available
- Respects context cancellation
- Returns immediately if token is ready

### Reserve() — Advanced Usage

```go
reservation := limiter.Reserve()
if !reservation.OK() {
    // Rate is Inf or limiter is full
    return
}

// Wait for reservation
time.Sleep(reservation.Delay())

// Or cancel reservation
reservation.Cancel()
```

**Use case:** Conditional rate limiting (reserve now, decide later).

### AllowN(n) — Consume Multiple Tokens

```go
// Consume 5 tokens at once
if limiter.AllowN(time.Now(), 5) {
    // Process 5 items
}
```

**Use case:** Batch operations where each batch counts as N requests.

---

## Rate Limit Patterns

### Per-Resource Rate Limiting

Each resource (user, IP, API key) gets its own limiter:

```go
type RateLimiter struct {
    mu       sync.RWMutex
    limiters map[string]*rate.Limiter
}

func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
    rl.mu.RLock()
    limiter, exists := rl.limiters[key]
    rl.mu.RUnlock()

    if exists {
        return limiter
    }

    // Create new limiter
    rl.mu.Lock()
    defer rl.mu.Unlock()

    // Double-check after acquiring write lock
    if limiter, exists := rl.limiters[key]; exists {
        return limiter
    }

    limiter = rate.NewLimiter(10, 20)
    rl.limiters[key] = limiter
    return limiter
}
```

**Pattern:** Lazy initialization with double-checked locking.

### Global Rate Limiting

Single shared limiter for all requests:

```go
var globalLimiter = rate.NewLimiter(1000, 2000)

func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !globalLimiter.Allow() {
            http.Error(w, "Too Many Requests", 429)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**Use case:** Protect server from total overload (all sources combined).

### Dynamic Rate Adjustment

Change rate limits at runtime:

```go
limiter := rate.NewLimiter(10, 20)

// Later: make it stricter
limiter.SetLimit(1)
limiter.SetBurst(5)
```

**Use case:** Adaptive rate limiting during attacks.

---

## Concurrency Safety

All methods are safe for concurrent use:

```go
var limiter = rate.NewLimiter(100, 200)

// Safe to call from multiple goroutines
go func() { limiter.Allow() }()
go func() { limiter.Allow() }()
go func() { limiter.Allow() }()
```

**Implementation:** Uses internal mutex, no external locking needed.

---

## Common Patterns

### HTTP Middleware

```go
func RateLimitMiddleware(limiter *rate.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                w.Header().Set("Retry-After", "1")
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Background Worker

```go
func worker(ctx context.Context, limiter *rate.Limiter) {
    for {
        // Wait for rate limit permission
        if err := limiter.Wait(ctx); err != nil {
            return  // Context canceled
        }

        // Process item
        processItem()
    }
}
```

### Retry with Backoff

```go
func retryWithRateLimit(ctx context.Context, limiter *rate.Limiter, fn func() error) error {
    for {
        if err := limiter.Wait(ctx); err != nil {
            return err
        }

        if err := fn(); err == nil {
            return nil
        }

        // Function failed, rate limiter ensures we don't retry too fast
    }
}
```

---

## Performance

### Overhead

**Per-call cost:**
- `Allow()`: ~50ns (single atomic operation)
- `Wait()`: ~50ns if token ready, blocks otherwise

**Memory:**
- Single `rate.Limiter`: ~80 bytes
- 10,000 limiters: ~800KB

**Benchmark:**
```
BenchmarkLimiter_Allow-8    30000000    50 ns/op    0 B/op    0 allocs/op
```

Negligible overhead for typical HTTP server workloads.

### Scaling

**Single shared limiter:**
- Handles millions of requests/second
- Slight contention at high concurrency (atomic operations)

**Per-resource limiters:**
- No contention (each resource has separate limiter)
- Scales linearly with goroutines

---

## Edge Cases

### Zero Rate

```go
limiter := rate.NewLimiter(0, 10)

limiter.Allow()  // true (uses burst)
limiter.Allow()  // true (uses burst)
// ... 10 times
limiter.Allow()  // false (burst exhausted, rate is 0 = no refill)
```

**Use case:** One-time allowances (burst only, no refill).

### Infinite Rate

```go
limiter := rate.NewLimiter(rate.Inf, 0)

limiter.Allow()  // Always true
```

**Use case:** Disable rate limiting dynamically.

### Fractional Rates

```go
// 1 request every 5 seconds
limiter := rate.NewLimiter(0.2, 1)

// 1 request every minute
limiter := rate.NewLimiter(1.0/60, 1)
```

**Use case:** Very slow rates (anti-abuse for expensive operations).

---

## Integration with Context

### Timeout on Wait

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := limiter.Wait(ctx); err != nil {
    // Waited 5 seconds, still no token available
    return err
}
```

### Graceful Shutdown

```go
func worker(ctx context.Context, limiter *rate.Limiter) {
    for {
        // Blocks until token available OR context canceled
        if err := limiter.Wait(ctx); err != nil {
            log.Println("Shutdown signal received")
            return
        }

        processItem()
    }
}

// Later: signal shutdown
cancel()
```

---

## Comparison with Alternatives

### vs. Third-Party Libraries

**golang.org/x/time/rate advantages:**
- Official Go extended library (well-maintained)
- Zero dependencies
- Battle-tested (used by Google, Kubernetes)
- Excellent performance

**Alternatives:**
- `github.com/juju/ratelimit`: Similar API, less popular
- `github.com/uber-go/ratelimit`: Simpler (no burst), less flexible

### vs. Manual Implementation

**Don't roll your own:**
- Token bucket math is subtle (continuous refill, edge cases)
- Concurrency-safe implementation is tricky
- Performance optimization is non-trivial

**Use rate.Limiter** unless you have very specific needs.

---

## Real-World Usage

### Kubernetes

Rate limits API server requests to prevent overload:

```go
// k8s.io/client-go uses golang.org/x/time/rate
limiter := rate.NewLimiter(50, 100)  // 50 QPS, burst 100
```

### Google Cloud SDK

Rate limits API calls to respect quotas:

```go
limiter := rate.NewLimiter(rate.Limit(quotaQPS), quotaBurst)
```

### gRPC

Controls client-side request pacing:

```go
// Per-connection rate limiting
conn.WithStreamInterceptor(rateLimitInterceptor(limiter))
```

---

## Tips & Tricks

### Choose Burst Based on Use Case

**Bursty traffic (web servers):**
```go
rate.NewLimiter(10, 50)  // High burst = allow page loads
```

**Steady traffic (background jobs):**
```go
rate.NewLimiter(10, 10)  // Low burst = smooth rate
```

**Rule of thumb:** Burst = 2× to 5× rate for typical web traffic.

### Monitor Token Availability

```go
// Check tokens without consuming
reservation := limiter.Reserve()
tokensAvailable := reservation.OK()
reservation.Cancel()  // Don't actually consume
```

### Combine with Circuit Breakers

```go
if limiter.Allow() && circuitBreaker.Allow() {
    // Request proceeds
}
```

**Use case:** Rate limiting + health checks.

---

## References

- [Package documentation](https://pkg.go.dev/golang.org/x/time/rate)
- [Source code](https://github.com/golang/time/tree/master/rate)
- [Token bucket algorithm](https://en.wikipedia.org/wiki/Token_bucket)

**Used in Firewatch:**
- `internal/middleware/ratelimit.go` — Per-IP rate limiting
- Pattern: Lazy initialization + TTL cleanup
- Config: 10 req/sec, burst 20, cleanup 5 min
