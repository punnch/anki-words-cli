package config

import (
	"strings"
	"testing"
)

func TestConfigValidateRequiresOpenAIKey(t *testing.T) {
	cfg := validConfig()
	cfg.OpenAIAPIKey = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("expected missing OpenAI key, got %q", err)
	}
}

func TestConfigValidateAcceptsRequiredSettings(t *testing.T) {
	cfg := validConfig()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func validConfig() Config {
	return Config{
		OpenAIAPIKey:   "openai-key",
		OpenAIBaseURL:  "https://api.openai.com/v1",
		OpenAIModel:    "gpt-test",
		AnkiConnectURL: "http://localhost:8765",
		SettingsFile:   "./out/settings/settings.json",
		LoggerLevel:    "DEBUG",
		LoggerFolder:   "./out/logs",
	}
}
