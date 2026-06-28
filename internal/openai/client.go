// Package openai builds prompts and calls OpenAI chat completions to generate
// structured Anki card data.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/punnch/ankiwords/internal/model"
)

// Client calls OpenAI's chat completions endpoint and decodes generated cards.
type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error"`
}

// NewClient creates an OpenAI API client.
func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateCards asks OpenAI to produce one structured Anki card per target
// word, then decodes the JSON response into GeneratedCard values.
func (c *Client) GenerateCards(
	ctx context.Context,
	template string,
	words []string,
	preferredModel string,
	modelFields []string,
	formatter model.SentenceFormatter,
) ([]model.GeneratedCard, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("openai api key is missing")
	}

	if len(words) == 0 {
		return nil, fmt.Errorf("no words supplied")
	}

	systemPrompt := buildSystemPrompt(formatter)
	userPrompt := buildUserPrompt(template, words, preferredModel, modelFields, formatter)

	reqBody, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, openAIStatusError(resp)
	}

	var envelope chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}

	if len(envelope.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	content := strings.TrimSpace(envelope.Choices[0].Message.Content)

	var cards []model.GeneratedCard
	if err := json.Unmarshal([]byte(content), &cards); err != nil {
		return nil, fmt.Errorf("invalid openai json: %w", err)
	}

	return cards, nil
}

// openAIStatusError extracts useful diagnostics from non-2xx OpenAI responses.
func openAIStatusError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("openai returned HTTP %d: could not read error body: %w", resp.StatusCode, err)
	}

	var envelope errorResponse
	if err := json.Unmarshal(body, &envelope); err == nil && strings.TrimSpace(envelope.Error.Message) != "" {
		message := strings.TrimSpace(envelope.Error.Message)
		if envelope.Error.Type != "" {
			return fmt.Errorf("openai returned HTTP %d: %s (%s)", resp.StatusCode, message, envelope.Error.Type)
		}

		return fmt.Errorf("openai returned HTTP %d: %s", resp.StatusCode, message)
	}

	bodyText := strings.TrimSpace(string(body))
	if bodyText == "" {
		return fmt.Errorf("openai returned HTTP %d with empty body", resp.StatusCode)
	}

	return fmt.Errorf("openai returned HTTP %d: %s", resp.StatusCode, bodyText)
}

// buildSystemPrompt defines global generation rules that should not vary by
// user template.
func buildSystemPrompt(formatter model.SentenceFormatter) string {
	return fmt.Sprintf(
		`You generate structured data for Anki cards.
Return JSON only.
No prose, no markdown, no code fences.
Return an array of objects with this shape:
[
  {
    "word": "target word",
    "fields": {
      "Exact Anki field name": "field value"
    }
  }
]
The target sentence highlight must use this exact wrapper:
%s`, formatter.Highlight("WORD"))
}

// buildUserPrompt combines user preferences, target words, and the exact
// sentence highlighting contract.
func buildUserPrompt(
	template string,
	words []string,
	preferredModel string,
	modelFields []string,
	formatter model.SentenceFormatter,
) string {
	var b strings.Builder

	b.WriteString("Generation template:\n")
	b.WriteString(template)
	b.WriteString("\n\nTarget model:\n")
	b.WriteString(preferredModel)
	b.WriteString("\n\nTarget Anki fields, in order:\n")
	for i, field := range modelFields {
		fmt.Fprintf(&b, "%d. %s\n", i+1, field)
	}
	b.WriteString("\n\nTarget words in order:\n")

	for i, word := range words {
		fmt.Fprintf(&b, "%d. %s\n", i+1, word)
	}

	b.WriteString("\nReturn a JSON array with one object per target word.\n")
	b.WriteString("Each object must have a word string matching the target word and a fields object.\n")
	b.WriteString("The fields object must use exactly the target Anki field names listed above. Do not add, omit, or rename fields.\n")
	b.WriteString("Decide the best content for each field from its name and the target model. For example, Basic Front/Back cards should use Front and Back; vocabulary models can use fields such as Word, Sentence, Translation, Definition, or Transcription when those fields exist.\n")
	b.WriteString("If a Sentence field exists, it must contain the highlighted target word using this exact style:\n")
	b.WriteString(formatter.Highlight("WORD"))

	return b.String()
}
