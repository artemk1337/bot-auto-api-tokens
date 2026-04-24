package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type Chat struct {
	ID int64 `json:"id"`
}

func NewClient(token string) Client {
	return Client{
		baseURL:    "https://api.telegram.org/bot" + token,
		httpClient: &http.Client{Timeout: time.Minute},
	}
}

func (c Client) GetUpdates(ctx context.Context, offset int, timeout time.Duration) ([]Update, error) {
	values := url.Values{}
	values.Set("offset", strconv.Itoa(offset))
	values.Set("timeout", strconv.Itoa(int(timeout.Seconds())))
	values.Set("allowed_updates", `["message"]`)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/getUpdates?"+values.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create getUpdates request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call getUpdates: %w", err)
	}
	defer resp.Body.Close()

	var out updatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode getUpdates response: %w", err)
	}
	if !out.OK {
		return nil, fmt.Errorf("telegram getUpdates failed with status %s", resp.Status)
	}
	return out.Result, nil
}

func (c Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	body, err := json.Marshal(sendMessageRequest{
		ChatID: chatID,
		Text:   strings.TrimSpace(text),
	})
	if err != nil {
		return fmt.Errorf("encode sendMessage request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create sendMessage request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call sendMessage: %w", err)
	}
	defer resp.Body.Close()

	var out telegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("decode sendMessage response: %w", err)
	}
	if !out.OK {
		return fmt.Errorf("telegram sendMessage failed with status %s", resp.Status)
	}
	return nil
}

type updatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type telegramResponse struct {
	OK bool `json:"ok"`
}

type sendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}
