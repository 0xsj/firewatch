package deception

import (
	"encoding/base64"

	"github.com/0xsj/firewatch/pkg/crypto"
)

// GenerateAWSAccessKey returns an AKIA-prefixed fake AWS access key ID.
// Format: AKIA + 16 uppercase alphanumeric characters.
func GenerateAWSAccessKey() string {
	b, _ := crypto.RandomBytes(16)
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 20)
	copy(result, "AKIA")
	for i := 0; i < 16; i++ {
		result[4+i] = charset[int(b[i])%len(charset)]
	}
	return string(result)
}

// GenerateAWSSecretKey returns a 40-character base64-like fake AWS secret key.
func GenerateAWSSecretKey() string {
	b, _ := crypto.RandomBytes(30)
	return base64.StdEncoding.EncodeToString(b)[:40]
}

// GenerateSessionToken returns a fake AWS session token.
// Format: FwoGZXIvYXdz + 52 random base64 characters.
func GenerateSessionToken() string {
	b, _ := crypto.RandomBytes(39)
	return "FwoGZXIvYXdz" + base64.StdEncoding.EncodeToString(b)[:52]
}

// GenerateIMDSToken returns a fake IMDSv2 token.
// Format: AQAAANjCpMCZjg_ + 32 hex characters.
func GenerateIMDSToken() string {
	hex, _ := crypto.RandomHex(16)
	return "AQAAANjCpMCZjg_" + hex
}

// GenerateAPIKey returns a fake API key with the given prefix.
// Format: {prefix}_ + 24 hex characters.
func GenerateAPIKey(prefix string) string {
	hex, _ := crypto.RandomHex(12)
	return prefix + "_" + hex
}

// GenerateDBPassword returns a 20-character mixed alphanumeric password.
func GenerateDBPassword() string {
	b, _ := crypto.RandomBytes(20)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 20)
	for i, v := range b {
		result[i] = charset[int(v)%len(charset)]
	}
	return string(result)
}
