// Package config loads and validates runtime configuration from environment
// variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config contains all runtime settings needed to start the TUI and call its
// external dependencies.
type Config struct {
	OpenAIAPIKey            string
	OpenAIBaseURL           string
	OpenAIModel             string
	AnkiConnectURL          string
	SettingsFile            string
	DefaultGenerationPrompt string
	SentenceHighlightColor  string
	LoggerLevel             string
	LoggerFolder            string
}

// Load reads configuration from environment variables and applies safe
// defaults for optional settings.
func Load() Config {
	return Config{
		OpenAIAPIKey:            os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:           envDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:             envDefault("OPENAI_MODEL", "gpt-5.4-mini"),
		AnkiConnectURL:          os.Getenv("ANKICONNECT_URL"),
		SettingsFile:            envDefault("SETTINGS_FILE", "./out/settings/settings.json"),
		DefaultGenerationPrompt: defaultGenerationPrompt(),
		SentenceHighlightColor:  envDefault("SENTENCE_HIGHLIGHT_COLOR", "#00557f"),
		LoggerLevel:             envDefault("LOGGER_LEVEL", "DEBUG"),
		LoggerFolder:            envDefault("LOGGER_FOLDER", "./out/logs"),
	}
}

// Validate checks required settings and catches invalid values before the TUI
// starts.
func (c Config) Validate() error {
	var errs []error

	if strings.TrimSpace(c.OpenAIAPIKey) == "" {
		errs = append(errs, fmt.Errorf("OPENAI_API_KEY is required"))
	}
	if strings.TrimSpace(c.OpenAIBaseURL) == "" {
		errs = append(errs, fmt.Errorf("OPENAI_BASE_URL is required"))
	}
	if strings.TrimSpace(c.OpenAIModel) == "" {
		errs = append(errs, fmt.Errorf("OPENAI_MODEL is required"))
	}
	if strings.TrimSpace(c.AnkiConnectURL) == "" {
		errs = append(errs, fmt.Errorf("ANKICONNECT_URL is required"))
	}
	if strings.TrimSpace(c.SettingsFile) == "" {
		errs = append(errs, fmt.Errorf("SETTINGS_FILE is required"))
	}
	if strings.TrimSpace(c.LoggerLevel) == "" {
		errs = append(errs, fmt.Errorf("LOGGER_LEVEL is required"))
	}
	if strings.TrimSpace(c.LoggerFolder) == "" {
		errs = append(errs, fmt.Errorf("LOGGER_FOLDER is required"))
	}

	return errors.Join(errs...)
}

// envDefault returns an environment variable value or a fallback when unset.
func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// defaultGenerationPrompt is the initial card generation template.
func defaultGenerationPrompt() string {
	return `Create one Anki card per target word.

Requirements:
- Read the selected Anki model's field names and match each field with the most appropriate content.
- Use field names as the source of truth. For example, Front/Back should become a question/prompt and answer; Word/Sentence/Translation should become the target word, example sentence, and translation.
- Fill every available field with useful vocabulary-learning content.
- Include English definitions, IPA transcription, Russian translation, and natural examples when the field names ask for them.
- If a Sentence field exists, highlight the target word with the exact HTML provided by the application.`
}
