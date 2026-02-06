package netutil

import "net"

// NormalizeIP parses and normalizes an IP address string.
// IPv4-mapped IPv6 addresses (::ffff:1.2.3.4) are converted to IPv4.
// Returns the normalized string, or the original input if parsing fails.
func NormalizeIP(raw string) string {
	ip := net.ParseIP(raw)
	if ip == nil {
		return raw
	}

	// Convert IPv4-mapped IPv6 to plain IPv4
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.String()
}

// IsPrivate reports whether the IP address is in a private/reserved range
// (RFC 1918, loopback, link-local).
func IsPrivate(raw string) bool {
	ip := net.ParseIP(raw)
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()
}

// InCIDR reports whether the IP address falls within the given CIDR block.
func InCIDR(raw string, cidr string) bool {
	ip := net.ParseIP(raw)
	if ip == nil {
		return false
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(ip)
}

// InAnyCIDR reports whether the IP address falls within any of the
// given CIDR blocks.
func InAnyCIDR(raw string, cidrs []string) bool {
	ip := net.ParseIP(raw)
	if ip == nil {
		return false
	}

	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// IsValidIP reports whether the string is a valid IPv4 or IPv6 address.
func IsValidIP(raw string) bool {
	return net.ParseIP(raw) != nil
}

// IsIPv6 reports whether the string is an IPv6 address.
func IsIPv6(raw string) bool {
	ip := net.ParseIP(raw)
	if ip == nil {
		return false
	}
	return ip.To4() == nil
}
