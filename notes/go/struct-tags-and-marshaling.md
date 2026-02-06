# Struct Tags and Marshaling

## Struct Tags

Struct tags are string metadata attached to fields, read at runtime
via reflection. They control how encoding packages map struct fields
to external formats.

```go
type ServerConfig struct {
    Domain string    `yaml:"domain"`
    Port   int       `yaml:"port"`
    TLS    TLSConfig `yaml:"tls"`
}
```

- The tag is the backtick-quoted string after the type
- Format: `key:"value"` — multiple tags separated by spaces
- A field can have multiple tags: `yaml:"port" json:"port" env:"PORT"`

### Common tag keys

| Key    | Package          | Purpose                    |
|--------|------------------|----------------------------|
| `json` | `encoding/json`  | JSON field names           |
| `yaml` | `gopkg.in/yaml.v3` | YAML field names        |
| `db`   | `sqlx`           | Database column names      |
| `xml`  | `encoding/xml`   | XML element names          |

### Tag options

```go
type Event struct {
    ID    string     `json:"id"`                    // rename to "id"
    GeoIP *GeoIPInfo `json:"geoip,omitempty"`        // omit if nil/zero
    inner string     `json:"-"`                      // skip entirely
}
```

- `omitempty`: omit the field if it's the zero value (0, "", nil, empty slice)
- `-`: never include this field in output
- No tag at all: uses the field name as-is (`ID` → `"ID"`)

**Used in:** `internal/config/config.go` (yaml), `internal/storage/models/` (json)

---

## Unmarshaling into Pre-populated Structs

A powerful pattern: create a struct with defaults, then unmarshal
over it. Only fields present in the input get overwritten.

```go
func Load(path string) (*Config, error) {
    cfg := Default()  // fully populated with defaults

    data, err := os.ReadFile(path)
    // ...

    if err := yaml.Unmarshal(data, cfg); err != nil {
        // ...
    }
    return cfg, nil
}
```

If the YAML file only sets `server.port: 9090`, every other field
keeps its default value. This works because `yaml.Unmarshal` (and
`json.Unmarshal`) only touch fields that appear in the input.

**Used in:** `internal/config/config.go` — `Load()`

---

## json.Marshal / json.Unmarshal

```go
// Struct → JSON bytes
data, err := json.Marshal(event.Headers)
// map[string]string{"Host":"example.com"} → []byte(`{"Host":"example.com"}`)

// JSON bytes → struct
var headers map[string]string
err := json.Unmarshal([]byte(jsonString), &headers)
```

- `Marshal` returns `[]byte` and an error
- `Unmarshal` takes `[]byte` and a **pointer** to the target
- Passing a non-pointer to Unmarshal is a common mistake — it compiles
  but silently does nothing

### Storing complex types in SQLite

SQLite doesn't have array or map column types. We serialize them:

```go
// Save: struct → JSON string → TEXT column
headers, _ := json.Marshal(event.Headers)
db.Exec("INSERT INTO events (headers) VALUES (?)", string(headers))

// Load: TEXT column → JSON string → struct
var raw sql.NullString
row.Scan(&raw)
if raw.Valid {
    json.Unmarshal([]byte(raw.String), &event.Headers)
}
```

**Used in:** `internal/storage/sqlite.go` — all Save/scan functions

---

## os.ReadFile and File Existence

```go
// Read entire file into memory
data, err := os.ReadFile(path)

// Check if file exists before reading
if _, err := os.Stat(path); os.IsNotExist(err) {
    // file doesn't exist
}
```

- `os.ReadFile` returns the full contents — fine for config files,
  not for large files (use `os.Open` + buffered reading instead)
- `os.Stat` returns file info or an error
- `os.IsNotExist(err)` checks if the error means "file not found"
  vs some other failure (permissions, etc.)

**Used in:** `internal/config/config.go` — `Load()`, `LoadOrDefault()`

---

## Pointer Types for Optional Fields

```go
type Event struct {
    GeoIP *GeoIPInfo `json:"geoip,omitempty"`
}
```

Using a pointer (`*GeoIPInfo` vs `GeoIPInfo`) means:
- `nil` = "no GeoIP data" — distinct from "empty GeoIP data"
- `omitempty` works correctly — nil pointer is omitted from JSON
- A non-pointer struct is never nil, so `omitempty` won't omit it
  even when all its fields are zero

Rule of thumb: use pointers for optional nested structs where
"not present" is meaningful.

**Used in:** `internal/storage/models/event.go`, `attacker.go`, `ioc.go`
