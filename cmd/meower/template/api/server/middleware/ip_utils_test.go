package middleware

import (
	"testing"
)

func TestIsTrustedProxy(t *testing.T) {
	trustedRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.1.0/24",
	}

	tests := []struct {
		name string
		ip   string
		want bool
	}{
		// Valid trusted IPs
		{"Trusted IP in 10.0.0.0/8 - start", "10.0.0.1", true},
		{"Trusted IP in 10.0.0.0/8 - mid", "10.128.5.10", true},
		{"Trusted IP in 172.16.0.0/12", "172.16.0.1", true},
		{"Trusted IP in 192.168.1.0/24", "192.168.1.100", true},
		{"Edge of range - 10.0.0.0/8 upper", "10.255.255.255", true},
		{"Edge of range - 192.168.1.0/24 upper", "192.168.1.254", true},

		// Untrusted IPs
		{"Untrusted public IP", "1.2.3.4", false},
		{"Outside range - just above 10.0.0.0/8", "11.0.0.1", false},
		{"Outside range - 192.168.2.x", "192.168.2.1", false},
		{"Outside range - 172.15.x.x", "172.15.255.255", false},

		// IPs with port notation
		{"IPv4 with port - trusted", "10.0.0.1:12345", true},
		{"IPv4 with port - untrusted", "1.2.3.4:8080", false},
		{"IPv4 with port - edge case", "192.168.1.1:443", true},

		// IPv6
		{"IPv6 loopback", "::1", false},
		{"IPv6 with brackets", "[::1]", false},
		{"IPv6 with port", "[::1]:8080", false},

		// Invalid inputs
		{"Invalid IP", "not-an-ip", false},
		{"Empty string", "", false},
		{"Malformed IP", "300.300.300.300", false},
		{"Just port number", ":8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTrustedProxy(tt.ip, trustedRanges)
			if got != tt.want {
				t.Errorf("isTrustedProxy(%q, trustedRanges) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsTrustedProxy_EmptyTrustedList(t *testing.T) {
	emptyRanges := []string{}

	tests := []struct {
		name string
		ip   string
	}{
		{"Valid private IP", "10.0.0.1"},
		{"Valid public IP", "8.8.8.8"},
		{"Localhost", "127.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTrustedProxy(tt.ip, emptyRanges)
			if got != false {
				t.Errorf("isTrustedProxy(%q, emptyRanges) = %v, want false (no proxies should be trusted with empty list)", tt.ip, got)
			}
		})
	}
}

func TestIsTrustedProxy_InvalidCIDR(t *testing.T) {
	invalidRanges := []string{
		"not-a-cidr",
		"10.0.0.0/33",      // Invalid netmask
		"192.168.1.0",      // Missing netmask
		"256.256.256.0/24", // Invalid IP
	}

	// Should not panic, should return false for valid IPs
	ip := "10.0.0.1"
	got := isTrustedProxy(ip, invalidRanges)
	if got != false {
		t.Errorf("isTrustedProxy(%q, invalidRanges) = %v, want false (invalid CIDR should be skipped)", ip, got)
	}
}

func TestExtractIPFromForwardedFor(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		// Standard cases
		{"Single IP", "192.168.1.1", "192.168.1.1"},
		{"Multiple IPs - client first", "192.168.1.1, 10.0.0.1, 172.16.0.1", "192.168.1.1"},
		{"Three IPs", "203.0.113.1, 198.51.100.1, 192.0.2.1", "203.0.113.1"},

		// Spacing variations
		{"With spaces", "  192.168.1.1  , 10.0.0.1", "192.168.1.1"},
		{"No spaces after comma", "192.168.1.1,10.0.0.1", "192.168.1.1"},
		{"Mixed spacing", "192.168.1.1  ,  10.0.0.1  , 172.16.0.1", "192.168.1.1"},

		// Edge cases
		{"Empty string", "", ""},
		{"Only commas", ",,", ""},
		{"Single IP with trailing comma", "192.168.1.1,", "192.168.1.1"},

		// IPv6
		{"IPv6 single", "2001:db8::1", "2001:db8::1"},
		{"IPv6 multiple", "2001:db8::1, 2001:db8::2", "2001:db8::1"},
		{"IPv6 with spaces", "  2001:db8::1  , 2001:db8::2", "2001:db8::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIPFromForwardedFor(tt.header)
			if got != tt.want {
				t.Errorf("extractIPFromForwardedFor(%q) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}

func TestIsTrustedProxy_RealWorldScenarios(t *testing.T) {
	// Common cloud provider private ranges
	awsVpcRanges := []string{
		"10.0.0.0/8",    // AWS VPC common range
		"172.31.0.0/16", // AWS default VPC
	}

	tests := []struct {
		name string
		ip   string
		want bool
	}{
		// AWS scenarios
		{"AWS VPC internal", "10.1.2.3", true},
		{"AWS default VPC", "172.31.5.10", true},
		{"AWS ELB external IP", "52.1.2.3", false},
		{"Public internet", "8.8.8.8", false},

		// Common reverse proxy scenarios
		{"Nginx internal", "10.0.1.50", true},
		{"HAProxy internal", "172.31.0.100", true},
		{"Client direct connection", "203.0.113.45", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTrustedProxy(tt.ip, awsVpcRanges)
			if got != tt.want {
				t.Errorf("isTrustedProxy(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
