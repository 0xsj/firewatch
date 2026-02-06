package crypto

import "testing"

func TestSHA256(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
	}
	for _, tt := range tests {
		got := SHA256String(tt.input)
		if got != tt.want {
			t.Errorf("SHA256String(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMD5(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
	}
	for _, tt := range tests {
		got := MD5String(tt.input)
		if got != tt.want {
			t.Errorf("MD5String(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
