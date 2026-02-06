package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// UUID4 generates a random UUID v4 string.
// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
func UUID4() string {
	var uuid [16]byte
	if _, err := io.ReadFull(rand.Reader, uuid[:]); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// RandomBytes returns n cryptographically random bytes.
func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, fmt.Errorf("generating random bytes: %w", err)
	}
	return b, nil
}

// RandomHex returns a hex-encoded string of n random bytes (2n chars).
func RandomHex(n int) (string, error) {
	b, err := RandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RandomToken generates a 32-byte hex token (64 chars).
// Suitable for honey tokens and session identifiers.
func RandomToken() (string, error) {
	return RandomHex(32)
}
