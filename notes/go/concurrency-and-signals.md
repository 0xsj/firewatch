# Concurrency and Signal Handling

## Goroutines

A goroutine is a lightweight concurrent function. Launch with `go`:

```go
go func() {
    errCh <- s.Start()
}()
```

- Goroutines are multiplexed onto OS threads by the Go runtime
- They start with ~8KB of stack (grows as needed)
- No return value — communicate results through channels

**Used in:** `internal/server/graceful.go` — start server in background

---

## Channels

Channels are typed conduits for communication between goroutines:

```go
// Unbuffered — sender blocks until receiver is ready
ch := make(chan error)

// Buffered — sender can put up to N values without blocking
errCh := make(chan error, 1)
```

### Buffered vs unbuffered

```go
// Buffered channel of size 1:
errCh := make(chan error, 1)
go func() {
    errCh <- s.Start()  // never blocks (buffer has room)
}()
```

If we used an unbuffered channel and nobody reads from it (e.g., we
received a signal first), the goroutine would block forever. The
buffer of 1 lets the goroutine exit cleanly.

### Signal channels

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

sig := <-quit  // blocks until signal arrives
```

- `os.Interrupt` = Ctrl+C (SIGINT)
- `syscall.SIGTERM` = `kill <pid>` or container shutdown
- Buffer of 1 ensures the signal isn't dropped if we're not
  waiting on the channel when it arrives

**Used in:** `internal/server/graceful.go`

---

## Select

`select` multiplexes on multiple channel operations. It blocks until
one of them is ready:

```go
select {
case err := <-errCh:
    // Server failed to start
    return err
case sig := <-quit:
    // Shutdown signal received
    logger.Info("signal", "sig", sig)
}
```

- Only one case executes
- If multiple are ready simultaneously, one is chosen at random
- `default` makes it non-blocking (we don't use this here)

### Pattern: start-and-wait

```go
func (s *Server) ListenAndShutdown() error {
    errCh := make(chan error, 1)
    go func() {
        errCh <- s.Start()  // blocks in goroutine
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    select {
    case err := <-errCh:   // start failed
        return err
    case <-quit:           // signal received
    }

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    return s.Shutdown(ctx)
}
```

This is the standard Go server lifecycle:
1. Start server in a goroutine
2. Block on signal or startup error
3. On signal: graceful shutdown with a deadline

**Used in:** `internal/server/graceful.go` — `ListenAndShutdown()`

---

## context.WithTimeout for Shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
s.http.Shutdown(ctx)
```

`http.Server.Shutdown(ctx)`:
- Stops accepting new connections immediately
- Waits for active requests to complete
- If the context deadline expires, forcefully closes remaining connections
- Returns `nil` on clean shutdown, `context.DeadlineExceeded` on timeout

The 30-second timeout prevents a hung connection from blocking
shutdown indefinitely.

**Used in:** `internal/server/graceful.go`
