package fingerprint

import (
	"crypto/tls"
	"testing"
)

func TestJA4(t *testing.T) {
	tests := []struct {
		name       string
		hello      *TLSClientHello
		serverName string
		remoteAddr string
		wantPrefix string // Check prefix since hashes are deterministic but complex
	}{
		{
			name: "TLS 1.3 with domain SNI",
			hello: &TLSClientHello{
				Version:      tls.VersionTLS13,
				CipherSuites: []uint16{0x1301, 0x1302, 0x1303},
				Curves:       []tls.CurveID{29, 23, 24},
				PointFormats: []uint8{0},
			},
			serverName: "example.com",
			remoteAddr: "192.168.1.1:443",
			wantPrefix: "t13d03",
		},
		{
			name: "TLS 1.2 with IP address",
			hello: &TLSClientHello{
				Version:      tls.VersionTLS12,
				CipherSuites: []uint16{0xc02f, 0xc030, 0x009e},
				Curves:       []tls.CurveID{23, 24},
			},
			serverName: "10.0.0.1",
			remoteAddr: "10.0.0.1:443",
			wantPrefix: "t12i03",
		},
		{
			name: "Empty ciphers",
			hello: &TLSClientHello{
				Version:      tls.VersionTLS13,
				CipherSuites: []uint16{},
			},
			serverName: "test.local",
			wantPrefix: "t13d00",
		},
		{
			name:       "Nil hello",
			hello:      nil,
			serverName: "example.com",
			wantPrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JA4(tt.hello, tt.serverName, tt.remoteAddr)

			if tt.wantPrefix == "" {
				if got != "" {
					t.Errorf("JA4() = %q, want empty string", got)
				}
				return
			}

			if len(got) < len(tt.wantPrefix) {
				t.Errorf("JA4() = %q, too short (expected at least %d chars)", got, len(tt.wantPrefix))
				return
			}

			if got[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("JA4() prefix = %q, want %q (full: %s)", got[:len(tt.wantPrefix)], tt.wantPrefix, got)
			}

			// Verify format: metadata_hash1_hash2
			parts := splitJA4(got)
			if len(parts) != 3 {
				t.Errorf("JA4() = %q, expected 3 parts separated by '_', got %d", got, len(parts))
			}

			// Metadata should be 10 chars: protocol(1) + version(2) + sni(1) + ciphers(2) + ext(2) + alpn(2)
			if len(parts[0]) != 10 {
				t.Errorf("JA4() metadata = %q, expected 10 chars", parts[0])
			}

			// Each hash should be 12 chars
			if len(parts[1]) != 12 {
				t.Errorf("JA4() cipher hash = %q, expected 12 chars", parts[1])
			}
			if len(parts[2]) != 12 {
				t.Errorf("JA4() extension hash = %q, expected 12 chars", parts[2])
			}
		})
	}
}

