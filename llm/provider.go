// Package llm defines the LLM abstraction layer.
package llm

import (
	"context"
	"time"
)

// ─── Types ─────────────────────────────────────────────────────

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ToolCall struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type Choice struct {
	Index        int        `json:"index"`
	Message      Message    `json:"message"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []ToolDef `json:"tools,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// StreamChunk is a single chunk from a streaming response.
type StreamChunk struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls"`
	Done      bool       `json:"done"`
	Error     error      `json:"-"`
}

// ─── Provider Interface ────────────────────────────────────────

type Provider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
	SupportsStreaming() bool
	SupportsTools() bool
}

// ─── Callbacks ─────────────────────────────────────────────────

type Callbacks struct {
	OnStart    func()
	OnChunk    func(StreamChunk)
	OnComplete func(*ChatResponse)
	OnError    func(error)
	OnRetry    func(attempt int, err error)
}

// ─── Helper constructors ───────────────────────────────────────

func UserMsg(content string) Message {
	return Message{Role: "user", Content: content}
}

func AssistantMsg(content string) Message {
	return Message{Role: "assistant", Content: content}
}

func SystemMsg(content string) Message {
	return Message{Role: "system", Content: content}
}

func ToolMsg(toolCallID, name, content string) Message {
	return Message{Role: "tool", Content: content, ToolCallID: toolCallID, Name: name}
}

func AssistantToolCalls(calls []ToolCall) Message {
	return Message{Role: "assistant", ToolCalls: calls}
}

// Duration helper
func Duration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}
