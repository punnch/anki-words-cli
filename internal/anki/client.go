// Package anki contains a small typed client for the AnkiConnect API.
package anki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client wraps the AnkiConnect HTTP API used by the app.
type Client struct {
	baseURL string
	http    *http.Client
}

// Note is the payload shape accepted by AnkiConnect's addNote action.
type Note struct {
	DeckName  string            `json:"deckName"`
	ModelName string            `json:"modelName"`
	Fields    map[string]string `json:"fields"`
	Tags      []string          `json:"tags,omitempty"`
	Options   map[string]any    `json:"options,omitempty"`
	Audio     []any             `json:"audio,omitempty"`
	Pictures  []any             `json:"picture,omitempty"`
}

// ankiRequest is the common AnkiConnect request envelope.
type ankiRequest struct {
	Action  string         `json:"action"`
	Version int            `json:"version"`
	Params  map[string]any `json:"params,omitempty"`
}

// ankiResponse is the common AnkiConnect response envelope.
type ankiResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *string         `json:"error"`
}

// NewClient creates an AnkiConnect client for the configured base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DeckNames returns all decks available in the connected Anki profile.
func (c *Client) DeckNames(ctx context.Context) ([]string, error) {
	var out []string
	if err := c.call(ctx, "deckNames", nil, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// ModelNames returns all note model names available in Anki.
func (c *Client) ModelNames(ctx context.Context) ([]string, error) {
	var out []string
	if err := c.call(ctx, "modelNames", nil, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// ModelFieldNames returns the field names defined by a note model.
func (c *Client) ModelFieldNames(ctx context.Context, model string) ([]string, error) {
	var out []string
	if err := c.call(ctx, "modelFieldNames", map[string]any{"modelName": model}, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// AddNote creates one Anki note and returns the note ID assigned by Anki.
func (c *Client) AddNote(ctx context.Context, note Note) (int64, error) {
	var out int64
	if err := c.call(ctx, "addNote", map[string]any{"note": note}, &out); err != nil {
		return 0, err
	}

	return out, nil
}

// call sends a single AnkiConnect action and unmarshals its result payload into
// out when a result is expected.
func (c *Client) call(ctx context.Context, action string, params map[string]any, out any) error {
	reqBody, err := json.Marshal(ankiRequest{Action: action, Version: 6, Params: params})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return statusError("ankiconnect "+action, resp)
	}

	var envelope ankiResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return err
	}

	if envelope.Error != nil {
		return fmt.Errorf("ankiconnect %s: %s", action, *envelope.Error)
	}

	if out == nil {
		return nil
	}

	return json.Unmarshal(envelope.Result, out)
}

// statusError includes upstream HTTP status and response body in transport
// errors.
func statusError(prefix string, resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("%s returned HTTP %d: could not read error body: %w", prefix, resp.StatusCode, err)
	}

	bodyText := strings.TrimSpace(string(body))
	if bodyText == "" {
		return fmt.Errorf("%s returned HTTP %d with empty body", prefix, resp.StatusCode)
	}

	return fmt.Errorf("%s returned HTTP %d: %s", prefix, resp.StatusCode, bodyText)
}
