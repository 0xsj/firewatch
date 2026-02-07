package middleware

import (
	"log/slog"
	"net/http"

	"github.com/0xsj/firewatch/internal/geoip"
	"github.com/0xsj/firewatch/pkg/httputil"
)

// GeoIP performs geolocation lookups on the client IP and stores
// the result in the request context for downstream handlers.
func GeoIP(reader *geoip.Reader, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := httputil.ClientIP(r)
			info, err := reader.Lookup(ip)
			if err != nil {
				logger.Debug("geoip lookup failed",
					"ip", ip,
					"error", err,
					"request_id", RequestID(r.Context()),
				)
			}

			if info != nil {
				ctx := geoip.WithGeoIP(r.Context(), info)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
