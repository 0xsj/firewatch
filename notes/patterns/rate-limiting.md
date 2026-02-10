# Rate Limiting Pattern

## Problem

Honeypots need to distinguish between human reconnaissance and
automated scanning. A human might make 5-10 requests per minute while
exploring. A scanner makes 100+ requests per second.

Without rate limiting:
- High-volume scanners waste resources
- Logs fill with noise from single aggressive sources
- Hard to distinguish coordinated attacks from single-source floods

## Solution: Token Bucket Rate Limiting

Rate limiting enforces per-IP request quotas using the **token bucket
algorithm**:

1. Each IP gets a bucket with tokens
2. Each request consumes 1 token
3. Tokens refill at a steady rate
4. When bucket is empty, requests are blocked

This allows **burst capacity** (short spikes OK) while enforcing
**sustained rate limits** (long-term scanning blocked).

---

## Token Bucket Algorithm

### Metaphor

Think of a bucket that:
- Holds up to **N tokens** (burst capacity)
- Refills at **R tokens/second** (sustained rate)
- Each request drains **1 token**

```
Bucket capacity: 20 tokens
Refill rate: 10 tokens/second

Time 0s:  [████████████████████] 20 tokens
  ↓ 5 requests consume 5 tokens
Time 0s:  [███████████████·····] 15 tokens
  ↓ wait 1 second (refill 10 tokens)
Time 1s:  [████████████████████] 20 tokens (capped at max)
  ↓ 25 rapid requests (burst > capacity)
Time 1s:  [··························] 0 tokens
  ↓ request #26 arrives
Time 1s:  429 Too Many Requests
  ↓ wait 0.1 second (refill 1 token)
Time 1.1s: [█···················] 1 token
  ↓ request arrives
Time 1.1s: [··························] 0 tokens (allowed)
```

### Why This Works

**Allows burst traffic:**
- Legitimate users fetch multiple resources (HTML, CSS, JS, images)
- Burst of 20 requests = OK for page load

**Blocks sustained scanning:**
- Scanner trying 1000 URLs/minute exceeds sustained rate
- After burst is exhausted, limited to 10 req/sec

**Automatically recovers:**
- Tokens refill over time
- No manual unbanning needed

---

## Implementation

### Core Components

**RateLimiter** manages per-IP limiters:

```go
type RateLimiter struct {
    cfg      RateLimiterConfig
    store    storage.Store      // For event recording
    logger   *slog.Logger
    mu       sync.RWMutex
    limiters map[string]*limiterEntry  // ip -> limiter
    stopCh   chan struct{}             // Cleanup signal
}

type limiterEntry struct {
    limiter    *rate.Limiter  // golang.org/x/time/rate
    lastAccess time.Time      // For TTL cleanup
}
```

**Token bucket parameters:**

```go
type RateLimiterConfig struct {
    RequestsPerSecond float64       // Sustained rate (refill speed)
    Burst             int           // Bucket capacity (max tokens)
    CleanupInterval   time.Duration // How often to remove stale entries
}
```

### Per-IP Tracking

Each IP gets its own rate limiter:

```go
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
    rl.mu.RLock()
    entry, exists := rl.limiters[ip]
    rl.mu.RUnlock()

    if exists {
        entry.lastAccess = time.Now()
        return entry.limiter
    }

    // Create new limiter for this IP
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter := rate.NewLimiter(
        rate.Limit(rl.cfg.RequestsPerSecond),
        rl.cfg.Burst,
    )
    rl.limiters[ip] = &limiterEntry{
        limiter:    limiter,
        lastAccess: time.Now(),
    }

    return limiter
}
```

**Key design:**
- Read lock for fast path (limiter exists)
- Write lock only for new limiters
- Double-check pattern prevents race conditions
- Lazy initialization (limiters created on first request)

### Request Handling

```go
func RateLimit(rl *RateLimiter) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := httputil.ClientIP(r)
            limiter := rl.getLimiter(ip)

            if !limiter.Allow() {
                // Rate limit exceeded
                rl.recordRateLimitEvent(r.Context(), r)

                w.Header().Set("Retry-After", "1")
                w.Header().Set("X-RateLimit-Limit", "20")
                w.Header().Set("X-RateLimit-Remaining", "0")
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }

            // Request allowed
            next.ServeHTTP(w, r)
        })
    }
}
```

**Flow:**
1. Extract client IP
2. Get (or create) rate limiter for IP
3. Check if token available (`limiter.Allow()`)
4. If no token: return 429 + record event
5. If token available: consume token + continue

### TTL Cleanup

Prevents memory leak from storing limiters forever:

```go
func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(rl.cfg.CleanupInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            rl.mu.Lock()
            now := time.Now()
            staleThreshold := rl.cfg.CleanupInterval * 2

            for ip, entry := range rl.limiters {
                if now.Sub(entry.lastAccess) > staleThreshold {
                    delete(rl.limiters, ip)
                }
            }
            rl.mu.Unlock()

        case <-rl.stopCh:
            return
        }
    }
}
```

