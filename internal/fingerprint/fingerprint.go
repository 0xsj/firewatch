package fingerprint

import (
	"context"
	"net/http"
)

type contextKey string

const fingerprintKey contextKey = "fingerprint"

// Result holds all fingerprinting data collected for a request.
type Result struct {
	// JA3 fields (populated when TLS is active)
	JA3Raw  string `json:"ja3_raw,omitempty"`
	JA3Hash string `json:"ja3_hash,omitempty"`

	// Header analysis
	HeaderOrderHash string   `json:"header_order_hash"`
	HeaderKeys      []string `json:"header_keys"`
	UserAgent       string   `json:"user_agent"`
	KnownClient     string   `json:"known_client,omitempty"`
	Anomalies       []string `json:"anomalies,omitempty"`
}

// Engine orchestrates all fingerprinting techniques against
// an incoming request.
type Engine struct {
	ja3Store *JA3Store
}

// NewEngine creates a fingerprint engine. If ja3Store is nil,
// JA3 fingerprinting is skipped (non-TLS mode).
func NewEngine(ja3Store *JA3Store) *Engine {
	return &Engine{ja3Store: ja3Store}
}

// Analyze runs all fingerprinting techniques on the request
// and returns a combined Result.
func (e *Engine) Analyze(r *http.Request) Result {
	result := Result{}

	// JA3 — only available when TLS is active and the store has data
	if e.ja3Store != nil {
		if hello := e.ja3Store.Take(r.RemoteAddr); hello != nil {
			result.JA3Raw, result.JA3Hash = JA3(hello)
		}
	}

	// Header analysis
	hfp := AnalyzeHeaders(r)
	result.HeaderOrderHash = hfp.OrderHash
	result.HeaderKeys = hfp.Keys
	result.UserAgent = hfp.UserAgent
	result.KnownClient = hfp.KnownClient
	result.Anomalies = hfp.Anomalies

	return result
}

// WithResult stores a fingerprint Result in the request context.
func WithResult(ctx context.Context, result Result) context.Context {
	return context.WithValue(ctx, fingerprintKey, result)
}

// GetResult extracts the fingerprint Result from the context.
// Returns an empty Result if none is present.
func GetResult(ctx context.Context) Result {
	result, _ := ctx.Value(fingerprintKey).(Result)
	return result
}
