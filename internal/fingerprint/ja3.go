package fingerprint

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"github.com/0xsj/firewatch/pkg/crypto"
)

// TLSClientHello holds the parameters captured from a TLS ClientHello
// message during the handshake. These are used to compute JA3.
type TLSClientHello struct {
	Version      uint16
	CipherSuites []uint16
	Curves       []tls.CurveID
	PointFormats []uint8
}

// JA3 computes the JA3 fingerprint string and its MD5 hash from
// the captured TLS ClientHello parameters.
//
// JA3 format: TLSVersion,Ciphers,Extensions,Curves,PointFormats
// Each section is a dash-separated list of decimal values.
//
// Note: Go's tls.ClientHelloInfo does not expose the raw extensions
// list, so the extensions section is left empty. The remaining fields
// (version, ciphers, curves, point formats) still provide a strong
// fingerprinting signal.
func JA3(hello *TLSClientHello) (raw string, hash string) {
	if hello == nil {
		return "", ""
	}

	version := fmt.Sprintf("%d", hello.Version)
	ciphers := uint16sToString(hello.CipherSuites)
	extensions := "" // not available from Go's TLS API
	curves := curveIDsToString(hello.Curves)
	points := uint8sToString(hello.PointFormats)

	raw = strings.Join([]string{version, ciphers, extensions, curves, points}, ",")
	hash = crypto.MD5String(raw)
	return raw, hash
}

// JA3FromClientHello extracts TLS parameters from Go's ClientHelloInfo
// and computes the JA3 fingerprint.
func JA3FromClientHello(info *tls.ClientHelloInfo) (raw string, hash string) {
	if info == nil {
		return "", ""
	}

	hello := &TLSClientHello{
		CipherSuites: info.CipherSuites,
		Curves:       info.SupportedCurves,
		PointFormats: info.SupportedPoints,
	}

	// Get the highest supported version.
	if len(info.SupportedVersions) > 0 {
		hello.Version = info.SupportedVersions[0]
	}

	return JA3(hello)
}

func uint16sToString(vals []uint16) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, "-")
}

func curveIDsToString(curves []tls.CurveID) string {
	parts := make([]string, len(curves))
	for i, c := range curves {
		parts[i] = fmt.Sprintf("%d", c)
	}
	return strings.Join(parts, "-")
}

func uint8sToString(vals []uint8) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, "-")
}

// JA3Store is a thread-safe store for TLS ClientHello data captured
// during the TLS handshake. The fingerprint middleware reads from it
// after the HTTP request is received.
//
// Entries are keyed by remote address (ip:port). Each entry is
// consumed once (read-and-delete) to prevent memory growth.
type JA3Store struct {
	mu     sync.RWMutex
	hellos map[string]*TLSClientHello
}

// NewJA3Store creates an empty JA3 store.
func NewJA3Store() *JA3Store {
	return &JA3Store{
		hellos: make(map[string]*TLSClientHello),
	}
}

// Put stores a ClientHello keyed by remote address.
func (s *JA3Store) Put(remoteAddr string, hello *TLSClientHello) {
	s.mu.Lock()
	s.hellos[remoteAddr] = hello
	s.mu.Unlock()
}

// Take retrieves and removes a ClientHello by remote address.
// Returns nil if no entry exists.
func (s *JA3Store) Take(remoteAddr string) *TLSClientHello {
	s.mu.Lock()
	hello, ok := s.hellos[remoteAddr]
	if ok {
		delete(s.hellos, remoteAddr)
	}
	s.mu.Unlock()
	return hello
}

// TLSConfigCallback returns a GetConfigForClient function that
// captures ClientHello data into the store. Wire this into the
// server's tls.Config.
func (s *JA3Store) TLSConfigCallback() func(*tls.ClientHelloInfo) (*tls.Config, error) {
	return func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		hello := &TLSClientHello{
			CipherSuites: info.CipherSuites,
			Curves:       info.SupportedCurves,
			PointFormats: info.SupportedPoints,
		}
		if len(info.SupportedVersions) > 0 {
			hello.Version = info.SupportedVersions[0]
		}
		s.Put(info.Conn.RemoteAddr().String(), hello)
		return nil, nil // nil config means use the default
	}
}
