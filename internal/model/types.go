// Package model defines shared data structures used across the TUI workflow,
// generation, validation, and Anki note creation.
package model

import (
	"fmt"
	"html"
	"strings"
)

// UserPrefs stores the local defaults that affect card creation.
type UserPrefs struct {
	UserID             string `json:"user_id"`
	ActiveDeck         string `json:"active_deck"`
	GenerationTemplate string `json:"generation_template"`
	PreferredModel     string `json:"preferred_model"`
}

// GeneratedCard is the JSON object shape expected from the OpenAI response.
type GeneratedCard struct {
	Word   string            `json:"word"`
	Fields map[string]string `json:"fields"`
}

// SentenceFormatter owns the exact HTML used to highlight vocabulary words.
type SentenceFormatter struct {
	Color string
}

// Highlight returns the canonical HTML wrapper for a target word.
func (f SentenceFormatter) Highlight(word string) string {
	color := f.Color
	if color == "" {
		color = "#00557f"
	}

	return fmt.Sprintf(`<span style="color:%s;"><b>%s</b></span>`, html.EscapeString(color), html.EscapeString(word))
}

// Apply replaces the first case-insensitive occurrence of word with its
// highlighted HTML form.
func (f SentenceFormatter) Apply(sentence, word string) string {
	if sentence == "" || word == "" {
		return sentence
	}

	highlight := f.Highlight(word)

	return replaceFirstInsensitive(sentence, word, highlight)
}

// replaceFirstInsensitive replaces the first case-insensitive match while
// preserving the original text around the match.
func replaceFirstInsensitive(s, old, new string) string {
	lowerS := strings.ToLower(s)
	lowerOld := strings.ToLower(old)

	idx := strings.Index(lowerS, lowerOld)
	if idx < 0 {
		return s
	}

	return s[:idx] + new + s[idx+len(old):]
}
