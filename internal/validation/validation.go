// Package validation checks generated card data before it is written to Anki.
package validation

import (
	"fmt"
	"strings"

	"github.com/punnch/ankiwords/internal/model"
)

// ValidateGeneratedCards enforces the contract between OpenAI generation and
// Anki note creation.
func ValidateGeneratedCards(
	cards []model.GeneratedCard,
	words []string,
	modelFields []string,
	formatter model.SentenceFormatter,
) error {
	if len(cards) != len(words) {
		return fmt.Errorf("card count mismatch: got %d want %d", len(cards), len(words))
	}
	if len(modelFields) == 0 {
		return fmt.Errorf("model has no fields")
	}

	expectedFields := make(map[string]bool, len(modelFields))
	for _, field := range modelFields {
		field = strings.TrimSpace(field)
		if field == "" {
			return fmt.Errorf("model has an empty field name")
		}
		expectedFields[field] = true
	}

	for i, card := range cards {
		if strings.TrimSpace(card.Word) == "" {
			return fmt.Errorf("card %d has empty word", i)
		}
		if card.Fields == nil {
			return fmt.Errorf("card %d has no fields", i)
		}

		if !strings.EqualFold(strings.TrimSpace(card.Word), strings.TrimSpace(words[i])) {
			return fmt.Errorf("card %d word mismatch: got %q want %q", i, card.Word, words[i])
		}

		for field := range card.Fields {
			if !expectedFields[field] {
				return fmt.Errorf("card %d has unknown field %q", i, field)
			}
		}

		for _, field := range modelFields {
			value, ok := card.Fields[field]
			if !ok {
				return fmt.Errorf("card %d is missing field %q", i, field)
			}
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("card %d has empty field %q", i, field)
			}
		}

		if sentence := card.Fields["Sentence"]; sentence != "" {
			expected := formatter.Highlight(strings.TrimSpace(card.Word))
			if !strings.Contains(sentence, expected) {
				return fmt.Errorf("card %d sentence is missing required highlight for %q", i, card.Word)
			}
		}
	}

	return nil
}
