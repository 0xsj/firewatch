# IO, Net, and Context

## io.LimitReader

Wraps a reader to stop after N bytes — prevents unbounded memory use
when reading untrusted input:

```go
io.ReadAll(io.LimitReader(r.Body, maxSize))
```

Without the limit, a malicious client could send gigabytes and crash
the server. Always limit reads from network sources.

Returns `io.EOF` when the limit is hit, not an error — so `io.ReadAll`
returns the truncated data without error. If you need to detect
truncation, check `len(result) >= maxSize`.

**Used in:** `pkg/httputil/request.go`

---

## net.ParseIP

Parses an IPv4 or IPv6 string into a `net.IP` (which is `[]byte`):

```go
ip := net.ParseIP("192.168.1.1")     // returns 16-byte representation
ip := net.ParseIP("::ffff:192.168.1.1") // IPv4-mapped IPv6, also 16 bytes
ip := net.ParseIP("not-an-ip")       // returns nil
```

### IPv4 normalization

Go internally stores all IPs as 16 bytes (IPv6 form). IPv4 addresses
are stored as IPv4-mapped IPv6 (`::ffff:x.x.x.x`).

```go
ip := net.ParseIP("192.168.1.1")
ip.To4()  // returns 4-byte form, or nil if it's a true IPv6
ip.To4().String()  // "192.168.1.1"
```

This is why `NormalizeIP` checks `To4()` first — to convert
`::ffff:192.168.1.1` back to plain `192.168.1.1`.

**Used in:** `pkg/netutil/ip.go`

---

## net.ParseCIDR

Parses CIDR notation and returns the network:

```go
ip, network, err := net.ParseCIDR("192.168.1.0/24")
// ip      = 192.168.1.0 (the specific IP in the notation)
// network = 192.168.1.0/24 (the *net.IPNet)
// network.Contains(net.ParseIP("192.168.1.50")) → true
```

Note: `ip` is the IP from the string, `network` is the masked network.
For `10.0.0.5/8`, `ip` = `10.0.0.5` but `network` = `10.0.0.0/8`.

**Used in:** `pkg/netutil/ip.go`

---

## net.SplitHostPort

Splits `host:port` from `RemoteAddr`:

```go
host, port, err := net.SplitHostPort("192.168.1.1:8080")
// host = "192.168.1.1", port = "8080"

host, port, err := net.SplitHostPort("[::1]:8080")
// host = "::1", port = "8080" (brackets stripped for IPv6)
```

Returns an error if the format is wrong (no port, missing brackets for
IPv6). That's why `ClientIP` falls back to `r.RemoteAddr` on error.

**Used in:** `pkg/httputil/request.go`

---

## context.WithTimeout

Creates a context that cancels itself after a duration:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()  // always defer cancel to release resources

names, err := resolver.LookupAddr(ctx, ip)
```

Key rules:
- **Always `defer cancel()`** — even if the operation completes before
  the timeout, cancel releases the timer.
- Pass `ctx` to any function that accepts it — this propagates the
  deadline down the call chain.
- Check `ctx.Err()` to distinguish timeout (`DeadlineExceeded`) from
  explicit cancel (`Canceled`).

**Used in:** `pkg/netutil/dns.go`

---

## net.Resolver

`net.DefaultResolver` is the package-level DNS resolver:

```go
resolver := net.DefaultResolver
names, err := resolver.LookupAddr(ctx, "8.8.8.8")
// names = ["dns.google."]
```

`LookupAddr` does a reverse DNS lookup (PTR record). The returned
hostnames have a trailing dot (FQDN format) — we strip it:

```go
if name[len(name)-1] == '.' {
    name = name[:len(name)-1]
}
```

**Used in:** `pkg/netutil/dns.go`

---

## IP Classification Methods

`net.IP` has built-in classifiers:

```go
ip.IsPrivate()          // RFC 1918 (10.x, 172.16-31.x, 192.168.x)
ip.IsLoopback()         // 127.0.0.0/8, ::1
ip.IsLinkLocalUnicast() // 169.254.x.x, fe80::/10
ip.IsGlobalUnicast()    // everything else routable
ip.IsMulticast()        // 224.0.0.0/4, ff00::/8
```

We combine `IsPrivate`, `IsLoopback`, and `IsLinkLocalUnicast` in
`netutil.IsPrivate()` to cover all non-routable addresses. Useful for
filtering out internal traffic from honeypot analysis.

**Used in:** `pkg/netutil/ip.go`
