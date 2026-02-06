# Clustering & Deduplication Patterns

## Map-Based Deduplication

The core idiom: use a map keyed by a composite string to merge duplicates.

```go
seen := make(map[string]*models.IOC)

for _, event := range events {
    for _, ioc := range extractFromEvent(event) {
        key := string(ioc.Type) + ":" + ioc.Value

        if existing, ok := seen[key]; ok {
            // Merge: update timestamps, promote severity, merge tags
            existing.LastSeen = ioc.LastSeen
            if severityRank[ioc.Severity] > severityRank[existing.Severity] {
                existing.Severity = ioc.Severity
            }
            existing.Tags = mergeTags(existing.Tags, ioc.Tags)
        } else {
            seen[key] = ioc
        }
    }
}
```

### Merge semantics

When merging duplicates, decide per-field:
- **Timestamps**: Keep earliest `first_seen`, latest `last_seen`
- **Severity**: Keep the highest (promote)
- **Tags**: Union (deduplicated)
- **ID**: Keep the first one's ID

## Set Operations with Maps

Go has no built-in set type. Use `map[string]struct{}`:

```go
// Create a set
ips := make(map[string]struct{})
ips[event.SourceIP] = struct{}{}

// Check membership
if _, ok := ips[ip]; ok { /* exists */ }

// Size
count := len(ips)

// Convert to slice
func setToSlice(s map[string]struct{}) []string {
    result := make([]string, 0, len(s))
    for k := range s {
        result = append(result, k)
    }
    sort.Strings(result)  // Deterministic order
    return result
}
```

`struct{}` takes zero bytes — more memory efficient than `map[string]bool`.

## Tag Deduplication

```go
func mergeTags(a, b []string) []string {
    seen := make(map[string]struct{}, len(a)+len(b))
    for _, t := range a { seen[t] = struct{}{} }
    for _, t := range b { seen[t] = struct{}{} }

    result := make([]string, 0, len(seen))
    for t := range seen {
        result = append(result, t)
    }
    return result
}
```

Pre-size the map with `len(a)+len(b)` to avoid rehashing.

## Clustering by Shared Attributes

Campaign detection groups events by shared characteristics:

### Strategy 1: Signature clustering

```go
// Group events by their sorted signature set
clusters := make(map[string]*cluster)

for _, event := range events {
    key := signatureKey(event.Signatures)  // Sort + join
    c, ok := clusters[key]
    if !ok {
        c = &cluster{ips: make(map[string]struct{})}
        clusters[key] = c
    }
    c.events = append(c.events, event)
    c.ips[event.SourceIP] = struct{}{}
}

// Filter: only clusters with multiple distinct IPs are campaigns
for _, c := range clusters {
    if len(c.ips) >= 2 {
        // This is a campaign — multiple IPs share the same behavior
    }
}
```

### Strategy 2: Module-set coordination

```go
// Group IPs by the set of modules they target
ipModules := make(map[string]map[string]struct{})
for _, event := range events {
    if ipModules[event.SourceIP] == nil {
        ipModules[event.SourceIP] = make(map[string]struct{})
    }
    ipModules[event.SourceIP][event.Module] = struct{}{}
}

// Group IPs that target the same module combination
groups := make(map[string]*coordGroup)
for ip, modules := range ipModules {
    key := moduleSetKey(modules)  // Sort + join
    groups[key].ips = append(groups[key].ips, ip)
}
```

### Threshold filtering

Both strategies use a minimum threshold (`len(ips) >= 2`) to avoid false positives. A single IP hitting multiple endpoints is normal reconnaissance; multiple IPs with identical behavior suggests automation or coordination.

## Nested Maps

Campaign detection uses `map[string]map[string]struct{}` — a map of sets:

```go
ipModules := make(map[string]map[string]struct{})

// Must initialize inner map before use
if ipModules[ip] == nil {
    ipModules[ip] = make(map[string]struct{})
}
ipModules[ip][module] = struct{}{}
```

Writing to a nil inner map panics. Always check and initialize.
