package validate

import "testing"

func TestIP(t *testing.T) {
	tests := []struct {
		ip      string
		wantErr bool
	}{
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"::1", false},
		{"2001:db8::1", false},
		{"invalid", true},
		{"", true},
		{"999.999.999.999", true},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			err := IP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("IP(%q) error = %v, wantErr %v", tt.ip, err, tt.wantErr)
			}
		})
	}
}

func TestCIDR(t *testing.T) {
	tests := []struct {
		cidr    string
		wantErr bool
	}{
		{"10.0.0.0/8", false},
		{"192.168.1.0/24", false},
		{"::1/128", false},
		{"invalid", true},
		{"10.0.0.0", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			err := CIDR(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("CIDR(%q) error = %v, wantErr %v", tt.cidr, err, tt.wantErr)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.com", false},
		{"http://localhost:8080", false},
		{"https://hooks.slack.com/services/T/B/x", false},
		{"not-a-url", true},
		{"://missing-scheme", true},
		{"", true},
		{"ftp://files.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := URL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("URL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestPort(t *testing.T) {
	tests := []struct {
		port    int
		wantErr bool
	}{
		{80, false},
		{443, false},
		{8080, false},
		{1, false},
		{65535, false},
		{0, true},
		{-1, true},
		{65536, true},
		{100000, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := Port(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("Port(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

func TestSeverity(t *testing.T) {
	tests := []struct {
		severity string
		wantErr  bool
	}{
		{"critical", false},
		{"high", false},
		{"medium", false},
		{"low", false},
		{"info", false},
		{"Critical", false},
		{"HIGH", false},
		{"invalid", true},
		{"", true},
		{"warning", true},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			err := Severity(tt.severity)
			if (err != nil) != tt.wantErr {
				t.Errorf("Severity(%q) error = %v, wantErr %v", tt.severity, err, tt.wantErr)
			}
		})
	}
}

func TestNonEmpty(t *testing.T) {
	tests := []struct {
		field   string
		value   string
		wantErr bool
	}{
		{"name", "hello", false},
		{"name", "  hello  ", false},
		{"name", "", true},
		{"name", "   ", true},
		{"name", "\t\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := NonEmpty(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NonEmpty(%q, %q) error = %v, wantErr %v", tt.field, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestOneOf(t *testing.T) {
	allowed := []string{"sqlite", "postgres"}

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"sqlite", false},
		{"postgres", false},
		{"mysql", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := OneOf("storage.type", tt.value, allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("OneOf(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
