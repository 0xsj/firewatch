# Detection Engine Architecture

## Two Layers of Detection

The detection system has two complementary layers:

### Signatures — specific, precise

A signature matches a **single request** against exact conditions.
All matchers in a signature must match (AND logic):

```go
Signature{
    ID:       "wp-bruteforce-001",
    Matchers: []Matcher{
        {Field: FieldMethod, Operator: OpEquals, Value: "POST"},     // AND
        {Field: FieldPath,   Operator: OpEquals, Value: "/wp-login.php"}, // AND
    },
}
```

Both conditions must be true. A GET to `/wp-login.php` doesn't match.
A POST to `/api/login` doesn't match. Only a POST to `/wp-login.php`.

### Patterns — broad, behavioral

A pattern matches against **any of its rules** (OR logic), where
each rule is a set of AND conditions:

```go
Pattern{
    ID: "recon-sweep-001",
    Rules: []Rule{
        {Matchers: []Matcher{{Field: FieldPath, Operator: OpRegex, Value: `^/\.env`}}},        // OR
        {Matchers: []Matcher{{Field: FieldPath, Operator: OpRegex, Value: `^/(backup|dump)`}}}, // OR
        {Matchers: []Matcher{{Field: FieldPath, Operator: OpSuffix, Value: `.bak`}}},           // OR
    },
}
```

Any of these paths triggers the pattern. Patterns cast a wider net.

### Together

```
Request → Detector
           ├── Check all Signatures (AND within each)
           ├── Check all Patterns (OR across rules, AND within each rule)
           └── Combine → DetectionResult
                           ├── matched signatures
                           ├── matched patterns
                           └── highest severity
```

**Used in:** `internal/detection/detector.go`

---

## Field Extraction

The detector flattens a request into a `requestFields` map for
uniform matching:

```go
type requestFields map[MatchField]string

fields := requestFields{
    FieldPath:      "/wp-login.php",
    FieldMethod:    "POST",
    FieldBody:      "log=admin&pwd=password",
    FieldUserAgent: "python-requests/2.28",
    FieldQuery:     "",
    HeaderField("Content-Type"): "application/x-www-form-urlencoded",
    HeaderField("Accept"):       "*/*",
}
```

This abstraction lets matchers work uniformly regardless of where
the data comes from. A matcher doesn't know if it's checking a
path, a header, or the body — it just receives a field name and
a string value.

### Header field encoding

Headers use a prefixed field name:

```go
func HeaderField(name string) MatchField {
    return MatchField("header:" + name)
}

// "header:Next-Action" → targets the Next-Action header
// "header:Content-Type" → targets Content-Type
```

The `IsHeaderField()` and `HeaderName()` methods decode this
prefix for canonical header key fallback.

**Used in:** `internal/detection/detector.go` — `extractFields()`

---

## Severity Ranking

When multiple signatures/patterns match, the result carries the
**highest** severity:

```go
var severityRank = map[string]int{
    "info":     0,
    "low":      1,
    "medium":   2,
    "high":     3,
    "critical": 4,
}

func highestSeverity(a, b string) string {
    if severityRank[b] > severityRank[a] {
        return b
    }
    return a
}
```

A request that matches both a "low" signature and a "critical"
pattern gets severity "critical" in the result.

This ranking is also used for alert filtering — the alert system
can be configured to only fire on `medium` or above.

**Used in:** `internal/detection/detector.go`

---

## Matcher Design

Matchers are the atomic unit of detection. Each one checks a
single field with a single operator:

```go
type Matcher struct {
    Field    MatchField  // what to check
    Operator MatchOp     // how to compare
    Value    string      // expected value
    Negate   bool        // invert result
}
```

### Operators

| Operator   | Behavior                       | Example use                     |
|------------|--------------------------------|---------------------------------|
| `equals`   | Exact string match             | Method = "POST"                 |
| `contains` | Substring present              | Body contains "${jndi:"         |
| `prefix`   | Starts with value              | Path starts with "/_next/"      |
| `suffix`   | Ends with value                | Path ends with ".map"           |
| `regex`    | Regular expression match       | Path matches `^/\.env`          |
| `exists`   | Field is non-empty             | Next-Action header is present   |

### Negation

`Negate: true` inverts the result. This lets you express conditions
like "requests WITHOUT an Accept header":

```go
Matcher{
    Field:    HeaderField("Accept"),
    Operator: OpExists,
    Negate:   true,  // matches when Accept is MISSING
}
```

### Data-driven rules

Signatures and patterns are **data**, not code. This means they can
be loaded from YAML/JSON config files, stored in a database, or
shared between deployments. The `DefaultSignatures()` and
`DefaultPatterns()` functions provide built-in rules, but the
detector accepts any `[]*Signature` and `[]*Pattern`.

**Used in:** `internal/detection/signatures.go`
