package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/internal/logger"
)

// Client is the primary LLM client that selects the right provider
// and adds retry, caching, and telemetry.
type Client struct {
	cfg       *config.Config
	providers map[string]Provider
	cache     *Cache
	mu        sync.RWMutex
}

func NewClient(cfg *config.Config) *Client {
	c := &Client{
		cfg:       cfg,
		providers: make(map[string]Provider),
		cache:     NewCache(200, 30*time.Minute),
	}

	// Register built-in providers
	openai := NewOpenAIProvider(cfg)
	c.providers["openai"] = openai
	c.providers["deepseek"] = openai // DeepSeek is OpenAI-compatible
	c.providers["ollama"] = NewOllamaProvider(cfg)
	// "custom" falls through to openai as well
	c.providers["custom"] = openai

	return c
}

func (c *Client) getProvider() Provider {
	name := c.cfg.LLM.Provider
	c.mu.RLock()
	p, ok := c.providers[name]
	c.mu.RUnlock()
	if ok {
		return p
	}
	// default to openai-compatible
	return c.providers["openai"]
}

// ─── Chat ──────────────────────────────────────────────────────

func (c *Client) Chat(ctx context.Context, system string, history []Message, tools []ToolDef) (*ChatResponse, error) {
	return c.ChatWithCallbacks(ctx, system, history, tools, nil)
}

func (c *Client) ChatWithCallbacks(ctx context.Context, system string, history []Message, tools []ToolDef, cb *Callbacks) (*ChatResponse, error) {
	provider := c.getProvider()
	model := c.cfg.LLM.Model

	msgs := make([]Message, 0, len(history)+2)
	if system != "" {
		msgs = append(msgs, SystemMsg(system))
	}
	msgs = append(msgs, history...)

	req := ChatRequest{
		Model:       model,
		Messages:    msgs,
		Tools:       tools,
		Temperature: c.cfg.LLM.Temperature,
		MaxTokens:   c.cfg.LLM.MaxTokens,
	}

	// Try cache first (for non-streaming, no-tools calls)
	cacheKey := ""
	if len(tools) == 0 {
		cacheKey = c.cacheKey(system, history)
		if entry, ok := c.cache.Get(cacheKey); ok {
			logger.Log.Debugw("cache hit")
			return entry, nil
		}
	}

	// Execute with retry
	var resp *ChatResponse
	var err error
	timeout := Duration(c.cfg.LLM.TimeoutSecs)

	for attempt := 0; attempt <= c.cfg.LLM.RetryCount; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, timeout)
		resp, err = provider.Chat(attemptCtx, req)
		cancel()

		if err == nil {
			break
		}
		if cb != nil && cb.OnRetry != nil {
			cb.OnRetry(attempt, err)
		}
		logger.Log.Warnw("llm retry", "attempt", attempt, "error", err)
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	if err != nil {
		// Attempt fallback to uncached call
		return nil, fmt.Errorf("chat: %w", err)
	}

	// Cache successful non-tool responses
	if cacheKey != "" && resp != nil {
		c.cache.Set(cacheKey, resp)
	}

	return resp, nil
}

// ─── Streaming ─────────────────────────────────────────────────

func (c *Client) ChatStream(ctx context.Context, system string, history []Message, tools []ToolDef) (<-chan StreamChunk, error) {
	provider := c.getProvider()
	if !provider.SupportsStreaming() {
		return nil, fmt.Errorf("provider %s does not support streaming", provider.Name())
	}

	msgs := make([]Message, 0, len(history)+2)
	if system != "" {
		msgs = append(msgs, SystemMsg(system))
	}
	msgs = append(msgs, history...)

	req := ChatRequest{
		Model:       c.cfg.LLM.Model,
		Messages:    msgs,
		Tools:       tools,
		Temperature: c.cfg.LLM.Temperature,
		MaxTokens:   c.cfg.LLM.MaxTokens,
		Stream:      true,
	}

	return provider.ChatStream(ctx, req)
}

// ─── Cache ─────────────────────────────────────────────────────

func (c *Client) cacheKey(system string, history []Message) string {
	h := fmt.Sprintf("%s|%s|%s|%v|%v",
		c.cfg.LLM.Provider, c.cfg.LLM.Model, system,
		history, c.cfg.LLM.Temperature)
	return fmt.Sprintf("%x", h)
}

func (c *Client) InvalidateCache() {
	c.cache.Clear()
}

func (c *Client) CacheStats() (size, capacity int, ttl time.Duration) {
	return c.cache.Len(), c.cache.Cap(), c.cache.TTL()
}
