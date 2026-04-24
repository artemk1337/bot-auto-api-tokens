package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artemk1337/bot-auto-api-tokens/internal/bot"
	"github.com/artemk1337/bot-auto-api-tokens/internal/config"
	"github.com/artemk1337/bot-auto-api-tokens/internal/ollama"
	"github.com/artemk1337/bot-auto-api-tokens/internal/telegram"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	queueSize := flag.Int("queue-size", 100, "message queue size")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	tg := telegram.NewClient(cfg.Telegram.Token)
	model := ollama.NewClient(cfg.Ollama.BaseURL, cfg.Ollama.Model, cfg.Ollama.Temperature, cfg.Ollama.Options)

	service, err := bot.NewService(tg, model, bot.Config{
		SystemPrompt:       cfg.Bot.SystemPrompt,
		DocumentationFiles: cfg.Bot.DocumentationFiles,
		HistoryLimit:       cfg.Bot.HistoryLimit,
		AllowedUserIDs:     cfg.Telegram.AllowedUserIDs,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner := bot.NewRunner(tg, &service, cfg.Telegram.PollTimeout(), *queueSize)
	if err := runner.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