func TestTLSVersionToJA4(t *testing.T) {
	tests := []struct {
		version uint16
		want    string
	}{
		{tls.VersionTLS13, "13"},
		{tls.VersionTLS12, "12"},
		{tls.VersionTLS11, "11"},
		{tls.VersionTLS10, "10"},
		{tls.VersionSSL30, "s3"}, //nolint:staticcheck // testing fingerprinting for legacy clients
		{0xFFFF, "00"},           // unknown version
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tlsVersionToJA4(tt.version)
			if got != tt.want {
				t.Errorf("tlsVersionToJA4(0x%04x) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestSNIIndicator(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		remoteAddr string
		want       string
	}{
		{"domain", "example.com", "1.2.3.4:443", "d"},
		{"ipv4 in SNI", "192.168.1.1", "192.168.1.1:443", "i"},
		{"ipv6 in SNI", "::1", "[::1]:443", "i"},
		{"no SNI", "", "10.0.0.1:443", "i"},
		{"empty", "", "", "i"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sniIndicator(tt.serverName, tt.remoteAddr)
			if got != tt.want {
				t.Errorf("sniIndicator(%q, %q) = %q, want %q", tt.serverName, tt.remoteAddr, got, tt.want)
			}
		})
	}
}

func TestIsGREASE(t *testing.T) {
	tests := []struct {
		val  uint16
		want bool
	}{
		// GREASE values (pattern: 0x?A?A)
		{0x0A0A, true},
		{0x1A1A, true},
		{0x2A2A, true},
		{0x3A3A, true},
		{0x4A4A, true},
		{0x5A5A, true},
		{0x6A6A, true},
		{0x7A7A, true},
		{0x8A8A, true},
		{0x9A9A, true},
		{0xAAAA, true},
		{0xBABA, true},
		{0xCACA, true},
		{0xDADA, true},
		{0xEAEA, true},
		{0xFAFA, true},

		// Non-GREASE values
		{0x0000, false},
		{0x1301, false}, // TLS_AES_128_GCM_SHA256
		{0x1302, false}, // TLS_AES_256_GCM_SHA384
		{0xc02f, false}, // TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		{0x0A0B, false}, // different bytes
		{0x1A2A, false}, // pattern broken
		{0xFFFF, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := isGREASE(tt.val)
			if got != tt.want {
				t.Errorf("isGREASE(0x%04X) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestHashCiphers(t *testing.T) {
	tests := []struct {
		name    string
		ciphers []uint16
		want    string
	}{
		{
			name:    "empty",
			ciphers: []uint16{},
			want:    "000000000000",
		},
		{
			name:    "single cipher",
			ciphers: []uint16{0x1301},
			want:    hashCiphers([]uint16{0x1301}), // deterministic
		},
		{
			name:    "with GREASE",
			ciphers: []uint16{0x1A1A, 0x1301, 0x1302},
			want:    hashCiphers([]uint16{0x1301, 0x1302}), // GREASE filtered
		},
		{
			name:    "all GREASE",
			ciphers: []uint16{0x0A0A, 0x1A1A, 0x2A2A},
			want:    "000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashCiphers(tt.ciphers)
			if len(got) != 12 {
				t.Errorf("hashCiphers() = %q, want 12 chars", got)
			}
			if tt.want != "" && got != tt.want {
				t.Errorf("hashCiphers() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHashExtensions(t *testing.T) {
	tests := []struct {
		name   string
		curves []tls.CurveID
	}{
		{"empty", []tls.CurveID{}},
		{"single", []tls.CurveID{29}},
		{"multiple", []tls.CurveID{29, 23, 24}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashExtensions(tt.curves)
			if len(got) != 12 {
				t.Errorf("hashExtensions() = %q, want 12 chars", got)
			}
		})
	}
}

func TestJA4FromClientHello(t *testing.T) {
	t.Run("nil info", func(t *testing.T) {
		got := JA4FromClientHello(nil)
		if got != "" {
			t.Errorf("JA4FromClientHello(nil) = %q, want empty", got)
		}
	})

	t.Run("valid info", func(t *testing.T) {
		info := &tls.ClientHelloInfo{
			CipherSuites:      []uint16{0x1301, 0x1302},
			SupportedCurves:   []tls.CurveID{29, 23},
			SupportedPoints:   []uint8{0},
			SupportedVersions: []uint16{tls.VersionTLS13},
			ServerName:        "example.com",
		}

		got := JA4FromClientHello(info)
		if got == "" {
			t.Error("JA4FromClientHello() returned empty string")
		}

		// Should start with t13d (TLS 1.3, domain)
		if len(got) < 4 || got[:4] != "t13d" {
			t.Errorf("JA4FromClientHello() = %q, want prefix 't13d'", got)
		}
	})
}

// Helper function to split JA4 string into parts
func splitJA4(ja4 string) []string {
	parts := make([]string, 0, 3)
	start := 0
	for i, c := range ja4 {
		if c == '_' {
			parts = append(parts, ja4[start:i])
			start = i + 1
		}
	}
	if start < len(ja4) {
		parts = append(parts, ja4[start:])
	}
	return parts
}
