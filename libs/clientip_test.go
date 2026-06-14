package libs

import (
	"net/http"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name        string
		remoteAddr  string
		xff         string // X-Forwarded-For
		xri         string // X-Real-IP
		expectedIP  string
	}{
		// --- RemoteAddr 基础场景 ---
		{
			name:       "IPv4 RemoteAddr with port",
			remoteAddr: "192.168.1.100:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name:       "IPv4 RemoteAddr without port",
			remoteAddr: "10.0.0.1",
			expectedIP: "10.0.0.1",
		},
		{
			name:       "IPv6 RemoteAddr with bracket and port",
			remoteAddr: "[::1]:8080",
			expectedIP: "::1",
		},
		{
			name:       "IPv6 full address RemoteAddr",
			remoteAddr: "[2001:db8::1]:443",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "IPv6 RemoteAddr without port",
			remoteAddr: "::1",
			expectedIP: "::1",
		},

		// --- X-Real-IP ---
		{
			name:       "X-Real-IP IPv4",
			remoteAddr: "127.0.0.1:9999",
			xri:        "203.0.113.50",
			expectedIP: "203.0.113.50",
		},
		{
			name:       "X-Real-IP IPv6",
			remoteAddr: "127.0.0.1:9999",
			xri:        "2001:db8::ff",
			expectedIP: "2001:db8::ff",
		},
		{
			name:       "X-Real-IP with whitespace",
			remoteAddr: "127.0.0.1:9999",
			xri:        "  10.0.0.5  ",
			expectedIP: "10.0.0.5",
		},
		{
			name:       "X-Real-IP unknown fallback to RemoteAddr",
			remoteAddr: "172.16.0.1:80",
			xri:        "unknown",
			expectedIP: "172.16.0.1",
		},

		// --- X-Forwarded-For ---
		{
			name:       "XFF single IP",
			remoteAddr: "127.0.0.1:9999",
			xff:        "203.0.113.10",
			expectedIP: "203.0.113.10",
		},
		{
			name:       "XFF multiple proxies - take first",
			remoteAddr: "127.0.0.1:9999",
			xff:        "203.0.113.10, 70.41.3.18, 150.172.238.178",
			expectedIP: "203.0.113.10",
		},
		{
			name:       "XFF with extra whitespace",
			remoteAddr: "127.0.0.1:9999",
			xff:        "  198.51.100.1 ,  10.0.0.1 ",
			expectedIP: "198.51.100.1",
		},
		{
			name:       "XFF IPv6 client",
			remoteAddr: "127.0.0.1:9999",
			xff:        "2001:db8::1, 10.0.0.1",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "XFF unknown first, take second",
			remoteAddr: "127.0.0.1:9999",
			xff:        "unknown, 203.0.113.20",
			expectedIP: "203.0.113.20",
		},
		{
			name:       "XFF all unknown fallback to X-Real-IP",
			remoteAddr: "127.0.0.1:9999",
			xff:        "unknown",
			xri:        "10.0.0.99",
			expectedIP: "10.0.0.99",
		},

		// --- 优先级：XFF > X-Real-IP > RemoteAddr ---
		{
			name:       "XFF takes priority over X-Real-IP",
			remoteAddr: "127.0.0.1:9999",
			xff:        "1.2.3.4",
			xri:        "5.6.7.8",
			expectedIP: "1.2.3.4",
		},
		{
			name:       "X-Real-IP takes priority over RemoteAddr",
			remoteAddr: "127.0.0.1:9999",
			xri:        "5.6.7.8",
			expectedIP: "5.6.7.8",
		},

		// --- 空 / 异常场景 ---
		{
			name:       "empty RemoteAddr",
			remoteAddr: "",
			expectedIP: "",
		},
		{
			name:       "nil request",
			remoteAddr: "",
			expectedIP: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil request" {
				got := GetClientIP(nil)
				if got != tt.expectedIP {
					t.Errorf("GetClientIP(nil) = %q, want %q", got, tt.expectedIP)
				}
				return
			}

			req, _ := http.NewRequest("GET", "http://example.com", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			got := GetClientIP(req)
			if got != tt.expectedIP {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.expectedIP)
			}
		})
	}
}

func TestNormalizeIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.1:8080", "192.168.1.1"},
		{"192.168.1.1", "192.168.1.1"},
		{"[::1]:8080", "::1"},
		{"::1", "::1"},
		{"[2001:db8::1]", "2001:db8::1"},
		{"  10.0.0.1  ", "10.0.0.1"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeIP(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeIP(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
