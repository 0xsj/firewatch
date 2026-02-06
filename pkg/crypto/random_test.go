package crypto

import (
	"regexp"
	"testing"
)

func TestUUID4(t *testing.T) {
	uuid := UUID4()

	// Check format: 8-4-4-4-12 hex characters.
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !re.MatchString(uuid) {
		t.Errorf("UUID4() = %q, does not match UUID v4 pattern", uuid)
	}

	// Two UUIDs should be unique.
	a, b := UUID4(), UUID4()
	if a == b {
		t.Errorf("UUID4() returned duplicate values: %s", a)
	}
}

func TestRandomHex(t *testing.T) {
	hex, err := RandomHex(16)
	if err != nil {
		t.Fatalf("RandomHex(16): %v", err)
	}
	if len(hex) != 32 {
		t.Errorf("RandomHex(16) length = %d, want 32", len(hex))
	}
}

func TestRandomToken(t *testing.T) {
	tok, err := RandomToken()
	if err != nil {
		t.Fatalf("RandomToken(): %v", err)
	}
	if len(tok) != 64 {
		t.Errorf("RandomToken() length = %d, want 64", len(tok))
	}
}
