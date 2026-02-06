package crypto

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 returns the hex-encoded SHA-256 hash of data.
func SHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// SHA256String returns the hex-encoded SHA-256 hash of s.
func SHA256String(s string) string {
	return SHA256([]byte(s))
}

// MD5 returns the hex-encoded MD5 hash of data.
// Used for fingerprint comparison (JA3), not for security.
func MD5(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

// MD5String returns the hex-encoded MD5 hash of s.
func MD5String(s string) string {
	return MD5([]byte(s))
}
