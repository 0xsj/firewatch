# Raw String Literals & Manual String Search

## Raw String Literals

Go has two string literal forms:

```go
// Interpreted strings — process escape sequences
s1 := "line1\nline2\ttab"

// Raw strings — backtick delimited, no escapes processed
s2 := `line1\nline2\ttab`  // Contains literal \n and \t characters
```

### When to use raw strings

**Multi-line content** — No need for `\n` concatenation:
```go
w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<methodResponse>
  <params>
    <param>
      <value><string>system.multicall</string></value>
    </param>
  </params>
</methodResponse>`))
```

**Regex patterns** — No double-escaping:
```go
re := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
// vs interpreted: regexp.MustCompile("\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}")
```

**JSON/HTML templates** — Preserves quotes:
```go
json := `{"key": "value"}`  // No escaping quotes
```

### Limitations of raw strings

- Cannot contain a backtick character (no escape mechanism)
- Cannot use `\n` for newlines — actual newlines are included
- Includes all whitespace literally (watch indentation)

## Manual String Search

Instead of importing `strings`, you can implement substring search directly:

```go
func contains(s, substr string) bool {
    return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

### Why manual over `strings.Contains`?

In Firewatch's GraphQL handler, a manual search avoids an import dependency for a simple operation. This is a style choice — `strings.Contains` is perfectly fine and uses the same O(n*m) approach for small strings (Rabin-Karp for larger).

### String slicing and comparison

```go
s[i:i+len(substr)]  // Slice: O(1) — returns view, no copy
s1 == s2             // Compare: O(n) — byte-by-byte
```

Slicing a string does NOT allocate. It creates a new string header pointing to the same underlying bytes. This makes the naive search reasonably efficient for short substrings.

### When to use `strings` package

For anything beyond basic search, use the stdlib:
```go
strings.Contains(s, sub)      // Optimized search
strings.HasPrefix(s, pre)     // O(len(pre))
strings.HasSuffix(s, suf)     // O(len(suf))
strings.ToLower(s)            // Unicode-aware
strings.Split(s, sep)         // Returns []string
strings.ReplaceAll(s, old, new)
```
