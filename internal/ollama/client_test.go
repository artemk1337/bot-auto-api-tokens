package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientChat(t *testing.T) {
	var got chatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":" answer "}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "llama3.2", 0.3, map[string]any{"num_ctx": float64(4096)})
	answer, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatal(err)
	}

	if answer != "answer" {
		t.Fatalf("answer = %q", answer)
	}
	if got.Model != "llama3.2" {
		t.Fatalf("model = %q", got.Model)
	}
	if got.Stream {
		t.Fatal("stream must be false")
	}
	if got.Options["temperature"] != 0.3 {
		t.Fatalf("temperature = %#v", got.Options["temperature"])
	}
	if got.Options["num_ctx"] != float64(4096) {
		t.Fatalf("num_ctx = %#v", got.Options["num_ctx"])
	}
}

func TestClientChatReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "llama3.2", 0, nil)
	if _, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}); err == nil {
		t.Fatal("expected status error")
	}
}
