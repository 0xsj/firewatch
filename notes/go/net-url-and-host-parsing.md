# net/url and Host Parsing

## url.Parse

Parses a raw URL string into components:

```go
u, err := url.Parse("https://example.com:8080/path?q=1")
if err != nil { /* malformed URL */ }

u.Scheme   // "https"
u.Host     // "example.com:8080" (includes port!)
u.Hostname()  // "example.com" (strips port)
u.Port()   // "8080"
u.Path     // "/path"
u.RawQuery // "q=1"
```

### Host vs Hostname()

| Method       | Returns            | Example              |
|--------------|--------------------|----------------------|
| `u.Host`     | host:port (raw)    | `"example.com:8080"` |
| `u.Hostname()` | host only        | `"example.com"`      |
| `u.Port()`   | port only          | `"8080"`             |

`Host` is a raw string field. `Hostname()` and `Port()` are methods that parse it.

## net.SplitHostPort

Splits a `host:port` string manually:

```go
host, port, err := net.SplitHostPort("example.com:8080")
// host="example.com", port="8080"

// IPv6 addresses must be bracketed
host, port, err := net.SplitHostPort("[::1]:8080")
// host="::1", port="8080"

// Fails if no port present
_, _, err := net.SplitHostPort("example.com")
// err: "missing port in address"
```

### Handling optional ports

```go
func stripPort(host string) string {
    h, _, err := net.SplitHostPort(host)
    if err != nil {
        return host // No port present — return as-is
    }
    return h
}
```

This is a common idiom: try to split, fall back to the original string.

## Distinguishing IPs from domains

```go
host := stripPort(u.Host)

if net.ParseIP(host) == nil {
    // It's a domain name (ParseIP returns nil for non-IPs)
    fmt.Println("Domain:", host)
} else {
    // It's an IP address
    fmt.Println("IP:", host)
}
```

`net.ParseIP` returns `nil` for anything that isn't a valid IPv4 or IPv6 address, making it useful as an IP validator/discriminator.

## Used in Firewatch

IOC extraction checks the Referer header for domains:

```go
if referer, ok := event.Headers["referer"]; ok {
    if u, err := url.Parse(referer); err == nil && u.Host != "" {
        host := stripPort(u.Host)
        if net.ParseIP(host) == nil {
            // Extract as domain IOC
        }
        // Also extract full URL as URL IOC
    }
}
```

The double-check (`err == nil && u.Host != ""`) handles both malformed URLs and relative paths (which parse successfully but have no host).
