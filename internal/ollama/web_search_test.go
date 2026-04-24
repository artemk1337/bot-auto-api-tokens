package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebSearchClientSearch(t *testing.T) {
	var got webSearchRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/web_search" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer key-123" {
			t.Fatalf("auth = %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"results":[{"title":"Doc","url":"https://example.com","content":"Snippet"}]}`))
	}))
	defer server.Close()

	client := NewWebSearchClient(server.URL, "key-123", 3)
	results, err := client.Search(context.Background(), "query")
	if err != nil {
		t.Fatal(err)
	}

	if got.Query != "query" || got.MaxResults != 3 {
		t.Fatalf("request = %#v", got)
	}
	if len(results) != 1 || results[0].Title != "Doc" {
		t.Fatalf("results = %#v", results)
	}
}
