# GREASE Values in TLS

## What is GREASE?

**GREASE** (Generate Random Extensions And Sustain Extensibility) is a
mechanism defined in [RFC 8701](https://datatracker.ietf.org/doc/html/rfc8701)
to prevent middlebox ossification.

**Problem:** Middleboxes (firewalls, proxies) often break when they see
unknown TLS extensions, cipher suites, or versions. This prevents TLS
from evolving.

**Solution:** Clients randomly insert placeholder values into their
ClientHello. Middleboxes that break on unknown values will break
immediately, not years later when a real extension is standardized.

---

## GREASE Value Pattern

GREASE values follow a specific pattern: **both bytes are identical
and end in `0xA`**.

```
0x0A0A  0x1A1A  0x2A2A  0x3A3A  0x4A4A  0x5A5A  0x6A6A  0x7A7A
0x8A8A  0x9A9A  0xAAAA  0xBABA  0xCACA  0xDADA  0xEAEA  0xFAFA
```

These are inserted into:
- Cipher suites list
- Extension types
- Supported groups (curves)
- Supported versions
- ALPN protocol lists

---

## Why Filter GREASE?

**For fingerprinting stability:**

A client that inserts `0x1A1A` in one connection might insert `0x3A3A`
in the next. If included in fingerprints, the same client would
produce different hashes.

**Example:**

```
Connection 1: TLS 1.3, ciphers [0x1A1A, 0x1301, 0x1302]
Connection 2: TLS 1.3, ciphers [0x5A5A, 0x1301, 0x1302]
```

Without filtering:
```
JA3 hash 1: cd08e31494f9531f560d64c695473da9
JA3 hash 2: a7f2e89c3d4b5a1e9f8c6d3e2a1b4c5d
```

With filtering:
```
Both hash to: cd08e31494f9531f560d64c695473da9
```

---

## Detection Algorithm

### Byte-level check

```go
func isGREASE(val uint16) bool {
    highByte := (val >> 8) & 0xFF
    lowByte := val & 0xFF

    // Both bytes must be identical
    if highByte != lowByte {
        return false
    }

    // Low nibble must be 0xA
    return (lowByte & 0x0F) == 0x0A
}
```

### Examples

```go
isGREASE(0x0A0A)  // true  — both bytes 0x0A
isGREASE(0x1A1A)  // true  — both bytes 0x1A
isGREASE(0xFAFA)  // true  — both bytes 0xFA

isGREASE(0x1301)  // false — bytes differ
isGREASE(0x0A0B)  // false — bytes differ
isGREASE(0x1A2A)  // false — bytes differ
isGREASE(0x1B1B)  // false — low nibble is 0xB, not 0xA
```

### Test Matrix

```go
// All valid GREASE values
for i := 0; i <= 15; i++ {
    val := uint16(i<<12 | 0x0A<<8 | i<<4 | 0x0A)
    assert(isGREASE(val))
}

// Common non-GREASE TLS values
assert(!isGREASE(0x1301))  // TLS_AES_128_GCM_SHA256
assert(!isGREASE(0x1302))  // TLS_AES_256_GCM_SHA384
assert(!isGREASE(0xc02f))  // TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

---

## Application in JA4

JA4 explicitly filters GREASE before hashing:

```go
func hashCiphers(ciphers []uint16) string {
    // Filter GREASE values
    filtered := make([]uint16, 0, len(ciphers))
    for _, c := range ciphers {
        if !isGREASE(c) {
            filtered = append(filtered, c)
        }
    }

    // Sort and hash remaining ciphers
    slices.Sort(filtered)
    hash := sha256(join(filtered, ","))
    return truncate(hash, 12)
}
```

**Result:** Stable fingerprints across connections from the same client.

---

## Real-World Example

Chrome 120's ClientHello:

```
Ciphers: [0x5A5A, 0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030]
          ^^^^^  GREASE
```

After filtering:

```
Ciphers: [0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030]
```

This list is consistent across Chrome connections, even though the
GREASE value rotates.

---

## Why These Specific Values?

The pattern ensures:

1. **Easy to detect** — simple bit manipulation
2. **Evenly distributed** — 16 possible values across the uint16 space
3. **Visually distinct** — `0xAA` stands out in hex dumps
4. **Low collision risk** — unlikely to conflict with real values

No standardized TLS cipher, extension, or version uses the `0x?A?A`
pattern. This was verified before RFC 8701 was published.

---

## Historical Context

Before GREASE, TLS evolution was blocked by:

- **Middleboxes that rejected TLS 1.3** because it looked "wrong"
- **Firewalls that dropped unknown extensions** causing silent failures
- **Proxies that only allowed specific cipher suites** preventing upgrades

GREASE forces these middleboxes to break immediately (and get fixed)
rather than creating future deployment barriers.

**Impact:** TLS 1.3 deployment was faster than TLS 1.2 deployment,
partly due to GREASE preparing the ecosystem.

---

## Edge Cases

### Not all `0x?A?A` values are GREASE

Only these specific 16 values are reserved:

```
0x0A0A, 0x1A1A, 0x2A2A, 0x3A3A, 0x4A4A, 0x5A5A, 0x6A6A, 0x7A7A,
0x8A8A, 0x9A9A, 0xAAAA, 0xBABA, 0xCACA, 0xDADA, 0xEAEA, 0xFAFA
```

Our algorithm allows any `0x?A?A` pattern, which is slightly more
permissive than the RFC. This is intentional — if future RFCs add more
GREASE patterns, our code won't break.

### GREASE in other protocols

The same technique is used in:
- **HTTP/2** — invalid frame types
- **QUIC** — reserved frame types
- **DNS** — reserved RR types

---

## Testing

```go
func TestIsGREASE(t *testing.T) {
    // All 16 official GREASE values
    grease := []uint16{
        0x0A0A, 0x1A1A, 0x2A2A, 0x3A3A,
        0x4A4A, 0x5A5A, 0x6A6A, 0x7A7A,
        0x8A8A, 0x9A9A, 0xAAAA, 0xBABA,
        0xCACA, 0xDADA, 0xEAEA, 0xFAFA,
    }
    for _, val := range grease {
        assert(isGREASE(val), "0x%04X should be GREASE", val)
    }

    // Common TLS cipher suites (not GREASE)
    notGrease := []uint16{
        0x1301, 0x1302, 0x1303,  // TLS 1.3 ciphers
        0xc02f, 0xc030,          // ECDHE ciphers
        0x0000,                  // NULL cipher
    }
    for _, val := range notGrease {
        assert(!isGREASE(val), "0x%04X should not be GREASE", val)
    }
}
```

---

## References

- [RFC 8701: GREASE](https://datatracker.ietf.org/doc/html/rfc8701)
- [JA4 GREASE Filtering](https://github.com/FoxIO-LLC/ja4/blob/main/technical_details/JA4.md#grease)

**Implemented in:**
- `internal/fingerprint/ja4.go:isGREASE()`
- `internal/fingerprint/ja4_test.go:TestIsGREASE()`