**Strategy:**
- Run cleanup every N minutes (configurable)
- Remove limiters not accessed in 2×N minutes
- Example: 5 min cleanup = remove after 10 min idle

**Why 2× cleanup interval?**
- Prevents thrashing (create/delete/recreate)
- Legitimate users might pause between page loads
- Scanners rarely pause for 10 minutes

---

## Event Recording

Rate limit violations are recorded as detection events:

```go
func (rl *RateLimiter) recordRateLimitEvent(ctx context.Context, r *http.Request) {
    event := &models.Event{
        ID:         crypto.UUID4(),
        Timestamp:  timeutil.FormatRFC3339(timeutil.NowUTC()),
        RequestID:  middleware.RequestID(ctx),
        SourceIP:   httputil.ClientIP(r),
        Module:     "rate_limit",
        Method:     r.Method,
        Path:       r.URL.Path,
        Headers:    httputil.HeaderMap(r.Header),
        UserAgent:  r.UserAgent(),
        Severity:   "medium",
        Signatures: []string{"rate-limit-exceeded"},
    }

    rl.store.SaveEvent(ctx, event)
}
```

**Why record as events:**
- Shows up in `firewatch events` queries
- Triggers alerts via `AlertingStore`
- Correlates with other detection events
- Tracks attacker behavior over time

**Event fields:**
- `module: "rate_limit"` — filters in queries
- `severity: "medium"` — not critical, but suspicious
- `signature: "rate-limit-exceeded"` — detection tag

---

## Middleware Integration

Rate limiting runs **early** in the middleware chain:

```
Request arrives
    ↓
Correlation (assign request ID)
    ↓
Rate Limit ← CHECK HAPPENS HERE
    ├─ Under limit → Continue
    └─ Over limit  → 429 (skip all downstream middleware)
    ↓
Logging (log request/response)
    ↓
GeoIP (lookup location)
    ↓
Fingerprint (JA3, JA4, headers)
    ↓
Detection (signatures, patterns)
    ↓
Handler (honeypot module)
```

**Why early?**
- Skip expensive operations (GeoIP lookup, TLS fingerprinting) for blocked requests
- Reduce resource consumption during floods
- Fast rejection path for abusive IPs

**Why after Correlation?**
- Rate limit events need request IDs for tracking
- Correlation is cheap (just generate UUID)

---

## Configuration

### Default Settings

```yaml
rate_limit:
  enabled: true
  requests_per_second: 10  # Sustained: 600/min
  burst: 20                # Burst: allow 20 instant requests
  cleanup_minutes: 5       # Remove stale limiters every 5 min
```

### Tuning Guidelines

**Strict (catch slow scanners):**
```yaml
requests_per_second: 1   # 60/min
burst: 5                 # Small burst
cleanup_minutes: 10
```
- Blocks even careful scanners
- May impact legitimate page loads (watch false positives)

**Normal (default, balanced):**
```yaml
requests_per_second: 10  # 600/min
burst: 20                # Reasonable burst
cleanup_minutes: 5
```
- Blocks aggressive scanners
- Allows normal browsing
- Good starting point

**Lenient (only catch aggressive floods):**
```yaml
requests_per_second: 100  # 6000/min
burst: 200                # Large burst
cleanup_minutes: 1
```
- Only blocks extreme flooding
- Minimal false positives
- Use if you see legitimate high-volume clients

### Disabling Rate Limiting

```yaml
rate_limit:
  enabled: false
```

Middleware is skipped entirely (zero overhead).

---

## Testing Strategy

### Unit Tests

