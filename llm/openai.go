package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/issgo/issgo/config"
	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements Provider for OpenAI-compatible APIs.
type OpenAIProvider struct {
	cfg    *config.Config
	client *openai.Client
}

func NewOpenAIProvider(cfg *config.Config) *OpenAIProvider {
	ocfg := openai.DefaultConfig(cfg.LLM.APIKey)
	ocfg.BaseURL = strings.TrimRight(cfg.LLM.BaseURL, "/")
	return &OpenAIProvider{
		cfg:    cfg,
		client: openai.NewClientWithConfig(ocfg),
	}
}

func (p *OpenAIProvider) Name() string           { return "openai" }
func (p *OpenAIProvider) SupportsStreaming() bool { return true }
func (p *OpenAIProvider) SupportsTools() bool     { return true }

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	msgs := toOpenAIMsgs(req.Messages)
	tools := toOpenAITools(req.Tools)

	creq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Tools:       tools,
	}

	resp, err := p.client.CreateChatCompletion(ctx, creq)
	if err != nil {
		return nil, fmt.Errorf("openai chat: %w", err)
	}

	return fromOpenAI(resp), nil
}

func (p *OpenAIProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	req.Stream = true
	msgs := toOpenAIMsgs(req.Messages)
	tools := toOpenAITools(req.Tools)

	creq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Tools:       tools,
		Stream:      true,
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, creq)
	if err != nil {
		return nil, fmt.Errorf("openai stream: %w", err)
	}

	ch := make(chan StreamChunk, 100)
	go func() {
		defer close(ch)
		defer stream.Close()
		for {
			resp, err := stream.Recv()
			if err != nil {
				if !strings.Contains(err.Error(), "EOF") {
					ch <- StreamChunk{Error: err, Done: true}
				}
				return
			}
			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta
				chunk := StreamChunk{Content: delta.Content}
				for _, tc := range delta.ToolCalls {
					chunk.ToolCalls = append(chunk.ToolCalls, ToolCall{
						ID:        tc.ID,
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					})
				}
				ch <- chunk
			}
		}
	}()

	return ch, nil
}

// ─── converters ────────────────────────────────────────────────

func toOpenAIMsgs(msgs []Message) []openai.ChatCompletionMessage {
	out := make([]openai.ChatCompletionMessage, len(msgs))
	for i, m := range msgs {
		o := openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
		if m.ToolCallID != "" {
			o.ToolCallID = m.ToolCallID
		}
		if m.Name != "" {
			o.Name = m.Name
		}
		if len(m.ToolCalls) > 0 {
			o.ToolCalls = make([]openai.ToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				o.ToolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
		}
		out[i] = o
	}
	return out
}

func toOpenAITools(tools []ToolDef) []openai.Tool {
	out := make([]openai.Tool, len(tools))
	for i, t := range tools {
		out[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return out
}

func fromOpenAI(resp openai.ChatCompletionResponse) *ChatResponse {
	cr := &ChatResponse{
		ID: resp.ID,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
	for _, choice := range resp.Choices {
		c := Choice{
			Index:        choice.Index,
			FinishReason: string(choice.FinishReason),
			Message: Message{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
		}
		for _, tc := range choice.Message.ToolCalls {
			c.ToolCalls = append(c.ToolCalls, ToolCall{
				ID:        tc.ID,
				Type:      string(tc.Type),
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
			c.Message.ToolCalls = c.ToolCalls
		}
		cr.Choices = append(cr.Choices, c)
	}
	return cr
}
