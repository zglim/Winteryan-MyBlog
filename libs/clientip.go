package libs

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP 从 HTTP 请求中提取客户端真实 IP。
// 优先级：X-Forwarded-For（取第一个有效地址）> X-Real-IP > RemoteAddr。
// 返回的 IP 已去除端口和多余空白，IPv6 地址保持标准格式（无方括号）。
func GetClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	// 1. X-Forwarded-For：多级代理场景下取第一个非空、非 unknown 的 IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, part := range strings.Split(xff, ",") {
			ip := strings.TrimSpace(part)
			if ip != "" && !strings.EqualFold(ip, "unknown") {
				return normalizeIP(ip)
			}
		}
	}

	// 2. X-Real-IP：Nginx 等反向代理常用此头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" && !strings.EqualFold(ip, "unknown") {
			return normalizeIP(ip)
		}
	}

	// 3. RemoteAddr：直接用 net.SplitHostPort 解析，正确处理 IPv6
	return normalizeIP(r.RemoteAddr)
}

// normalizeIP 去掉端口号并整理 IP 格式。
// 输入可以是 "ip:port"、"[ipv6]:port"、纯 IP 或带方括号的 IPv6。
func normalizeIP(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}

	// 先尝试 SplitHostPort，处理 "ip:port" 或 "[ipv6]:port" 格式
	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		return host
	}

	// 若 SplitHostPort 失败，可能是纯 IP（无端口）或格式异常
	// 去掉 IPv6 方括号
	addr = strings.TrimPrefix(addr, "[")
	addr = strings.TrimSuffix(addr, "]")
	return addr
}
