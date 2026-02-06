package fingerprint

import (
	"net/http"
	"sort"
	"strings"

	"github.com/0xsj/firewatch/pkg/crypto"
)

// HeaderFingerprint captures header-based signals from a request.
type HeaderFingerprint struct {
	// OrderHash is a hash of the header key order. Different HTTP
	// clients send headers in different orders, providing a signal
	// even when values are identical.
	OrderHash string

	// Keys is the ordered list of header keys as received.
	// Note: Go's http.Header is a map, so true wire order is not
	// preserved. This captures the canonical key set.
	Keys []string

	// UserAgent is the raw User-Agent string.
	UserAgent string

	// Anomalies lists any detected header irregularities.
	Anomalies []string

	// KnownClient is set if the User-Agent matches a known scanner
	// or HTTP library pattern.
	KnownClient string
}

// AnalyzeHeaders extracts a header fingerprint from the request.
func AnalyzeHeaders(r *http.Request) HeaderFingerprint {
	fp := HeaderFingerprint{
		UserAgent: r.UserAgent(),
	}

	// Collect and sort header keys deterministically.
	// Since Go's map iteration is random, sorting gives us a stable
	// fingerprint across requests from the same client.
	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fp.Keys = keys
	fp.OrderHash = crypto.SHA256String(strings.Join(keys, ","))

	// Detect header anomalies
	fp.Anomalies = detectAnomalies(r)

	// Match known clients
	fp.KnownClient = matchKnownClient(r.UserAgent())

	return fp
}

// detectAnomalies checks for header patterns that deviate from
// what a normal browser would send.
func detectAnomalies(r *http.Request) []string {
	var anomalies []string

	// Missing Accept header — most browsers always send this
	if r.Header.Get("Accept") == "" {
		anomalies = append(anomalies, "missing_accept")
	}

	// Missing Accept-Language — browsers always send this
	if r.Header.Get("Accept-Language") == "" {
		anomalies = append(anomalies, "missing_accept_language")
	}

	// Missing Accept-Encoding — very unusual for real browsers
	if r.Header.Get("Accept-Encoding") == "" {
		anomalies = append(anomalies, "missing_accept_encoding")
	}

	// Empty or missing User-Agent
	if r.UserAgent() == "" {
		anomalies = append(anomalies, "missing_user_agent")
	}

	// Connection: close without keep-alive (unusual for HTTP/1.1)
	if r.Header.Get("Connection") == "close" && r.ProtoMajor == 1 && r.ProtoMinor == 1 {
		anomalies = append(anomalies, "connection_close_http11")
	}

	return anomalies
}

// Known scanner and HTTP library patterns.
var knownClients = []struct {
	pattern string
	name    string
}{
	{"python-requests", "python-requests"},
	{"Python-urllib", "python-urllib"},
	{"Go-http-client", "go-http-client"},
	{"curl/", "curl"},
	{"wget/", "wget"},
	{"Nuclei", "nuclei"},
	{"sqlmap", "sqlmap"},
	{"nikto", "nikto"},
	{"Nmap", "nmap"},
	{"masscan", "masscan"},
	{"zgrab", "zgrab"},
	{"httpx", "httpx"},
	{"Scrapy", "scrapy"},
	{"axios/", "axios"},
	{"node-fetch", "node-fetch"},
	{"Java/", "java"},
	{"Apache-HttpClient", "apache-httpclient"},
	{"okhttp", "okhttp"},
	{"libwww-perl", "libwww-perl"},
	{"Ruby", "ruby"},
	{"PHP/", "php"},
}

// matchKnownClient checks the User-Agent against known scanner
// and HTTP library patterns. Returns the client name or empty string.
func matchKnownClient(ua string) string {
	lower := strings.ToLower(ua)
	for _, kc := range knownClients {
		if strings.Contains(lower, strings.ToLower(kc.pattern)) {
			return kc.name
		}
	}
	return ""
}
