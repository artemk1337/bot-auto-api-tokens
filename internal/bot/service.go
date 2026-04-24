package bot

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/artemk1337/bot-auto-api-tokens/internal/ollama"
	"github.com/artemk1337/bot-auto-api-tokens/internal/telegram"
)

type Telegram interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type Model interface {
	Chat(ctx context.Context, messages []ollama.Message) (string, error)
}

type Searcher interface {
	Search(ctx context.Context, query string) ([]ollama.SearchResult, error)
}

type Service struct {
	tg             Telegram
	model          Model
	searcher       Searcher
	systemPrompt   string
	historyLimit   int
	allowedUserIDs map[int64]struct{}

	mu      sync.Mutex
	history map[int64][]ollama.Message
}

type Config struct {
	SystemPrompt       string
	DocumentationFiles []string
	HistoryLimit       int
	AllowedUserIDs     []int64
	Searcher           Searcher
}

func NewService(tg Telegram, model Model, cfg Config) (Service, error) {
	prompt, err := buildSystemPrompt(cfg.SystemPrompt, cfg.DocumentationFiles)
	if err != nil {
		return Service{}, err
	}

	allowedUserIDs := make(map[int64]struct{}, len(cfg.AllowedUserIDs))
	for _, id := range cfg.AllowedUserIDs {
		allowedUserIDs[id] = struct{}{}
	}

	return Service{
		tg:             tg,
		model:          model,
		searcher:       cfg.Searcher,
		systemPrompt:   prompt,
		historyLimit:   cfg.HistoryLimit,
		allowedUserIDs: allowedUserIDs,
		history:        map[int64][]ollama.Message{},
	}, nil
}

func (s *Service) HandleMessage(ctx context.Context, msg telegram.Message) error {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil
	}
	if !s.isAllowed(msg.From.ID) {
		return s.tg.SendMessage(ctx, msg.Chat.ID, "У вас нет доступа к этому боту.")
	}

	messages := s.messagesForRequest(msg.Chat.ID, text)
	if s.searcher != nil {
		searchResults, err := s.searcher.Search(ctx, text)
		if err != nil {
			return s.tg.SendMessage(ctx, msg.Chat.ID, "Не удалось выполнить поиск: "+err.Error())
		}
		messages = addSearchResults(messages, searchResults)
	}
	answer, err := s.model.Chat(ctx, messages)
	if err != nil {
		return s.tg.SendMessage(ctx, msg.Chat.ID, "Не удалось получить ответ от модели: "+err.Error())
	}
	if answer == "" {
		answer = "Модель вернула пустой ответ."
	}

	s.appendHistory(msg.Chat.ID, ollama.Message{Role: "user", Content: text})
	s.appendHistory(msg.Chat.ID, ollama.Message{Role: "assistant", Content: answer})

	return s.tg.SendMessage(ctx, msg.Chat.ID, answer)
}

func (s *Service) messagesForRequest(chatID int64, userText string) []ollama.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	history := s.trimmedHistoryLocked(chatID)
	messages := make([]ollama.Message, 0, len(history)+2)
	if s.systemPrompt != "" {
		messages = append(messages, ollama.Message{Role: "system", Content: s.systemPrompt})
	}
	messages = append(messages, history...)
	messages = append(messages, ollama.Message{Role: "user", Content: userText})
	return messages
}

func (s *Service) appendHistory(chatID int64, msg ollama.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history[chatID] = append(s.history[chatID], msg)
	s.history[chatID] = trimMessages(s.history[chatID], s.historyLimit)
}

func (s *Service) trimmedHistoryLocked(chatID int64) []ollama.Message {
	history := trimMessages(s.history[chatID], s.historyLimit)
	out := make([]ollama.Message, len(history))
	copy(out, history)
	return out
}

func (s *Service) isAllowed(userID int64) bool {
	if len(s.allowedUserIDs) == 0 {
		return true
	}
	_, ok := s.allowedUserIDs[userID]
	return ok
}

func trimMessages(messages []ollama.Message, limit int) []ollama.Message {
	if limit <= 0 || len(messages) <= limit {
		return messages
	}
	return messages[len(messages)-limit:]
}

func buildSystemPrompt(base string, files []string) (string, error) {
	parts := []string{strings.TrimSpace(base)}
	for _, path := range files {
		if strings.TrimSpace(path) == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read documentation file %q: %w", path, err)
		}
		parts = append(parts, "Документация из файла "+path+":\n"+strings.TrimSpace(string(data)))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n")), nil
}

func addSearchResults(messages []ollama.Message, results []ollama.SearchResult) []ollama.Message {
	if len(results) == 0 {
		return messages
	}

	searchMessage := ollama.Message{
		Role:    "system",
		Content: formatSearchResults(results),
	}
	if len(messages) == 0 {
		return []ollama.Message{searchMessage}
	}

	out := make([]ollama.Message, 0, len(messages)+1)
	out = append(out, messages[:len(messages)-1]...)
	out = append(out, searchMessage)
	out = append(out, messages[len(messages)-1])
	return out
}

func formatSearchResults(results []ollama.SearchResult) string {
	var b strings.Builder
	b.WriteString("Результаты веб-поиска для текущего вопроса. Используй их только если они релевантны, и не выдумывай факты:\n")
	for i, result := range results {
		b.WriteString("\n")
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(". ")
		b.WriteString(strings.TrimSpace(result.Title))
		if result.URL != "" {
			b.WriteString("\nURL: ")
			b.WriteString(strings.TrimSpace(result.URL))
		}
		if result.Content != "" {
			b.WriteString("\nФрагмент: ")
			b.WriteString(strings.TrimSpace(result.Content))
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
