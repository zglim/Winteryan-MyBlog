package libs

import "testing"

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		xff        string
		xRealIP    string
		remoteAddr string
		want       string
	}{
		// --- plain RemoteAddr (no proxy) ---
		{"ipv4 with port", "", "", "203.0.113.7:54321", "203.0.113.7"},
		{"ipv4 without port", "", "", "192.0.2.5", "192.0.2.5"},
		{"ipv6 bracketed with port", "", "", "[2001:db8::1]:443", "2001:db8::1"},
		{"ipv6 loopback bare", "", "", "::1", "::1"},
		{"ipv6 bracketed bare", "", "", "[fe80::abcd]", "fe80::abcd"},
		{"hostname falls back to host", "", "", "localhost:8080", "localhost"},

		// --- X-Forwarded-For ---
		{"xff single", "198.51.100.23", "", "10.0.0.1:1", "198.51.100.23"},
		{"xff multi-proxy takes leftmost", "198.51.100.23, 70.41.3.18, 150.172.238.178", "", "10.0.0.1:1", "198.51.100.23"},
		{"xff trims spaces and ipv6", " 2001:db8::2 , 10.0.0.1 ", "", "10.0.0.1:1", "2001:db8::2"},
		{"xff invalid first entry skipped", "unknown, 203.0.113.5", "", "10.0.0.1:1", "203.0.113.5"},
		{"xff all invalid falls through to real-ip", "garbage, junk", "203.0.113.6", "10.0.0.1:1", "203.0.113.6"},

		// --- X-Real-IP ---
		{"real-ip when no xff", "", "203.0.113.9", "10.0.0.1:1", "203.0.113.9"},

		// --- priority ordering ---
		{"xff beats real-ip and remote", "198.51.100.1", "203.0.113.9", "10.0.0.1:1", "198.51.100.1"},
		{"real-ip beats remote", "", "203.0.113.9", "10.0.0.1:1", "203.0.113.9"},

		// --- canonicalization ---
		{"ipv6 expanded canonicalized", "", "2001:0DB8:0000:0000:0000:0000:0000:0001", "10.0.0.1:1", "2001:db8::1"},
		{"ipv6 zone stripped", "", "fe80::1%eth0", "10.0.0.1:1", "fe80::1"},
		{"ipv4-mapped ipv6", "", "::ffff:192.0.2.128", "10.0.0.1:1", "192.0.2.128"},

		// --- missing everything ---
		{"all empty", "", "", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClientIP(tc.xff, tc.xRealIP, tc.remoteAddr); got != tc.want {
				t.Errorf("ClientIP(%q,%q,%q) = %q, want %q", tc.xff, tc.xRealIP, tc.remoteAddr, got, tc.want)
			}
		})
	}
}

func TestNormalizeIP(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"[::1]", "::1"},
		{"  10.0.0.1 ", "10.0.0.1"},
		{"fe80::1%eth0", "fe80::1"},
		{"not-an-ip", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := normalizeIP(tc.in); got != tc.want {
			t.Errorf("normalizeIP(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// signAuth mirrors how the controllers build the auth-cookie signature, so the
// tests below assert exactly the invariant the bug was breaking: the IP signed
// at login time must equal the IP recomputed when the cookie is verified.
func signAuth(xff, xRealIP, remoteAddr string) string {
	const password = "5f4dcc3b5aa765d61d8327deb882cf99"
	const personalkey = "k3yk3y"
	return Md5([]byte(ClientIP(xff, xRealIP, remoteAddr) + "|" + password + personalkey))
}

// TestAuthSignatureStableAcrossProxyAndIPv6 covers the cookie-verification
// path: equivalent representations of the same client (issued at login vs. seen
// on a later proxied/IPv6 request) must produce an identical signature, while a
// genuinely different client must not.
func TestAuthSignatureStableAcrossProxyAndIPv6(t *testing.T) {
	// Cookie issued at login: direct IPv6 connection, bracketed host:port.
	loginSig := signAuth("", "", "[2001:DB8::1]:50000")

	equivalent := []struct {
		name       string
		xff        string
		xRealIP    string
		remoteAddr string
	}{
		{"same client via X-Real-IP (canonical)", "", "2001:db8::1", "10.0.0.1:1"},
		{"same client via X-Forwarded-For chain", "2001:db8::1, 10.0.0.2", "", "10.0.0.1:1"},
		{"same client bare ipv6 remoteaddr", "", "", "2001:db8::1"},
		{"same client expanded ipv6", "", "2001:0DB8:0000:0000:0000:0000:0000:0001", "10.0.0.1:1"},
	}
	for _, tc := range equivalent {
		t.Run("match/"+tc.name, func(t *testing.T) {
			if got := signAuth(tc.xff, tc.xRealIP, tc.remoteAddr); got != loginSig {
				t.Errorf("signature drifted for equivalent client: got %s want %s", got, loginSig)
			}
		})
	}

	// A different client must yield a different signature (cookie won't validate).
	if other := signAuth("", "", "203.0.113.99:443"); other == loginSig {
		t.Errorf("different client produced same signature %s", other)
	}
}
