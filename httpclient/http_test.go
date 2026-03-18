package httpclient

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestSendReturnsContextErrorWhenCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewHTTP(ctx)
	_, err := client.Get("http://example.com", map[string]any{"q": "x"}, nil)
	if err == nil {
		t.Fatalf("expected canceled context error")
	}
}

func TestDoRequestWithTimeoutReturnsExpiredContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := doRequestWithTimeout(ctx, 10*time.Millisecond, req, resp); err == nil {
		t.Fatalf("expected deadline error")
	}
}

func TestDefaultContentType(t *testing.T) {
	t.Parallel()

	if got := defaultContentType("POST"); got != "application/x-www-form-urlencoded" {
		t.Fatalf("unexpected post content type: %s", got)
	}
	if got := defaultContentType("JsonBody"); got != "application/json" {
		t.Fatalf("unexpected json content type: %s", got)
	}
}

func TestAppendQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		q    string
		want string
	}{
		{name: "no existing query", url: "http://example.com/path", q: "a=1", want: "http://example.com/path?a=1"},
		{name: "existing query", url: "http://example.com/path?b=2", q: "a=1", want: "http://example.com/path?b=2&a=1"},
		{name: "trailing question", url: "http://example.com/path?", q: "a=1", want: "http://example.com/path?a=1"},
		{name: "empty query", url: "http://example.com/path", q: "", want: "http://example.com/path"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := appendQuery(tt.url, tt.q); got != tt.want {
				t.Fatalf("appendQuery() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUploadFileReturnsContextErrorWhenCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewHTTP(ctx)
	_, err := client.UploadFile("http://example.com/upload", bytes.NewBufferString("demo"), "demo.txt")
	if err == nil {
		t.Fatalf("expected canceled context error")
	}
}

func TestBuildUploadRequestSpecReturnsErrorWhenReaderNil(t *testing.T) {
	t.Parallel()

	client := NewHTTP(context.Background())
	_, err := client.buildUploadRequestSpec("http://example.com/upload", nil, "demo.txt")
	if err == nil {
		t.Fatalf("expected nil reader error")
	}
}
