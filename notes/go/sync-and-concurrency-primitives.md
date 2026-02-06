# Sync and Concurrency Primitives

## sync.RWMutex

A reader/writer mutual exclusion lock. Multiple readers can hold
the lock simultaneously, but a writer gets exclusive access:

```go
type JA3Store struct {
    mu     sync.RWMutex
    hellos map[string]*TLSClientHello
}

// Write path — exclusive lock
func (s *JA3Store) Put(addr string, hello *TLSClientHello) {
    s.mu.Lock()         // blocks until all readers and writers release
    s.hellos[addr] = hello
    s.mu.Unlock()
}

// Read-and-delete — also needs exclusive lock (it modifies the map)
func (s *JA3Store) Take(addr string) *TLSClientHello {
    s.mu.Lock()
    hello := s.hellos[addr]
    delete(s.hellos, addr)
    s.mu.Unlock()
    return hello
}
```

### When to use RWMutex vs Mutex

| Scenario                         | Use           |
|----------------------------------|---------------|
| Reads heavily outnumber writes   | `sync.RWMutex` |
| Reads and writes are balanced    | `sync.Mutex`   |
| Lock hold time is very short     | `sync.Mutex`   |

`RWMutex` has higher overhead per operation than `Mutex`, so it
only pays off when the read-to-write ratio is high.

For `JA3Store`, each connection does one `Put` (write) and one
`Take` (write), so `sync.Mutex` would actually suffice. We use
`RWMutex` in case we add read-only methods later.

**Used in:** `internal/fingerprint/ja3.go`

---

## Maps Are Not Concurrent-Safe

Go maps panic if read and written simultaneously from different
goroutines:

```go
// WRONG — data race, will eventually panic
m := make(map[string]int)
go func() { m["a"] = 1 }()
go func() { _ = m["a"] }()
```

Solutions:
1. **Mutex-protected map** (what we use)
2. **sync.Map** — optimized for append-only or read-heavy workloads
3. **Channel-based access** — serialize through a goroutine

We use option 1 because the access pattern (put, then take) is
simple and `sync.Map`'s API is less ergonomic for typed values.

**Used in:** `internal/fingerprint/ja3.go` — `JA3Store`

---

## The Put/Take Pattern

Also called "producer/consumer with a map." One goroutine stores
data, another retrieves and removes it:

```go
// Producer (TLS callback goroutine)
store.Put(remoteAddr, data)

// Consumer (HTTP handler goroutine)
data := store.Take(remoteAddr)
```

`Take` deletes the entry after reading — this prevents unbounded
memory growth from completed connections. Without deletion, the
map would grow forever.

Alternative: use a TTL-based cache or `sync.Map` with periodic
cleanup. For our use case, the 1:1 put/take pattern is cleaner.

**Used in:** `internal/fingerprint/ja3.go`

---

## Comma-Ok Pattern on Map Access

```go
hello, ok := s.hellos[remoteAddr]
if ok {
    delete(s.hellos, remoteAddr)
}
```

- `ok` is `true` if the key exists, `false` otherwise
- Without the comma-ok form, accessing a missing key returns the
  zero value silently (nil for pointers, 0 for ints, "" for strings)
- Always use comma-ok when the zero value is a valid result or when
  you need to distinguish "not found" from "found with zero value"

This is the same pattern used with type assertions and channel
receives:

```go
// Map access
val, ok := m[key]

// Type assertion
val, ok := iface.(ConcreteType)

// Channel receive
val, ok := <-ch  // ok is false when channel is closed
```

**Used in:** `internal/fingerprint/ja3.go` — `Take()`
