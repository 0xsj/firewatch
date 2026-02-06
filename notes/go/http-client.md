# HTTP Client in Go

## http.Client

The standard HTTP client for making outbound requests:

```go
client := &http.Client{}
resp, err := client.Do(req)
```

Always create your own `*http.Client` rather than using
`http.DefaultClient`. The default client has no timeout, meaning
a request can hang forever.

```go
// Bad — no timeout
resp, err := http.Post(url, "application/json", body)

// Good — explicit client with control
client := &http.Client{Timeout: 30 * time.Second}
resp, err := client.Do(req)
```

**Used in:** `internal/alerts/slack.go`, `discord.go`, `webhook.go`

---

## Building Requests

`http.NewRequestWithContext` creates a request with a context for
cancellation and deadlines:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
if err != nil { ... }

req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer token")
```

### Why NewRequestWithContext over NewRequest?

`NewRequest` creates a request with `context.Background()` — it
can't be canceled. Always prefer `NewRequestWithContext` so the
caller can control the request lifecycle.

```go
// This request respects the parent context's deadline
ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
resp, err := client.Do(req)
// If ctx expires, client.Do returns immediately with an error
```

**Used in:** `internal/alerts/slack.go`, `discord.go`, `webhook.go`

---

## bytes.NewReader vs bytes.NewBuffer

Both create an `io.Reader` from bytes. Prefer `bytes.NewReader`:

```go
// NewReader — read-only, lightweight
req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))

// NewBuffer — read/write, heavier (has methods for appending)
req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
```

`NewReader` is more efficient for read-only use (like sending a
request body). `NewBuffer` is needed only if you also want to
write/append to the buffer.

**Used in:** `internal/alerts/slack.go`, `discord.go`, `webhook.go`

---

## Response Handling

Always close the response body, even if you don't read it:

```go
resp, err := client.Do(req)
if err != nil {
    return err  // no resp to close
}
defer resp.Body.Close()  // ALWAYS close

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("unexpected status: %d", resp.StatusCode)
}
```

### Why defer resp.Body.Close()?

The response body holds a network connection. If you don't close
it, the connection can't be reused by the connection pool and
eventually you'll exhaust available file descriptors.

### Status code checking

HTTP libraries in other languages throw on non-2xx status codes.
Go does not — a 500 response is a successful HTTP exchange. You
must check `resp.StatusCode` explicitly:

```go
// Slack returns 200 on success
if resp.StatusCode != http.StatusOK { ... }

// Discord returns 2xx range
if resp.StatusCode < 200 || resp.StatusCode >= 300 { ... }
```

**Used in:** `internal/alerts/slack.go`, `discord.go`, `webhook.go`

---

## JSON Encoding for Request Bodies

```go
body, err := json.Marshal(payload)
if err != nil { ... }

req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
```

The pattern: struct → `json.Marshal` → `[]byte` → `bytes.NewReader`
→ request body.

Alternative for streaming large payloads:

```go
pr, pw := io.Pipe()
go func() {
    json.NewEncoder(pw).Encode(payload)
    pw.Close()
}()
req, _ := http.NewRequestWithContext(ctx, "POST", url, pr)
```

For alert payloads (small JSON), `Marshal` + `NewReader` is simpler.

**Used in:** `internal/alerts/slack.go`, `discord.go`, `webhook.go`

---

## map[string]any for Dynamic JSON

When the JSON structure doesn't justify a dedicated struct (like
Slack Block Kit payloads), use `map[string]any`:

```go
payload := map[string]any{
    "blocks": []map[string]any{
        {
            "type": "header",
            "text": map[string]any{
                "type": "plain_text",
                "text": title,
            },
        },
    },
}
```

Trade-offs vs struct:

| Approach          | Pros                     | Cons                       |
|-------------------|--------------------------|----------------------------|
| `map[string]any`  | Flexible, no types needed | No compile-time checking   |
| Dedicated struct  | Type-safe, documented    | Verbose for nested one-offs |

For third-party API payloads that you don't deserialize back,
maps are pragmatic. For your own data models, always use structs.

**Used in:** `internal/alerts/slack.go`, `discord.go`