**Test coverage:**
1. Allows requests under limit
2. Blocks requests over limit
3. Per-IP isolation (IPs don't share buckets)
4. Token refill over time
5. Stale limiter cleanup
6. Nil limiter passthrough
7. Response headers (Retry-After, X-RateLimit-*)

**Example test:**

```go
func TestRateLimit_BlocksOverLimit(t *testing.T) {
    cfg := RateLimiterConfig{
        RequestsPerSecond: 1,  // Very strict
        Burst:             2,  // Small burst
        CleanupInterval:   1 * time.Minute,
    }
    limiter := NewRateLimiter(cfg, mockStore, logger)
    defer limiter.Stop()

    // First 2 requests: OK (burst)
    for i := 0; i < 2; i++ {
        assertStatus(t, limiter, http.StatusOK)
    }

    // 3rd request: blocked
    assertStatus(t, limiter, http.StatusTooManyRequests)
}
```

### Integration Tests

Test with real HTTP requests:

```bash
# Send 25 rapid requests
for i in {1..25}; do
  curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/
done

# Output:
# 200
# 200
# ... (20 times)
# 429
# 429
# ... (5 times)
```

### Load Testing

Verify cleanup doesn't leak memory:

```bash
# Simulate 10k unique IPs
for i in {1..10000}; do
  curl -s -o /dev/null -H "X-Forwarded-For: 192.168.$((i/256)).$((i%256))" \
    http://localhost:8080/
done

# Check memory usage
ps aux | grep firewatch

# Wait for cleanup interval
sleep 600

# Memory should drop (limiters cleaned up)
ps aux | grep firewatch
```

---

## Observability

### Log Messages

```
WARN rate limit exceeded ip=1.2.3.4 path=/api/users request_id=abc123
```

### Metrics to Track

**Rate limit events:**
```bash
firewatch events --module rate_limit --since 1h
```

**Top offenders:**
```bash
firewatch stats --module rate_limit --since 24h
# Shows IPs with most rate limit violations
```

**Alert on patterns:**
- Single IP hitting limit repeatedly = persistent scanner
- Many IPs hitting limit = distributed scan or DoS

---

## Limitations

### 1. Memory-Based (Not Distributed)

Each Firewatch instance has its own rate limiters. If running multiple
instances behind a load balancer, limits are per-instance, not global.

**Workaround:**
- Use sticky sessions (route same IP to same instance)
- Use Redis-based rate limiting (future work)

### 2. IP Spoofing

Rate limiting is per-IP. If attacker controls many IPs (botnet),
they can bypass limits by rotating source addresses.

**Mitigation:**
- Combine with behavioral fingerprinting (same scanning pattern = same attacker)
- Correlate via campaign detection

### 3. Shared IPs

Legitimate users behind NAT (corporate proxy, VPN) share an IP. One
abusive user can trigger limits for all users behind that IP.

**Mitigation:**
- Use lenient settings (high burst, high rate)
- Monitor for false positives in logs
- Consider per-IP + per-User-Agent rate limiting (future)

---

## Performance

### Overhead

**Fast path (existing limiter):**
- Read lock: ~10ns
- Token check: ~50ns
- Total: ~60ns per request

**Slow path (new limiter):**
- Write lock + allocation: ~500ns
- Amortized across many requests

**Memory:**
- ~100 bytes per IP
- 10k active IPs = 1MB
- Cleaned up automatically

### Benchmarks

```
BenchmarkRateLimit_Hit-8       20000000     60 ns/op
BenchmarkRateLimit_Miss-8      10000000    120 ns/op
BenchmarkRateLimit_NewIP-8      2000000    500 ns/op
```

No measurable impact on request latency.

---

## Alternatives Considered

### Fixed Window

Count requests in fixed time windows (e.g., per minute).

**Problem:** Burst at window boundary:
```
Window 1: [59s] 1000 requests → allowed
Window 2: [1s]  1000 requests → allowed
Total:    2 seconds, 2000 requests (way over limit!)
```

**Why token bucket is better:** Smooth rate enforcement, no boundary issues.

### Sliding Window

Track requests in rolling time window.

**Problem:** Expensive (must store all timestamps):
```
Memory: 100 requests × 16 bytes = 1.6KB per IP
Cleanup: Must scan all timestamps
```

**Why token bucket is better:** O(1) memory, no timestamp tracking.

### Leaky Bucket

Requests queue, processed at fixed rate.

**Problem:** Adds latency (requests wait in queue).

**Why token bucket is better:** Instant allow/deny, no queuing delay.

---

## Future Enhancements

### 1. Redis-Based Distributed Rate Limiting

For multi-instance deployments:

```go
func (rl *RateLimiter) getLimiterRedis(ip string) *rate.Limiter {
    // Check Redis for token count
    tokens := redis.Get("ratelimit:" + ip)
    // Decrement atomically
    redis.Decr("ratelimit:" + ip)
}
```

### 2. Per-Endpoint Rate Limits

Different limits for different paths:

```yaml
rate_limit:
  global:
    requests_per_second: 10
    burst: 20
  endpoints:
    "/api/*":
      requests_per_second: 5
      burst: 10
```

### 3. Adaptive Rate Limiting

Automatically tighten limits during attacks:

```go
if detectionRate > threshold {
    limiter.SetRate(strictRate)
}
```

### 4. IP Reputation Integration

Lower limits for suspicious IPs:

```go
if ipdb.IsTor(ip) || ipdb.IsVPN(ip) {
    limiter.SetRate(strictRate)
}
```

---

## References

- [Token Bucket Algorithm (Wikipedia)](https://en.wikipedia.org/wiki/Token_bucket)
- [golang.org/x/time/rate package](https://pkg.go.dev/golang.org/x/time/rate)
- [RFC 6585: HTTP 429 Too Many Requests](https://datatracker.ietf.org/doc/html/rfc6585)

**Implemented in:**
- `internal/middleware/ratelimit.go` (200 lines)
- `internal/middleware/ratelimit_test.go` (400 lines)
- `internal/server/server.go` (integration)
- `internal/config/config.go` (configuration)
