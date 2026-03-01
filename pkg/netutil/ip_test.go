package netutil

import "testing"

func TestNormalizeIP(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ipv4", "192.168.1.1", "192.168.1.1"},
		{"ipv4 mapped", "::ffff:192.168.1.1", "192.168.1.1"},
		{"ipv6", "2001:db8::1", "2001:db8::1"},
		{"loopback", "127.0.0.1", "127.0.0.1"},
		{"invalid", "not-an-ip", "not-an-ip"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeIP(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeIP(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsPrivate(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"127.0.0.1", true},
		{"169.254.1.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := IsPrivate(tt.ip)
			if got != tt.want {
				t.Errorf("IsPrivate(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestInCIDR(t *testing.T) {
	tests := []struct {
		ip   string
		cidr string
		want bool
	}{
		{"192.168.1.100", "192.168.1.0/24", true},
		{"192.168.2.1", "192.168.1.0/24", false},
		{"10.0.0.5", "10.0.0.0/8", true},
		{"invalid", "10.0.0.0/8", false},
		{"10.0.0.5", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip+"_"+tt.cidr, func(t *testing.T) {
			got := InCIDR(tt.ip, tt.cidr)
			if got != tt.want {
				t.Errorf("InCIDR(%q, %q) = %v, want %v", tt.ip, tt.cidr, got, tt.want)
			}
		})
	}
}

func TestInAnyCIDR(t *testing.T) {
	cidrs := []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

	tests := []struct {
		ip   string
		want bool
	}{
		{"10.0.0.1", true},
		{"172.16.5.5", true},
		{"192.168.1.1", true},
		{"8.8.8.8", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := InAnyCIDR(tt.ip, cidrs)
			if got != tt.want {
				t.Errorf("InAnyCIDR(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestInAnyCIDR_InvalidCIDR(t *testing.T) {
	cidrs := []string{"invalid", "10.0.0.0/8"}
	got := InAnyCIDR("10.0.0.1", cidrs)
	if !got {
		t.Error("InAnyCIDR should skip invalid CIDRs and match valid ones")
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"192.168.1.1", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"invalid", false},
		{"", false},
		{"999.999.999.999", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := IsValidIP(tt.ip)
			if got != tt.want {
				t.Errorf("IsValidIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsIPv6(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"::1", true},
		{"2001:db8::1", true},
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := IsIPv6(tt.ip)
			if got != tt.want {
				t.Errorf("IsIPv6(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
