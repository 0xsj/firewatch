package wordpress

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigXMLRPCProbe   = "wp-xmlrpc-001"
	sigXMLRPCPayload = "wp-xmlrpc-002"
)

func (wp *WordPress) handleXMLRPC(w http.ResponseWriter, r *http.Request) {
	sigs := []string{sigXMLRPCProbe}
	severity := "high"

	if r.Method == http.MethodPost {
		body, _ := httputil.ReadBody(r, 0)
		if len(body) > 0 {
			sigs = append(sigs, sigXMLRPCPayload)
			severity = "critical"
		}

		wp.logger.Info("xmlrpc probe",
			"body_size", len(body),
			"ip", httputil.ClientIP(r),
		)
	}

	handlers.RecordEvent(wp.store, wp.logger, r, moduleName, severity, sigs)

	// Return a valid XML-RPC response
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("X-Powered-By", "PHP/8.1.0")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<methodResponse>
  <params>
    <param>
      <value>
        <array><data>
          <value><string>system.multicall</string></value>
          <value><string>system.listMethods</string></value>
          <value><string>wp.getUsersBlogs</string></value>
        </data></array>
      </value>
    </param>
  </params>
</methodResponse>`))
}
