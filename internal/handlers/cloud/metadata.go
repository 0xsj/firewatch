package cloud

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/deception"
	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/internal/middleware"
	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/httputil"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

const (
	sigMetadataProbe = "cloud-metadata-001"
	sigIAMProbe      = "cloud-iam-001"
	sigIMDSv2Probe   = "cloud-imdsv2-001"
)

func (c *Cloud) handleMetadata(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("cloud metadata probe",
		"path", r.URL.Path,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "critical", []string{sigMetadataProbe})

	if c.deception.Breadcrumbs {
		deception.BreadcrumbHeaders(w, moduleName, c.breadcrumbCfg())
	}

	// Return fake AWS-style metadata
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ami-id\nami-launch-index\nami-manifest-path\nhostname\ninstance-action\ninstance-id\ninstance-life-cycle\ninstance-type\nlocal-hostname\nlocal-ipv4\nmac\nnetwork\nplacement\nprofile\npublic-hostname\npublic-ipv4\npublic-keys\nsecurity-groups\nservices\n"))
}

func (c *Cloud) handleIAM(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("IAM credential probe",
		"path", r.URL.Path,
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "critical", []string{sigIAMProbe})

	var accessKey, secretKey, sessionToken string

	if c.deception.HoneyTokens {
		accessKey = deception.GenerateAWSAccessKey()
		secretKey = deception.GenerateAWSSecretKey()
		sessionToken = deception.GenerateSessionToken()

		reqID := middleware.RequestID(r.Context())
		now := timeutil.FormatRFC3339(timeutil.NowUTC())
		ip := httputil.ClientIP(r)

		c.saveToken(r, "aws_access_key", accessKey, now, ip, r.URL.Path, reqID)
		c.saveToken(r, "aws_secret_key", secretKey, now, ip, r.URL.Path, reqID)
		c.saveToken(r, "session_token", sessionToken, now, ip, r.URL.Path, reqID)
	} else {
		accessKey = "AKIAIOSFODNN7EXAMPLE"
		secretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		sessionToken = "FwoGZXIvYXdzEBAaDHoney+Token+Do+Not+Use"
	}

	httputil.JSON(w, http.StatusOK, map[string]any{
		"Code":            "Success",
		"LastUpdated":     "2024-01-15T12:00:00Z",
		"Type":            "AWS-HMAC",
		"AccessKeyId":     accessKey,
		"SecretAccessKey": secretKey,
		"Token":           sessionToken,
		"Expiration":      "2099-12-31T23:59:59Z",
	})
}

func (c *Cloud) handleIMDSv2(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("IMDSv2 token request",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "critical", []string{sigIMDSv2Probe})

	var token string
	if c.deception.HoneyTokens {
		token = deception.GenerateIMDSToken()

		reqID := middleware.RequestID(r.Context())
		now := timeutil.FormatRFC3339(timeutil.NowUTC())
		ip := httputil.ClientIP(r)
		c.saveToken(r, "imds_token", token, now, ip, r.URL.Path, reqID)
	} else {
		token = "AQAAANjCpMCZjg_fake_imds_token_do_not_use"
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Aws-Ec2-Metadata-Token-Ttl-Seconds", "21600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(token))
}

func (c *Cloud) saveToken(r *http.Request, kind, value, issuedAt, sourceIP, path, requestID string) {
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
	if err := c.store.SaveHoneyToken(r.Context(), token); err != nil {
		c.logger.Error("failed to save honey token",
			"error", err,
			"kind", kind,
		)
	}
}

func (c *Cloud) breadcrumbCfg() deception.BreadcrumbConfig {
	return deception.BreadcrumbConfig{
		Domain:         "",
		EnabledModules: []string{"admin", "api", "cloud", "cve", "exposure", "nextjs", "wordpress"},
	}
}
