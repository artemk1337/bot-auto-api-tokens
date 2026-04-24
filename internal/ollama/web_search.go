package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type WebSearchClient struct {
	baseURL    string
	apiKey     string
	maxResults int
	httpClient *http.Client
}

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

func NewWebSearchClient(baseURL, apiKey string, maxResults int) WebSearchClient {
	return WebSearchClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		maxResults: maxResults,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c WebSearchClient) Search(ctx context.Context, query string) ([]SearchResult, error) {
	body, err := json.Marshal(webSearchRequest{
		Query:      query,
		MaxResults: c.maxResults,
	})
	if err != nil {
		return nil, fmt.Errorf("encode web search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/web_search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create web search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call web search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("web search returned status %s", resp.Status)
	}

	var out webSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode web search response: %w", err)
	}
	return out.Results, nil
}

type webSearchRequest struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
}

type webSearchResponse struct {
	Results []SearchResult `json:"results"`
}
