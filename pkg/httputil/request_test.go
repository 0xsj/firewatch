package httputil

import (
	"net/http"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "from RemoteAddr",
			remoteAddr: "192.168.1.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "from X-Forwarded-For single",
			remoteAddr: "10.0.0.1:999",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			want:       "203.0.113.50",
		},
		{
			name:       "from X-Forwarded-For chain",
			remoteAddr: "10.0.0.1:999",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18, 150.172.238.178"},
			want:       "203.0.113.50",
		},
		{
			name:       "from X-Real-IP",
			remoteAddr: "10.0.0.1:999",
			headers:    map[string]string{"X-Real-Ip": "198.51.100.10"},
			want:       "198.51.100.10",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr: "10.0.0.1:999",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.50",
				"X-Real-Ip":       "198.51.100.10",
			},
			want: "203.0.113.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{
				RemoteAddr: tt.remoteAddr,
				Header:     http.Header{},
			}
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}

			got := ClientIP(r)
			if got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasHeader(t *testing.T) {
	r := &http.Request{Header: http.Header{}}
	r.Header.Set("Authorization", "Bearer token")

	if !HasHeader(r, "Authorization") {
		t.Error("HasHeader(Authorization) = false, want true")
	}
	if HasHeader(r, "X-Custom") {
		t.Error("HasHeader(X-Custom) = true, want false")
	}
}

func TestHeaderMap(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "text/html")

	m := HeaderMap(h)
	if m["Content-Type"] != "application/json" {
		t.Errorf("HeaderMap[Content-Type] = %q, want %q", m["Content-Type"], "application/json")
	}
	if m["Accept"] != "text/html" {
		t.Errorf("HeaderMap[Accept] = %q, want %q", m["Accept"], "text/html")
	}
}
