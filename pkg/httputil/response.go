package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/0xsj/firewatch/pkg/errors"
)

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes an error response, mapping the error's Kind to
// an HTTP status code. The response body is JSON.
func Error(w http.ResponseWriter, err error) {
	status := errors.HTTPStatus(err)
	JSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

// HTML writes an HTML response with the given status code.
func HTML(w http.ResponseWriter, status int, body string) {
	w.Header().Set(HeaderContentType, ContentTypeHTML)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// Text writes a plain text response with the given status code.
func Text(w http.ResponseWriter, status int, body string) {
	w.Header().Set(HeaderContentType, ContentTypeText)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// Raw writes raw bytes with the given content type and status code.
func Raw(w http.ResponseWriter, status int, contentType string, body []byte) {
	w.Header().Set(HeaderContentType, contentType)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
