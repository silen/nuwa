package middleware

import (
	"net/http"
	"testing"
)

func TestSanitizeHeadersRedactsSensitiveValues(t *testing.T) {
	t.Parallel()

	headers := http.Header{
		"Authorization": []string{"Bearer secret"},
		"Token":         []string{"abc"},
		"X-Request-Id":  []string{"req-1"},
	}

	got := sanitizeHeaders(headers)

	if got.Get("Authorization") != "[REDACTED]" {
		t.Fatalf("expected authorization to be redacted, got %q", got.Get("Authorization"))
	}
	if got.Get("Token") != "[REDACTED]" {
		t.Fatalf("expected token to be redacted, got %q", got.Get("Token"))
	}
	if got.Get("X-Request-Id") != "req-1" {
		t.Fatalf("expected non-sensitive headers to be preserved, got %q", got.Get("X-Request-Id"))
	}
}
