package exposure

import (
	"fmt"
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/middleware"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/httputil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

const sigEnvProbe = "exposure-env-001"

func (e *Exposure) handleEnv(w http.ResponseWriter, r *http.Request) {
	e.logger.Info("env file probe",
		"path", r.URL.Path,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(e.store, e.logger, r, moduleName, "high", []string{sigEnvProbe})

	var content string

	if e.cfg.FakeEnv != "" {
		content = e.cfg.FakeEnv
	} else if e.deception.HoneyTokens {
		content = e.generateEnv(r)
	} else {
		content = deception.ExposedEnvFile()
	}

	if e.deception.Breadcrumbs && e.cfg.FakeEnv == "" {
		cfg := deception.BreadcrumbConfig{
			Domain:         "",
			EnabledModules: []string{"admin", "api", "cloud", "cve", "exposure", "nextjs", "wordpress"},
		}
		content = deception.InjectEnv(content, moduleName, cfg)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(content))
}

func (e *Exposure) generateEnv(r *http.Request) string {
	dbPass := deception.GenerateDBPassword()
	accessKey := deception.GenerateAWSAccessKey()
	secretKey := deception.GenerateAWSSecretKey()
	apiKey := deception.GenerateAPIKey("sk_live")

	reqID := middleware.RequestID(r.Context())
	now := timeutil.FormatRFC3339(timeutil.NowUTC())
	ip := httputil.ClientIP(r)

	e.saveToken(r, "db_password", dbPass, now, ip, r.URL.Path, reqID)
	e.saveToken(r, "aws_access_key", accessKey, now, ip, r.URL.Path, reqID)
	e.saveToken(r, "aws_secret_key", secretKey, now, ip, r.URL.Path, reqID)
	e.saveToken(r, "api_key", apiKey, now, ip, r.URL.Path, reqID)

	return fmt.Sprintf(`APP_NAME=MyApp
APP_ENV=production
APP_DEBUG=false
APP_URL=https://app.example.com

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=myapp_prod
DB_USERNAME=myapp_user
DB_PASSWORD=%s

REDIS_HOST=127.0.0.1
REDIS_PASSWORD=redis_secret_pass

MAIL_MAILER=smtp
MAIL_HOST=smtp.mailgun.org
MAIL_PORT=587
MAIL_USERNAME=postmaster@mg.example.com
MAIL_PASSWORD=mail_key_abc123

AWS_ACCESS_KEY_ID=%s
AWS_SECRET_ACCESS_KEY=%s
AWS_DEFAULT_REGION=us-east-1
AWS_BUCKET=myapp-uploads

STRIPE_KEY=%s
STRIPE_SECRET=whsec_fake_webhook_secret
`, dbPass, accessKey, secretKey, apiKey)
}

func (e *Exposure) saveToken(r *http.Request, kind, value, issuedAt, sourceIP, path, requestID string) {
	token := &models.HoneyToken{
		ID:        crypto.UUID4(),
		Kind:      kind,
		Value:     value,
		IssuedAt:  issuedAt,
		SourceIP:  sourceIP,
		Module:    moduleName,
		Path:      path,
		RequestID: requestID,
	}
	if err := e.store.SaveHoneyToken(r.Context(), token); err != nil {
		e.logger.Error("failed to save honey token",
			"error", err,
			"kind", kind,
		)
	}
}
