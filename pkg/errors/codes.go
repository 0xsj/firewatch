package errors

// Code is a machine-readable identifier for a specific error condition.
// Codes are scoped by domain and follow the format "domain_detail".
// Unlike Kind (broad category), Code pinpoints exactly what went wrong.
type Code string

// Config errors.
const (
	CodeConfigInvalid Code = "config_invalid" // Config file is malformed
	CodeConfigMissing Code = "config_missing" // Required config field absent
)

// Server errors.
const (
	CodeServerBind     Code = "server_bind"     // Failed to bind address
	CodeServerShutdown Code = "server_shutdown" // Error during graceful shutdown
	CodeServerTLS      Code = "server_tls"      // TLS configuration error
)

// Storage errors.
const (
	CodeStorageConnect Code = "storage_connect" // Failed to connect to database
	CodeStorageQuery   Code = "storage_query"   // Query execution failed
	CodeStorageMigrate Code = "storage_migrate" // Schema migration failed
)

// Handler errors.
const (
	CodeHandlerInit  Code = "handler_init"  // Module failed to initialize
	CodeHandlerPanic Code = "handler_panic" // Handler panicked
)

// Fingerprint errors.
const (
	CodeFingerprintJA3 Code = "fingerprint_ja3" // JA3 extraction failed
	CodeFingerprintGeo Code = "fingerprint_geo" // GeoIP lookup failed
)

// Alert errors.
const (
	CodeAlertSend      Code = "alert_send"       // Failed to send alert
	CodeAlertRateLimit Code = "alert_rate_limit" // Alert rate limit hit
)

// Intel errors.
const (
	CodeIntelEnrich Code = "intel_enrich" // Enrichment source failed
	CodeIntelExport Code = "intel_export" // Export generation failed
)
