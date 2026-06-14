package libs

import (
	"net"
	"strings"
)

// ClientIP derives a single, stable client IP string from the request's
// X-Forwarded-For / X-Real-IP headers and its RemoteAddr.
//
// The same normalization is used both when the auth cookie is issued
// (LoginController.Login) and when it is later verified (BaseController.auth),
// so the IP that participates in the cookie signature stays identical across
// requests of the same session — even behind a reverse proxy or over IPv6.
//
// Resolution order:
//  1. X-Forwarded-For: the left-most entry is the original client. The header
//     may look like "client, proxy1, proxy2"; only the first parseable IP is
//     used, never the whole raw header value.
//  2. X-Real-IP: a single IP set by the proxy.
//  3. RemoteAddr: usually "host:port" (IPv6 as "[::1]:port").
//
// The returned IP is canonicalized via net.IP so that differing textual
// representations of the same address (bracketed vs. bare, upper vs. lower
// case, zone-suffixed, IPv4-mapped) all collapse to one stable form.
func ClientIP(xForwardedFor, xRealIP, remoteAddr string) string {
	// 1. X-Forwarded-For — take the left-most valid entry (the real client).
	if xForwardedFor != "" {
		for _, part := range strings.Split(xForwardedFor, ",") {
			if ip := normalizeIP(part); ip != "" {
				return ip
			}
		}
	}

	// 2. X-Real-IP — a single value provided by the proxy.
	if ip := normalizeIP(xRealIP); ip != "" {
		return ip
	}

	// 3. RemoteAddr — strip the port, then normalize the host.
	if remoteAddr != "" {
		host := remoteAddr
		if h, _, err := net.SplitHostPort(remoteAddr); err == nil {
			host = h
		}
		if ip := normalizeIP(host); ip != "" {
			return ip
		}
		// Not a parseable IP (e.g. "localhost"); keep it deterministic.
		return strings.TrimSpace(host)
	}

	return ""
}

// normalizeIP returns the canonical string form of a single IP address, or ""
// if the input is not a valid IP. It tolerates surrounding brackets and an
// IPv6 zone identifier so values pulled straight from headers or RemoteAddr
// can be fed in directly.
func normalizeIP(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Drop surrounding brackets, e.g. "[::1]".
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	// Drop an IPv6 zone identifier, e.g. "fe80::1%eth0".
	if i := strings.IndexByte(s, '%'); i >= 0 {
		s = s[:i]
	}
	ip := net.ParseIP(strings.TrimSpace(s))
	if ip == nil {
		return ""
	}
	return ip.String()
}
