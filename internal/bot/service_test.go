package bot

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artemk1337/bot-auto-api-tokens/internal/ollama"
	"github.com/artemk1337/bot-auto-api-tokens/internal/telegram"
)

func TestServiceAddsPromptDocsAndHistory(t *testing.T) {
	docPath := filepath.Join(t.TempDir(), "support.md")
	if err := os.WriteFile(docPath, []byte("Use product docs."), 0o600); err != nil {
		t.Fatal(err)
	}

	tg := &fakeTelegram{}
	model := &fakeModel{answer: "first answer"}
	service, err := NewService(tg, model, Config{
		SystemPrompt:       "Base prompt.",
		DocumentationFiles: []string{docPath},
		HistoryLimit:       2,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = service.HandleMessage(context.Background(), telegram.Message{
		From: telegram.User{ID: 1},
		Chat: telegram.Chat{ID: 10},
		Text: "question one",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(model.requests) != 1 {
		t.Fatalf("requests = %d", len(model.requests))
	}
	first := model.requests[0]
	if first[0].Role != "system" || !strings.Contains(first[0].Content, "Base prompt.") || !strings.Contains(first[0].Content, "Use product docs.") {
		t.Fatalf("system prompt = %#v", first[0])
	}
	if first[len(first)-1].Content != "question one" {
		t.Fatalf("last message = %#v", first[len(first)-1])
	}
	if tg.sent[0].text != "first answer" {
		t.Fatalf("sent = %q", tg.sent[0].text)
	}

	model.answer = "second answer"
	err = service.HandleMessage(context.Background(), telegram.Message{
		From: telegram.User{ID: 1},
		Chat: telegram.Chat{ID: 10},
		Text: "question two",
	})
	if err != nil {
		t.Fatal(err)
	}

	second := model.requests[1]
	if len(second) != 4 {
		t.Fatalf("messages count = %d: %#v", len(second), second)
	}
	if second[1].Content != "question one" || second[2].Content != "first answer" || second[3].Content != "question two" {
		t.Fatalf("history was not included correctly: %#v", second)
	}
}

func TestServiceDeniesUnknownUser(t *testing.T) {
	tg := &fakeTelegram{}
	model := &fakeModel{answer: "answer"}
	service, err := NewService(tg, model, Config{
		AllowedUserIDs: []int64{42},
		HistoryLimit:   10,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = service.HandleMessage(context.Background(), telegram.Message{
		From: telegram.User{ID: 1},
		Chat: telegram.Chat{ID: 10},
		Text: "question",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(model.requests) != 0 {
		t.Fatalf("model requests = %d", len(model.requests))
	}
	if len(tg.sent) != 1 || !strings.Contains(tg.sent[0].text, "нет доступа") {
		t.Fatalf("sent = %#v", tg.sent)
	}
}

type fakeTelegram struct {
	sent []sentMessage
}

func (f *fakeTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	f.sent = append(f.sent, sentMessage{chatID: chatID, text: text})
	return nil
}

type sentMessage struct {
	chatID int64
	text   string
}

type fakeModel struct {
	answer   string
	requests [][]ollama.Message
}

func (f *fakeModel) Chat(ctx context.Context, messages []ollama.Message) (string, error) {
	copied := make([]ollama.Message, len(messages))
	copy(copied, messages)
	f.requests = append(f.requests, copied)
	return f.answer, nil
}
