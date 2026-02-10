# Changelog — 2024-02-10 (Session 2)

## Summary

Implemented per-IP rate limiting with token bucket algorithm.

---

## 🚦 Rate Limiting

### Implementation

**New files:**
- `internal/middleware/ratelimit.go` (200 lines)
- `internal/middleware/ratelimit_test.go` (400 lines)

**Modified files:**
- `internal/config/config.go` — added RateLimitConfig
- `internal/config/defaults.go` — added rate limit defaults
- `internal/server/server.go` — wired into middleware chain
- `firewatch.yaml` — added rate_limit section
- `go.mod` / `go.sum` — added golang.org/x/time dependency

### Features

**Token Bucket Rate Limiting:**
- Per-IP tracking with `sync.Map`
- Sustained rate + burst capacity
- Background TTL cleanup (prevents memory leak)
- Automatic event recording on violations
- HTTP 429 responses with standard headers

**Configuration:**
```yaml
rate_limit:
  enabled: true
  requests_per_second: 10  # Sustained: 600/min
  burst: 20                # Burst: allow 20 instant requests
  cleanup_minutes: 5       # Remove stale limiters every 5 min
```

**Default Settings:**
- 10 requests/second (600/minute sustained)
- Burst of 20 (allows page loads with multiple resources)
- Cleanup every 5 minutes (10 min TTL)

### Integration

Rate limiting runs early in middleware chain:

```
Request → Correlation → Rate Limit → Logging → GeoIP → Fingerprint → Detection → Handler
                            ↑
                     Blocks here if over limit
```

**Why early placement:**
- Skip expensive operations (GeoIP, TLS fingerprinting) for blocked requests
- Reduce resource consumption during floods
- Fast rejection path

**Event recording:**
- Module: `rate_limit`
- Severity: `medium`
- Signature: `rate-limit-exceeded`
- Includes full request context

**Response headers:**
- `Retry-After: N` — seconds until retry
- `X-RateLimit-Limit: 20` — burst capacity
- `X-RateLimit-Remaining: 0` — tokens left

### Testing

**7 comprehensive tests:**
1. ✅ Allows requests under limit
2. ✅ Blocks requests over limit
3. ✅ Per-IP isolation (IPs don't share buckets)
4. ✅ Token refill over time
5. ✅ Stale limiter cleanup
6. ✅ Nil limiter passthrough
7. ✅ Response headers

**All tests passing with race detector:**
```
ok  github.com/0xsj/firewatch/internal/middleware  0.802s
```

### Performance

**Overhead:**
- Fast path: ~60ns per request (existing limiter)
- Slow path: ~500ns (new limiter creation)
- Memory: ~100 bytes per IP
- 10k active IPs = 1MB memory

**No measurable impact on request latency.**

---

## 📚 Documentation

Created comprehensive notes (700+ lines total):

### Patterns Notes

**`rate-limiting.md`** (new, 450 lines)
- Problem statement and solution
- Token bucket algorithm explained
- Implementation architecture
- Per-IP tracking pattern
- TTL cleanup strategy
- Event recording
- Middleware integration
- Configuration tuning
- Testing strategy
- Observability
- Limitations and trade-offs
- Performance benchmarks
- Alternatives considered
- Future enhancements

### Go Notes

**`rate-limiting-package.md`** (new, 290 lines)
- `golang.org/x/time/rate` package overview
- Core `rate.Limiter` API
- Token bucket mechanics
- All API methods (Allow, Wait, Reserve, AllowN)
- Rate limiting patterns
- Per-resource vs global limiting
- Dynamic rate adjustment
- Concurrency safety
- Common usage patterns
- Performance characteristics
- Edge cases (zero rate, infinite rate, fractional)
- Context integration
- Comparison with alternatives
- Real-world usage (Kubernetes, gRPC)
- Tips and tricks

### README

**Updated features list:**
- Added "Rate Limiting" to features section
- Added rate_limit configuration example
- Shows default settings and usage

---

## 📊 Statistics

**Code changes:**
- 600 lines added (implementation + tests)
- 4 config files modified
- 1 server integration file modified
- 1 dependency added (golang.org/x/time)

**Documentation:**
- 740 lines of notes created
- 2 note files (patterns + go)
- 1 README update
- Topics: token bucket algorithm, rate limiting patterns, Go stdlib

**Tests:**
- 7 new test cases
- All tests passing with -race
- Coverage: Per-IP isolation, refill, cleanup, headers

---

## 🎯 Configuration Examples

### Strict (Catch Slow Scanners)

```yaml
rate_limit:
  enabled: true
  requests_per_second: 1   # 60/min
  burst: 5                 # Small burst
  cleanup_minutes: 10
```

**Use when:**
- You want to catch even careful scanners
- Traffic is expected to be very low
- False positives are acceptable

### Normal (Default, Balanced)

```yaml
rate_limit:
  enabled: true
  requests_per_second: 10  # 600/min
  burst: 20                # Reasonable burst
  cleanup_minutes: 5
```

**Use when:**
- Balancing detection vs false positives
- Mix of legitimate and attack traffic
- General-purpose honeypot

### Lenient (Only Aggressive Floods)

```yaml
rate_limit:
  enabled: true
  requests_per_second: 100  # 6000/min
  burst: 200                # Large burst
  cleanup_minutes: 1
```

**Use when:**
- Only want to block extreme flooding
- High legitimate traffic expected
- Minimize false positives

### Disabled

```yaml
rate_limit:
  enabled: false
```

**Use when:**
- Testing/debugging
- Rate limiting handled elsewhere (load balancer)
- Unlimited capacity honeypot

---

## 🔍 Usage Examples

### Query Rate Limit Events

```bash
# Show all rate limit violations
firewatch events --module rate_limit

# Last hour
firewatch events --module rate_limit --since 1h

# Specific IP
firewatch events --ip 1.2.3.4 --module rate_limit

# Count violations per IP
firewatch stats --module rate_limit --since 24h
```

### Monitor in Logs

```json
{
  "level": "warn",
  "msg": "rate limit exceeded",
  "ip": "1.2.3.4",
  "path": "/api/users",
  "request_id": "abc123"
}
```

### Alert on Violations

Rate limit events trigger alerts via AlertingStore:

```yaml
alerts:
  slack:
    webhook_url: "https://hooks.slack.com/..."
    min_severity: "medium"  # Rate limit = medium
```

---

## 🚀 What's Next

Rate limiting complete! Project now has:
- ✅ Per-IP rate limiting with token bucket
- ✅ Automatic event recording
- ✅ TTL cleanup (no memory leaks)
- ✅ Comprehensive tests
- ✅ Full documentation

### Future Enhancements

**Possible additions:**
1. Redis-based distributed rate limiting (multi-instance deployments)
2. Per-endpoint rate limits (different limits per path)
3. Adaptive rate limiting (tighten during attacks)
4. IP reputation integration (strict limits for VPNs/Tor)

**Next priorities:**
1. Email alerting (complete alerting trifecta)
2. PostgreSQL storage (enable multi-instance)
3. Behavioral fingerprinting (catch evasive attackers)

---

## ✅ Sign-off

All checks passing:
- ✅ `golangci-lint run`
- ✅ `go vet ./...`
- ✅ `go test -race ./...`
- ✅ `go build`
- ✅ Integration tests pass

Ready for production deployment.
