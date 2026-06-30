package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/issgo/issgo/config"
)

// OllamaProvider implements Provider for local Ollama instances.
type OllamaProvider struct {
	cfg    *config.Config
	client *http.Client
}

func NewOllamaProvider(cfg *config.Config) *OllamaProvider {
	return &OllamaProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: Duration(cfg.LLM.TimeoutSecs),
		},
	}
}

func (p *OllamaProvider) Name() string            { return "ollama" }
func (p *OllamaProvider) SupportsStreaming() bool { return true }
func (p *OllamaProvider) SupportsTools() bool     { return false } // not natively

func (p *OllamaProvider) baseURL() string {
	if p.cfg.LLM.BaseURL != "" {
		return strings.TrimRight(p.cfg.LLM.BaseURL, "/")
	}
	return "http://localhost:11434"
}

type ollamaReq struct {
	Model    string         `json:"model"`
	Messages []Message      `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

type ollamaResp struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

func (p *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body := ollamaReq{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   false,
		Options: map[string]any{
			"temperature": req.Temperature,
		},
	}

	b, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL()+"/api/chat", bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama read: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ollama %d: %s", resp.StatusCode, string(data))
	}

	var oresp ollamaResp
	if err := json.Unmarshal(data, &oresp); err != nil {
		return nil, fmt.Errorf("ollama decode: %w", err)
	}

	return &ChatResponse{
		Choices: []Choice{{
			Index:        0,
			Message:      oresp.Message,
			FinishReason: "stop",
		}},
	}, nil
}

func (p *OllamaProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	body := ollamaReq{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   true,
		Options: map[string]any{
			"temperature": req.Temperature,
		},
	}

	b, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL()+"/api/chat", bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama stream: %w", err)
	}

	ch := make(chan StreamChunk, 100)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var o ollamaResp
			if err := json.Unmarshal(scanner.Bytes(), &o); err != nil {
				continue
			}
			ch <- StreamChunk{Content: o.Message.Content, Done: o.Done}
			if o.Done {
				return
			}
		}
	}()

	return ch, nil
}

// OllamaURL sets the default ollama connection URL for convenience
func OllamaURL(host string) string {
	if host == "" {
		host = "localhost:11434"
	}
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}
	return host
}

// WaitForOllama polls ollama until it's ready.
func WaitForOllama(ctx context.Context, baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/api/tags")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("ollama not ready after %v", timeout)
}
