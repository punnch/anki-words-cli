package validation

import (
	"testing"

	"github.com/punnch/ankiwords/internal/model"
)

func TestValidateGeneratedCards(t *testing.T) {
	formatter := model.SentenceFormatter{Color: "#00557f"}
	cards := []model.GeneratedCard{{
		Word: "collapse",
		Fields: map[string]string{
			"Sentence":      "The system will <span style=\"color:#00557f;\"><b>collapse</b></span> under pressure.",
			"Definition":    "fall apart",
			"Transcription": "/kəˈlæps/",
			"Word":          "collapse",
			"Translation":   "разрушаться",
		},
	}}

	fields := []string{"Sentence", "Definition", "Transcription", "Word", "Translation"}
	if err := ValidateGeneratedCards(cards, []string{"collapse"}, fields, formatter); err != nil {
		t.Fatal(err)
	}
}

func TestValidateGeneratedCardsAllowsFrontBackModel(t *testing.T) {
	formatter := model.SentenceFormatter{Color: "#00557f"}
	cards := []model.GeneratedCard{{
		Word: "collapse",
		Fields: map[string]string{
			"Front": "collapse",
			"Back":  "to fall apart; разрушаться",
		},
	}}

	if err := ValidateGeneratedCards(cards, []string{"collapse"}, []string{"Front", "Back"}, formatter); err != nil {
		t.Fatal(err)
	}
}

func TestValidateGeneratedCardsRejectsUnknownField(t *testing.T) {
	formatter := model.SentenceFormatter{Color: "#00557f"}
	cards := []model.GeneratedCard{{
		Word: "collapse",
		Fields: map[string]string{
			"Front": "collapse",
			"Back":  "to fall apart",
			"Extra": "unexpected",
		},
	}}

	err := ValidateGeneratedCards(cards, []string{"collapse"}, []string{"Front", "Back"}, formatter)
	if err == nil {
		t.Fatal("expected validation error")
	}
}
