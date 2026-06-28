package openai

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/punnch/ankiwords/internal/model"
)

func TestOpenAIStatusErrorIncludesOpenAIErrorMessage(t *testing.T) {
	err := openAIStatusError(&http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"Incorrect API key provided","type":"invalid_request_error"}}`)),
	})

	message := err.Error()
	if !strings.Contains(message, "HTTP 401") {
		t.Fatalf("expected status in error, got %q", message)
	}
	if !strings.Contains(message, "Incorrect API key provided") {
		t.Fatalf("expected OpenAI message in error, got %q", message)
	}
}

func TestBuildUserPromptIncludesTargetModelFields(t *testing.T) {
	prompt := buildUserPrompt(
		"Create useful vocabulary cards.",
		[]string{"collapse"},
		"Basic",
		[]string{"Front", "Back"},
		model.SentenceFormatter{Color: "#00557f"},
	)

	for _, want := range []string{"Target Anki fields", "Front", "Back", "fields object", "Basic Front/Back"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}
