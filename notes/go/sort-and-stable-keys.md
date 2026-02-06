# sort Package & Stable Key Generation

## sort.Strings

Sorts a `[]string` in-place (ascending, lexicographic):

```go
names := []string{"cloud", "api", "wordpress"}
sort.Strings(names)
// names is now ["api", "cloud", "wordpress"]
```

Modifies the original slice — there is no return value.

## Why sort for keys?

Maps in Go have non-deterministic iteration order. When you need a stable key from a set of values (e.g., for deduplication or grouping), sort first:

```go
// Same set in different orders must produce the same key
sigs1 := []string{"wp-login-001", "wp-bruteforce-001"}
sigs2 := []string{"wp-bruteforce-001", "wp-login-001"}

func signatureKey(sigs []string) string {
    sorted := make([]string, len(sigs))
    copy(sorted, sigs)       // Don't mutate the original
    sort.Strings(sorted)
    return strings.Join(sorted, "+")
}

signatureKey(sigs1) // "wp-bruteforce-001+wp-login-001"
signatureKey(sigs2) // "wp-bruteforce-001+wp-login-001" — same!
```

### copy before sort

`sort.Strings` modifies in-place. If the original slice shouldn't be mutated:

```go
sorted := make([]string, len(original))
copy(sorted, original)
sort.Strings(sorted)
```

`copy(dst, src)` copies `min(len(dst), len(src))` elements. The destination must be pre-allocated.

## Set to sorted slice

Converting `map[string]struct{}` to a deterministic slice:

```go
func setToSlice(s map[string]struct{}) []string {
    result := make([]string, 0, len(s))
    for k := range s {
        result = append(result, k)
    }
    sort.Strings(result)
    return result
}
```

This pattern appears whenever you need to:
- Generate a stable composite key from a set
- Produce deterministic output from map data
- Compare sets by converting to sorted slices

## Composite keys for deduplication

```go
// Dedup IOCs by type+value
seen := make(map[string]*models.IOC)
key := string(ioc.Type) + ":" + ioc.Value  // "ip:192.168.1.1"

if existing, ok := seen[key]; ok {
    // Merge into existing
} else {
    seen[key] = ioc
}
```

The key format `type:value` ensures uniqueness — two IOCs with the same value but different types (unlikely but possible) remain distinct.
