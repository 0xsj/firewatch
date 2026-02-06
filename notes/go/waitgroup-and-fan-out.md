# WaitGroup and Fan-Out Concurrency

## The Fan-Out Pattern

Fan-out launches multiple goroutines to do work concurrently, then
waits for all of them to finish:

```
Send()
  ├── go → Slack.Send()
  ├── go → Discord.Send()
  └── go → Webhook.Send()
  WaitGroup.Wait() ← blocks until all three return
```

This is different from:
- **Sequential**: Slack, then Discord, then Webhook (slow)
- **Fire-and-forget**: Launch goroutines, don't wait (risky — errors lost)

Fan-out with WaitGroup gives concurrency AND completion awareness.

---

## sync.WaitGroup Deep Dive

```go
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)              // 1. increment BEFORE launching
    go func(it Item) {
        defer wg.Done()    // 2. decrement when done
        process(it)
    }(item)                // 3. pass loop var as argument
}

wg.Wait()                  // 4. blocks until counter hits 0
```

### Critical ordering: Add before go

```go
// WRONG — race condition
go func() {
    wg.Add(1)      // might run after Wait()
    defer wg.Done()
}()
wg.Wait()

// RIGHT — Add is guaranteed before Wait checks
wg.Add(1)
go func() {
    defer wg.Done()
}()
wg.Wait()
```

If `Add` runs inside the goroutine, there's a window where `Wait`
could see counter=0 and return before the goroutine even starts.

### defer wg.Done() for safety

```go
go func() {
    defer wg.Done()  // runs even if function panics
    riskyOperation()
}()
```

Without `defer`, a panic would leave the WaitGroup counter stuck
and `Wait()` would block forever.

**Used in:** `internal/alerts/manager.go`

---

## Loop Variable Capture

### The classic bug (pre-Go 1.22)

```go
for _, entry := range entries {
    go func() {
        fmt.Println(entry)  // BUG: all goroutines see the LAST entry
    }()
}
```

The closure captures the variable `entry`, not its value. By the
time the goroutines run, the loop has finished and `entry` holds
the last value.

### Fix: pass as argument

```go
for _, entry := range entries {
    go func(e Entry) {
        fmt.Println(e)  // each goroutine gets its own copy
    }(entry)            // value is copied at launch time
}
```

### Go 1.22+ fix

Go 1.22 changed loop variable semantics — each iteration creates
a new variable. The closure bug is fixed at the language level.
But the argument-passing style is still clearest and works everywhere.

**Used in:** `internal/alerts/manager.go` — `Send()`

---

## Fan-Out vs Fire-and-Forget

### Fan-out (what we use)

```go
var wg sync.WaitGroup
for _, a := range alerters {
    wg.Add(1)
    go func(a Alerter) {
        defer wg.Done()
        a.Send(ctx, alert)
    }(a)
}
wg.Wait()  // caller knows when all sends are done
```

- Caller blocks until complete
- Errors can be collected
- Clean shutdown is possible

### Fire-and-forget (avoid)

```go
for _, a := range alerters {
    go a.Send(ctx, alert)  // launch and forget
}
// caller continues immediately — no idea if sends succeeded
```

- Goroutines may outlive the caller
- Errors are silently lost
- On shutdown, in-flight sends may be killed

### Error collection pattern

For collecting errors from fan-out goroutines:

```go
errCh := make(chan error, len(alerters))
var wg sync.WaitGroup

for _, a := range alerters {
    wg.Add(1)
    go func(a Alerter) {
        defer wg.Done()
        if err := a.Send(ctx, alert); err != nil {
            errCh <- err
        }
    }(a)
}

wg.Wait()
close(errCh)

for err := range errCh {
    log.Error("send failed", "error", err)
}
```

Our Manager logs errors inline instead of collecting them, which
is simpler for the alert use case — we don't want to fail the
whole send if one alerter is down.

**Used in:** `internal/alerts/manager.go`
