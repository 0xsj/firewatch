package fingerprint

import (
	"crypto/tls"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/0xsj/firewatch/pkg/crypto"
)

// JA4 computes the JA4 TLS client fingerprint from ClientHello parameters.
//
// JA4 format: [protocol][version][sni][cipher_count][extension_count][alpn]_[cipher_hash]_[extension_hash]
//
// Example: t13d1516h2_8daaf6152771_e5627efa2ab1
//
// Note: Go's tls.ClientHelloInfo does not expose the raw extensions list,
// ALPN values, or signature algorithms. This implementation produces a
// partial JA4 using available fields: protocol, version, SNI, cipher count.
// The extension hash section uses supported curves as a substitute.
//
// For full JA4 support, packet capture at the TLS layer would be required.
func JA4(hello *TLSClientHello, serverName string, remoteAddr string) (fingerprint string) {
	if hello == nil {
		return ""
	}

	// Section 1: Metadata (12 chars)
	protocol := "t" // TLS over TCP (we don't support QUIC/DTLS in this honeypot)
	version := tlsVersionToJA4(hello.Version)
	sni := sniIndicator(serverName, remoteAddr)
	cipherCount := fmt.Sprintf("%02d", len(hello.CipherSuites))
	if len(cipherCount) > 2 {
		cipherCount = "99" // cap at 99
	}

	// Extension count: We don't have raw extensions, so estimate based on
	// what we know: curves, points, and a few standard extensions.
	// This is a limitation of Go's TLS API.
	extensionCount := estimateExtensionCount(hello)

	// ALPN: Not available from Go's ClientHelloInfo.
	// Default to "00" (no ALPN).
	alpn := "00"

	metadata := protocol + version + sni + cipherCount + extensionCount + alpn

	// Section 2: Cipher hash (12 chars)
	cipherHash := hashCiphers(hello.CipherSuites)

	// Section 3: Extension hash (12 chars)
	// Since we don't have raw extensions, hash the curves as a proxy.
	// This provides some fingerprinting signal, though not complete JA4.
	extensionHash := hashExtensions(hello.Curves)

	return metadata + "_" + cipherHash + "_" + extensionHash
}

// JA4FromClientHello extracts TLS parameters from Go's ClientHelloInfo
// and computes the JA4 fingerprint.
func JA4FromClientHello(info *tls.ClientHelloInfo) string {
	if info == nil {
		return ""
	}

	hello := &TLSClientHello{
		CipherSuites: info.CipherSuites,
		Curves:       info.SupportedCurves,
		PointFormats: info.SupportedPoints,
	}

	// Get the highest supported version (TLS 1.3, TLS 1.2, etc.)
	if len(info.SupportedVersions) > 0 {
		hello.Version = info.SupportedVersions[0]
	}

	remoteAddr := ""
	if info.Conn != nil {
		remoteAddr = info.Conn.RemoteAddr().String()
	}

	return JA4(hello, info.ServerName, remoteAddr)
}

// tlsVersionToJA4 converts a TLS version constant to JA4 format.
func tlsVersionToJA4(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "13"
	case tls.VersionTLS12:
		return "12"
	case tls.VersionTLS11:
		return "11"
	case tls.VersionTLS10:
		return "10"
	case tls.VersionSSL30: //nolint:staticcheck // SSLv3 kept for fingerprinting historical clients
		return "s3"
	default:
		return "00"
	}
}

// sniIndicator returns "d" if SNI is present (domain), "i" if connecting to IP.
func sniIndicator(serverName string, remoteAddr string) string {
	if serverName != "" {
		// Check if serverName is an IP address
		if net.ParseIP(serverName) != nil {
			return "i"
		}
		return "d"
	}

	// No SNI, check if remoteAddr indicates IP-only connection
	if remoteAddr != "" {
		host, _, _ := net.SplitHostPort(remoteAddr)
		if net.ParseIP(host) != nil {
			return "i"
		}
	}

	return "i" // default to IP if uncertain
}

// estimateExtensionCount estimates the extension count based on available data.
// This is a workaround since Go doesn't expose raw extensions.
func estimateExtensionCount(hello *TLSClientHello) string {
	count := 0

	// Common extensions we know about:
	// - supported_versions (if Version is set)
	// - supported_groups (if Curves exist)
	// - ec_point_formats (if PointFormats exist)
	// - server_name (usually present)
	// - Others we can't detect: session_ticket, status_request, etc.

	if hello.Version > 0 {
		count++ // supported_versions
	}
	if len(hello.Curves) > 0 {
		count++ // supported_groups
	}
	if len(hello.PointFormats) > 0 {
		count++ // ec_point_formats
	}

	// Assume a baseline of common extensions (SNI, status_request, etc.)
	count += 5 // typical ClientHello has 5-15 extensions

	return fmt.Sprintf("%02d", count)
}

// hashCiphers computes a 12-character SHA256 hash of sorted cipher suites.
// GREASE values are filtered out per JA4 spec.
func hashCiphers(ciphers []uint16) string {
	if len(ciphers) == 0 {
		return "000000000000"
	}

	// Filter GREASE values (RFC 8701)
	filtered := make([]uint16, 0, len(ciphers))
	for _, c := range ciphers {
		if !isGREASE(c) {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		return "000000000000"
	}

	// Sort in ascending order
	slices.Sort(filtered)

	// Build comma-delimited hex string
	parts := make([]string, len(filtered))
	for i, c := range filtered {
		parts[i] = fmt.Sprintf("%04x", c)
	}
	raw := strings.Join(parts, ",")

	// Compute SHA256 and truncate to 12 chars
	hash := crypto.SHA256String(raw)
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

// hashExtensions computes a 12-character SHA256 hash of sorted extension data.
// Since Go doesn't expose raw extensions, we use supported curves as a proxy.
func hashExtensions(curves []tls.CurveID) string {
	if len(curves) == 0 {
		return "000000000000"
	}

	// Convert to uint16 for sorting
	vals := make([]uint16, len(curves))
	for i, c := range curves {
		vals[i] = uint16(c)
	}

	// Filter GREASE
	filtered := make([]uint16, 0, len(vals))
	for _, v := range vals {
		if !isGREASE(v) {
			filtered = append(filtered, v)
		}
	}

	if len(filtered) == 0 {
		return "000000000000"
	}

	// Sort in ascending order
	slices.Sort(filtered)

	// Build comma-delimited hex string
	parts := make([]string, len(filtered))
	for i, v := range filtered {
		parts[i] = fmt.Sprintf("%04x", v)
	}
	raw := strings.Join(parts, ",")

	// Compute SHA256 and truncate to 12 chars
	hash := crypto.SHA256String(raw)
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

// isGREASE checks if a value is a GREASE value per RFC 8701.
// GREASE values follow the pattern: 0x?A?A where ? is any hex digit.
// Examples: 0x0A0A, 0x1A1A, 0x2A2A, ..., 0xFAFA
func isGREASE(val uint16) bool {
	// Both bytes must be identical and end with 0xA
	highByte := (val >> 8) & 0xFF
	lowByte := val & 0xFF
	return highByte == lowByte && (lowByte&0x0F) == 0x0A
}
