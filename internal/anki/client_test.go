package anki

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestStatusErrorIncludesHTTPBody(t *testing.T) {
	err := statusError("ankiconnect deckNames", &http.Response{
		StatusCode: http.StatusBadGateway,
		Body:       io.NopCloser(strings.NewReader("bad gateway")),
	})

	message := err.Error()

	if !strings.Contains(message, "HTTP 502") {
		t.Fatalf("expected HTTP status in error, got %q", message)
	}

	if !strings.Contains(message, "bad gateway") {
		t.Fatalf("expected body in error, got %q", message)
	}
}
