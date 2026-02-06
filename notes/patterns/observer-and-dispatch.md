# Observer Pattern and Alert Dispatch

## Observer Pattern

The observer pattern decouples event producers from event consumers.
Producers don't know (or care) who's listening — they just publish.
Consumers register interest and receive notifications.

In Firewatch, the detection engine produces alerts and the Manager
dispatches them to registered alerters:

```
Detection Engine → Alert → Manager
                              ├── Slack Alerter
                              ├── Discord Alerter
                              └── Webhook Alerter
```

Adding a new alerter (email, PagerDuty, etc.) doesn't touch the
detection code — just register another observer.

**Used in:** `internal/alerts/manager.go`

---

## Interface-Based Dispatch

Each alerter satisfies a common interface:

```go
type Alerter interface {
    Name() string
    Send(ctx context.Context, alert Alert) error
}
```

The manager holds a slice of alerters, not concrete types:

```go
type Manager struct {
    alerters []alerterEntry
}

type alerterEntry struct {
    alerter     Alerter
    minSeverity string
}
```

### Why a wrapper struct?

Each alerter has its own severity threshold. We can't add a
`MinSeverity` field to the interface (that's config, not behavior).
The `alerterEntry` struct pairs the alerter with its config:

```go
m.Register(slack, "medium")   // slack gets medium+ alerts
m.Register(webhook, "info")   // webhook gets everything
```

This is the **decorator** concept — wrapping an object with
additional metadata without changing its interface.

**Used in:** `internal/alerts/manager.go` — `Register()`

---

## Concurrent Fan-Out

The manager sends to all alerters concurrently:

```go
func (m *Manager) Send(ctx context.Context, alert Alert) {
    var wg sync.WaitGroup

    for _, entry := range m.alerters {
        if !MeetsSeverity(alert.Severity, entry.minSeverity) {
            continue  // skip below-threshold alerters
        }

        wg.Add(1)
        go func(e alerterEntry) {
            defer wg.Done()

            sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
            defer cancel()

            e.alerter.Send(sendCtx, alert)
        }(entry)
    }

    wg.Wait()
}
```

### sync.WaitGroup

A counter that blocks until all goroutines finish:

```go
var wg sync.WaitGroup

wg.Add(1)       // increment counter
go func() {
    defer wg.Done()  // decrement when goroutine finishes
    // ... work ...
}()

wg.Wait()  // blocks until counter reaches 0
```

Rules:
- Call `Add(n)` before launching the goroutine, not inside it
- Always `defer wg.Done()` to handle panics/early returns
- `Wait()` blocks the caller until all goroutines complete

### Loop variable capture

```go
go func(e alerterEntry) {
    e.alerter.Send(...)
}(entry)  // pass `entry` as argument
```

The `entry` variable is passed as a function argument. This
creates a copy for each goroutine. Without this, all goroutines
would share the same loop variable and race on it.

In Go 1.22+ the loop variable is per-iteration (so the closure
capture is safe), but passing as an argument is still the clearest
pattern and works in all Go versions.

### Per-alerter timeout

Each alerter gets its own 10-second deadline:

```go
sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
```

A slow Slack webhook doesn't delay the Discord send. If Slack
hangs, its context expires after 10 seconds while Discord may
have already returned.

**Used in:** `internal/alerts/manager.go` — `Send()`

---

## Severity as a Gate

Alerts flow through a severity check before dispatch:

```go
func MeetsSeverity(alertSeverity, minSeverity string) bool {
    return severityRank[alertSeverity] >= severityRank[minSeverity]
}
```

This lets operators tune noise per channel:
- Slack: `min_severity: high` — only critical/high alerts
- Webhook/SIEM: `min_severity: info` — everything for analysis
- Discord: `min_severity: medium` — moderate and above

The same `severityRank` map appears in both the detection engine
and the alert system. They share the same scale but serve different
purposes: detection assigns severity, alerting filters on it.

**Used in:** `internal/alerts/alerter.go`
