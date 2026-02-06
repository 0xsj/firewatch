# Crypto, Binary, and Encoding

## crypto/rand vs math/rand

Go has two `rand` packages:

| Package      | Use for                     | Deterministic? |
|--------------|-----------------------------|----------------|
| `math/rand`  | Shuffling, sampling, jitter | Yes (seeded)   |
| `crypto/rand`| Tokens, keys, UUIDs         | No (OS entropy)|

For anything security-related, always use `crypto/rand`:

```go
import "crypto/rand"

b := make([]byte, 16)
io.ReadFull(rand.Reader, b)
```

`rand.Reader` is an `io.Reader` backed by the OS CSPRNG (`/dev/urandom`
on Linux, `CryptGenRandom` on Windows).

`io.ReadFull` reads exactly `len(b)` bytes or returns an error — unlike
`rand.Read` which could theoretically short-read.

**Used in:** `pkg/crypto/random.go`

---

## Hashing

```go
import (
    "crypto/md5"
    "crypto/sha256"
    "encoding/hex"
)

// SHA-256: returns [32]byte (fixed-size array, not a slice)
h := sha256.Sum256(data)
hex.EncodeToString(h[:])  // h[:] converts array → slice

// MD5: returns [16]byte
h := md5.Sum(data)
hex.EncodeToString(h[:])
```

**Why `h[:]`?** — `Sum256` returns `[32]byte` (array), but `hex.EncodeToString`
wants `[]byte` (slice). The `[:]` slice operator converts.

MD5 is broken for security but still used in JA3 fingerprinting as a
compact representation. We keep it for compatibility with existing
fingerprint databases.

**Used in:** `pkg/crypto/hash.go`

---

## Bit Manipulation (UUID v4)

UUID v4 requires setting specific bits in a 16-byte random value:

```go
uuid[6] = (uuid[6] & 0x0f) | 0x40  // version = 4
uuid[8] = (uuid[8] & 0x3f) | 0x80  // variant = 10
```

Breaking it down:

### Byte 6 — version field (high nibble)

```
uuid[6] & 0x0f    // clear top 4 bits:  xxxx_yyyy → 0000_yyyy
        | 0x40    // set version 4:     0000_yyyy → 0100_yyyy
```

- `0x0f` = `0000_1111` — mask keeps bottom 4 bits
- `0x40` = `0100_0000` — sets bit 6 (version = 4)

### Byte 8 — variant field (top 2 bits)

```
uuid[8] & 0x3f    // clear top 2 bits:  xx_yyyyyy → 00_yyyyyy
        | 0x80    // set variant 10:    00_yyyyyy → 10_yyyyyy
```

- `0x3f` = `0011_1111` — mask keeps bottom 6 bits
- `0x80` = `1000_0000` — sets bit 7 (variant = RFC 4122)

### fmt.Sprintf formatting

```go
fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
    uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
```

- `%x` formats bytes as hex
- `%08x` pads to 8 hex chars with leading zeros
- Slice ranges select the right byte groups for each UUID section

**Used in:** `pkg/crypto/random.go`

---

## hex.EncodeToString

Converts raw bytes to their hex string representation:

```go
[]byte{0xde, 0xad, 0xbe, 0xef} → "deadbeef"
```

Each byte becomes 2 hex characters, so `n` bytes → `2n` character string.
`RandomHex(32)` produces a 64-character string.

**Used in:** `pkg/crypto/hash.go`, `pkg/crypto/random.go`
