# encoding/csv and encoding/json

## encoding/csv

Go's CSV package writes RFC 4180 compliant output via a buffered writer.

### Writer pattern

```go
var buf bytes.Buffer       // In-memory buffer implementing io.Writer
w := csv.NewWriter(&buf)   // Wraps any io.Writer

// Write header
w.Write([]string{"id", "type", "value", "severity"})

// Write rows
for _, ioc := range iocs {
    w.Write([]string{ioc.ID, string(ioc.Type), ioc.Value, ioc.Severity})
}

// MUST flush — csv.Writer buffers internally
w.Flush()

// Check for deferred write errors
if err := w.Error(); err != nil {
    return nil, err
}

return buf.Bytes(), nil
```

### Key details

- `Write()` returns an error, but the writer also accumulates errors internally
- `Flush()` writes any buffered data to the underlying writer
- `Error()` returns the first error from any `Write` or `Flush` call
- Fields containing commas, quotes, or newlines are automatically quoted
- `w.Comma = '\t'` switches to TSV format

### bytes.Buffer as io.Writer

`bytes.Buffer` implements `io.Writer`, making it useful for building output in memory:

```go
var buf bytes.Buffer      // Zero value is ready to use
buf.Write([]byte("data")) // Implements io.Writer
buf.WriteString("text")   // Convenience for strings
buf.Bytes()               // Get accumulated bytes (no copy)
buf.String()              // Get as string
```

No need to initialize — the zero value is a valid empty buffer.

## encoding/json

### MarshalIndent for readable output

```go
// Standard — compact JSON
data, err := json.Marshal(obj)
// {"type":"bundle","id":"bundle--abc"}

// Indented — human-readable
data, err := json.MarshalIndent(obj, "", "  ")
// {
//   "type": "bundle",
//   "id": "bundle--abc"
// }
```

Parameters: `MarshalIndent(v, prefix, indent)`
- `prefix`: prepended to each line (usually `""`)
- `indent`: prepended per nesting level (usually `"  "` or `"\t"`)

### Struct tags control JSON output

```go
type Indicator struct {
    Type     string   `json:"type"`           // Rename field
    Labels   []string `json:"labels,omitempty"` // Omit if nil/empty
    internal string   // Unexported — never marshaled
}
```

### any (interface{}) in JSON structs

When a struct field can hold different types, use `any`:

```go
type Bundle struct {
    Objects []any `json:"objects"`  // Can hold Indicator, Campaign, etc.
}

bundle.Objects = append(bundle.Objects, indicator)  // stixIndicator
bundle.Objects = append(bundle.Objects, campaign)   // stixCampaign
```

`json.Marshal` uses reflection to serialize the concrete type stored in the `any` slice. Each element can be a different struct type — the JSON output reflects whatever the actual type is.

### fmt.Sprintf for dynamic IDs

Common pattern for namespaced identifiers:

```go
id := fmt.Sprintf("indicator--%s", uuid)  // "indicator--abc-123-..."
id := fmt.Sprintf("bundle--%s", uuid)     // "bundle--abc-123-..."
```
