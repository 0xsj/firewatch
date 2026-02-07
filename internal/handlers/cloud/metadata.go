package cloud

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
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

	// Return fake IAM credentials — these are tracked honey tokens
	httputil.JSON(w, http.StatusOK, map[string]any{
		"Code":            "Success",
		"LastUpdated":     "2024-01-15T12:00:00Z",
		"Type":            "AWS-HMAC",
		"AccessKeyId":     "AKIAIOSFODNN7EXAMPLE",
		"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"Token":           "FwoGZXIvYXdzEBAaDHoney+Token+Do+Not+Use",
		"Expiration":      "2099-12-31T23:59:59Z",
	})
}

func (c *Cloud) handleIMDSv2(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("IMDSv2 token request",
		"ip", httputil.ClientIP(r),
	)

	handlers.RecordEvent(c.store, c.logger, r, moduleName, "critical", []string{sigIMDSv2Probe})

	// Return a fake IMDSv2 token
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Aws-Ec2-Metadata-Token-Ttl-Seconds", "21600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("AQAAANjCpMCZjg_fake_imds_token_do_not_use"))
}
