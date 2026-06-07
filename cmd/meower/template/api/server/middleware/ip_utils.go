package middleware

import (
	"net"
	"strings"
)

// isTrustedProxy checks if the given IP address is in the list of trusted proxy CIDR ranges
func isTrustedProxy(ip string, trustedRanges []string) bool {
	// Handle IPv6 with port notation (e.g., "[::1]:12345")
	if strings.HasPrefix(ip, "[") {
		if idx := strings.Index(ip, "]"); idx != -1 {
			ip = ip[1:idx]
		}
	} else {
		// Handle IPv4 with port notation (e.g., "192.168.1.1:12345")
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
	}

	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	// Check against each trusted CIDR range
	for _, cidr := range trustedRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			// Log error but continue checking other ranges
			continue
		}
		if ipNet.Contains(clientIP) {
			return true
		}
	}

	return false
}

// extractIPFromForwardedFor extracts the leftmost (original client) IP from X-Forwarded-For
func extractIPFromForwardedFor(header string) string {
	// X-Forwarded-For format: "client, proxy1, proxy2"
	// The leftmost IP is the original client
	ips := strings.Split(header, ",")
	if len(ips) > 0 {
		return strings.TrimSpace(ips[0])
	}
	return ""
}
