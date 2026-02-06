# Regular Expressions in Go

## regexp Package

Go uses the `regexp` package, which implements RE2 syntax — a
guaranteed-linear-time engine. This means no catastrophic backtracking,
but also no backreferences or lookaheads.

```go
import "regexp"

// Compile returns a Regexp or an error
re, err := regexp.Compile(`(?i)^/\.env`)
if err != nil { ... }

// MatchString tests if the string matches
re.MatchString("/.env")        // true
re.MatchString("/.environment") // true
re.MatchString("/api")          // false
```

### Compile vs MustCompile

```go
// Compile — returns error, use for runtime/user-supplied patterns
re, err := regexp.Compile(pattern)

// MustCompile — panics on error, use for hardcoded patterns
re := regexp.MustCompile(`^/\.env`)
```

We use `Compile` in `Matcher.Match()` because signatures can be
loaded from config files at runtime where an invalid regex shouldn't
crash the process.

**Used in:** `internal/detection/signatures.go` — `Matcher.Match()`

---

## Common Patterns Used

### Case-insensitive matching

```go
`(?i)(nuclei|sqlmap|nikto)`
```

`(?i)` sets the case-insensitive flag for the rest of the pattern.
In RE2, flags are set with `(?flags)` syntax, not `/pattern/i`.

### Anchoring

```go
`^/\.env`      // must start with /.env
`\.map$`       // must end with .map
`^/(admin|phpmyadmin)`  // must start with /admin or /phpmyadmin
```

- `^` = start of string
- `$` = end of string
- Without anchors, the pattern matches anywhere in the string

### Escaping special characters

```go
`^/\.env`      // \. matches a literal dot (not "any character")
`\.(sql|bak)$` // literal dot before extension
```

The dot `.` means "any character" in regex. Use `\.` for a literal dot.

In Go raw strings (backtick-quoted), backslashes are literal — no
double-escaping needed:

```go
`\.(sql|bak)$`     // raw string — one backslash
"\\.(sql|bak)$"    // quoted string — must double-escape
```

Always prefer backtick raw strings for regex patterns.

### Word boundary

```go
`^/(backup|dump|database)\b`
```

`\b` matches a word boundary — the point between a word character
(`\w`) and a non-word character. This prevents `/backups` from
matching a pattern meant for `/backup`.

**Used in:** `internal/detection/patterns.go`

---

## Performance Considerations

Compiling a regex is expensive. Matching is fast. For hot paths,
compile once and reuse:

```go
// Bad — recompiles on every call
func match(pattern, value string) bool {
    re, _ := regexp.Compile(pattern)
    return re.MatchString(value)
}

// Good — compile once, match many
var re = regexp.MustCompile(pattern)
func match(value string) bool {
    return re.MatchString(value)
}
```

Our `Matcher.Match()` currently compiles on each call for simplicity.
If detection becomes a bottleneck, the detector should pre-compile
regexes during initialization and cache them.

---

## RE2 vs PCRE Limitations

Go's RE2 engine doesn't support some features common in other
languages (Perl, Python, JavaScript):

| Feature            | RE2 (Go) | PCRE (Perl/Python) |
|--------------------|----------|-------------------|
| Backreferences     | No       | Yes `\1`          |
| Lookahead          | No       | Yes `(?=...)`     |
| Lookbehind         | No       | Yes `(?<=...)`    |
| Possessive quants  | No       | Yes `a++`         |
| Atomic groups      | No       | Yes `(?>...)`     |
| Case-insensitive   | Yes `(?i)` | Yes `/i`        |
| Named groups       | Yes `(?P<name>)` | Yes       |

This trade-off guarantees `O(n)` matching time, which matters for
a honeypot processing potentially malicious input — a PCRE engine
could be DoS'd with crafted regex input (ReDoS).
