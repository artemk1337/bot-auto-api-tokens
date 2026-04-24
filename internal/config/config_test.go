package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExpandsEnvAndDefaults(t *testing.T) {
	t.Setenv("TEST_TG_TOKEN", "token-123")

	path := filepath.Join(t.TempDir(), "config.json")
	err := os.WriteFile(path, []byte(`{
		"telegram": {"token": "${TEST_TG_TOKEN}"},
		"ollama": {"model": "llama3.2"},
		"bot": {}
	}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Telegram.Token != "token-123" {
		t.Fatalf("token = %q", cfg.Telegram.Token)
	}
	if cfg.Telegram.PollTimeoutSeconds != 30 {
		t.Fatalf("poll timeout = %d", cfg.Telegram.PollTimeoutSeconds)
	}
	if cfg.Ollama.BaseURL != "http://localhost:11434" {
		t.Fatalf("base url = %q", cfg.Ollama.BaseURL)
	}
	if cfg.Bot.HistoryLimit != 10 {
		t.Fatalf("history limit = %d", cfg.Bot.HistoryLimit)
	}
}

func TestLoadValidatesRequiredFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	err := os.WriteFile(path, []byte(`{"telegram": {}, "ollama": {}, "bot": {}}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadModelConfigs(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "token-123")
	t.Setenv("OLLAMA_BASE_URL", "http://localhost:11434")
	t.Setenv("OLLAMA_API_KEY", "ollama-key")

	paths, err := filepath.Glob("../../configs/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("model configs not found")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Ollama.Model == "" {
				t.Fatal("ollama model is empty")
			}
		})
	}
}
