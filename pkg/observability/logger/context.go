package logger

import (
	"context"

	pkgcontext "github.com/0xsj/hexagonal-go/pkg/context"
)

// ExtractContextFields extracts logging fields from context.
// Automatically includes: tenant_id, user_id, request_id, correlation_id, session_id
func ExtractContextFields(ctx context.Context) []Field {
	fields := make([]Field, 0, 5)

	if tenantID := pkgcontext.TenantID(ctx); tenantID != "" {
		fields = append(fields, String(FieldTenantID, tenantID))
	}

	if userID := pkgcontext.UserID(ctx); userID != "" {
		fields = append(fields, String(FieldUserID, userID))
	}

	if requestID := pkgcontext.RequestID(ctx); requestID != "" {
		fields = append(fields, String(FieldRequestID, requestID))
	}

	if correlationID := pkgcontext.CorrelationID(ctx); correlationID != "" {
		fields = append(fields, String(FieldCorrelationID, correlationID))
	}

	if sessionID := pkgcontext.SessionID(ctx); sessionID != "" {
		fields = append(fields, String(FieldSessionID, sessionID))
	}

	return fields
}
