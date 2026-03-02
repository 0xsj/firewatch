package deception

import (
	"regexp"
	"testing"
)

func TestGenerateAWSAccessKey_Format(t *testing.T) {
	re := regexp.MustCompile(`^AKIA[A-Z0-9]{16}$`)
	for i := 0; i < 100; i++ {
		key := GenerateAWSAccessKey()
		if !re.MatchString(key) {
			t.Fatalf("iteration %d: key %q does not match AKIA + 16 alphanum", i, key)
		}
	}
}

func TestGenerateAWSAccessKey_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key := GenerateAWSAccessKey()
		if seen[key] {
			t.Fatalf("duplicate key at iteration %d: %s", i, key)
		}
		seen[key] = true
	}
}

func TestGenerateAWSSecretKey_Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		key := GenerateAWSSecretKey()
		if len(key) != 40 {
			t.Fatalf("iteration %d: secret key length = %d, want 40", i, len(key))
		}
	}
}

func TestGenerateAWSSecretKey_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key := GenerateAWSSecretKey()
		if seen[key] {
			t.Fatalf("duplicate secret key at iteration %d", i)
		}
		seen[key] = true
	}
}

func TestGenerateSessionToken_Format(t *testing.T) {
	re := regexp.MustCompile(`^FwoGZXIvYXdz.{52}$`)
	for i := 0; i < 100; i++ {
		tok := GenerateSessionToken()
		if !re.MatchString(tok) {
			t.Fatalf("iteration %d: token %q does not match expected format (len=%d)", i, tok, len(tok))
		}
	}
}

func TestGenerateIMDSToken_Format(t *testing.T) {
	re := regexp.MustCompile(`^AQAAANjCpMCZjg_[a-f0-9]{32}$`)
	for i := 0; i < 100; i++ {
		tok := GenerateIMDSToken()
		if !re.MatchString(tok) {
			t.Fatalf("iteration %d: token %q does not match expected format", i, tok)
		}
	}
}

func TestGenerateAPIKey_Format(t *testing.T) {
	re := regexp.MustCompile(`^sk_live_[a-f0-9]{24}$`)
	for i := 0; i < 100; i++ {
		key := GenerateAPIKey("sk_live")
		if !re.MatchString(key) {
			t.Fatalf("iteration %d: key %q does not match expected format", i, key)
		}
	}
}

func TestGenerateDBPassword_Format(t *testing.T) {
	re := regexp.MustCompile(`^[a-zA-Z0-9]{20}$`)
	for i := 0; i < 100; i++ {
		pw := GenerateDBPassword()
		if !re.MatchString(pw) {
			t.Fatalf("iteration %d: password %q does not match expected format", i, pw)
		}
	}
}

func TestGenerateDBPassword_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		pw := GenerateDBPassword()
		if seen[pw] {
			t.Fatalf("duplicate password at iteration %d", i)
		}
		seen[pw] = true
	}
}
