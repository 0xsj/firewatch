# Maps as Custom Types

## Type Aliases for Maps

You can define a named type over a map for clarity and to attach
methods:

```go
type requestFields map[MatchField]string
```

This is the same underlying type as `map[MatchField]string`, but
the name documents intent. You can create values with either syntax:

```go
// Both work:
fields := requestFields{FieldPath: "/api"}
fields := make(requestFields)
```

### Why not use a struct?

For `requestFields`, the key set is dynamic — we don't know at
compile time which headers will be present. A struct requires
fixed fields. A map handles arbitrary keys naturally.

Compare:

```go
// Struct — fixed fields, can't handle arbitrary headers
type RequestData struct {
    Path      string
    Method    string
    UserAgent string
    // How do you add Header-X, Header-Y, ...?
}

// Map — dynamic keys, handles any header
type requestFields map[MatchField]string
fields[HeaderField("X-Custom")] = "value"  // works for any header
```

**Used in:** `internal/detection/detector.go`

---

## Maps as Lookup Tables

Maps with constant values serve as efficient lookup tables:

```go
var severityRank = map[string]int{
    "info":     0,
    "low":      1,
    "medium":   2,
    "high":     3,
    "critical": 4,
}
```

This is an alternative to the array-indexed approach used in
`kinds.go`. Trade-offs:

| Approach           | When to use                      |
|--------------------|---------------------------------|
| Array indexed      | Sequential iota-based keys       |
| Map lookup         | String keys or sparse values     |
| Switch statement   | Different return types per case  |

For severity, the keys are strings from user config and JSON
data, so a map is the natural choice. For `Kind` (uint8 with
iota values 0-10), array indexing is faster.

### Zero value on missing key

```go
rank := severityRank["unknown"]  // returns 0 (int zero value)
rank := severityRank["critical"] // returns 4
```

In `highestSeverity`, an unrecognized severity gets rank 0, which
means any valid severity will be considered higher. This is the
right default — unknown severity should never override a known one.

**Used in:** `internal/detection/detector.go`

---

## Struct Slices as Tables

The known client list in `headers.go` uses a slice of anonymous
structs as a simple lookup table:

```go
var knownClients = []struct {
    pattern string
    name    string
}{
    {"python-requests", "python-requests"},
    {"Nuclei", "nuclei"},
    {"sqlmap", "sqlmap"},
}
```

This is lighter than defining a named struct type for something
used only in one place. The anonymous struct is declared inline
with the variable.

### When to use which

| Structure              | When to use                          |
|------------------------|--------------------------------------|
| `map[string]string`    | Key-based lookup, O(1)               |
| `[]struct{k,v}`        | Sequential scan, small dataset       |
| Named struct + slice   | Complex records, used across packages |
| Array with iota index  | Enum-like mapping, O(1)              |

The known clients list uses sequential scan because:
- The dataset is small (~20 entries)
- We need substring matching (`strings.Contains`), not exact lookup
- A map can't do substring matching

**Used in:** `internal/fingerprint/headers.go`, `internal/detection/detector.go`
