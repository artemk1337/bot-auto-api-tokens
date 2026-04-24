package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Telegram TelegramConfig `json:"telegram"`
	Ollama   OllamaConfig   `json:"ollama"`
	Bot      BotConfig      `json:"bot"`
}

type TelegramConfig struct {
	Token              string  `json:"token"`
	PollTimeoutSeconds int     `json:"poll_timeout_seconds"`
	AllowedUserIDs     []int64 `json:"allowed_user_ids"`
}

type OllamaConfig struct {
	BaseURL     string         `json:"base_url"`
	Model       string         `json:"model"`
	Temperature float64        `json:"temperature"`
	Options     map[string]any `json:"options"`
}

type BotConfig struct {
	HistoryLimit       int      `json:"history_limit"`
	SystemPrompt       string   `json:"system_prompt"`
	DocumentationFiles []string `json:"documentation_files"`
}

func Load(path string) (Config, error) {
	if path == "" {
		return Config{}, errors.New("config path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	data = []byte(os.ExpandEnv(string(data)))

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	cfg.setDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) setDefaults() {
	if c.Telegram.PollTimeoutSeconds == 0 {
		c.Telegram.PollTimeoutSeconds = 30
	}
	if c.Ollama.BaseURL == "" {
		c.Ollama.BaseURL = "http://localhost:11434"
	}
	if c.Bot.HistoryLimit == 0 {
		c.Bot.HistoryLimit = 10
	}
	if c.Ollama.Options == nil {
		c.Ollama.Options = map[string]any{}
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Telegram.Token) == "" {
		return errors.New("telegram.token is required")
	}
	if strings.TrimSpace(c.Ollama.Model) == "" {
		return errors.New("ollama.model is required")
	}
	if c.Bot.HistoryLimit < 0 {
		return errors.New("bot.history_limit must be greater than or equal to 0")
	}
	if c.Telegram.PollTimeoutSeconds < 1 {
		return errors.New("telegram.poll_timeout_seconds must be greater than 0")
	}
	return nil
}

func (c TelegramConfig) PollTimeout() time.Duration {
	return time.Duration(c.PollTimeoutSeconds) * time.Second
}
