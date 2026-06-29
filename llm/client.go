package llm

import (
	"context"
	"fmt"

	"github.com/issgo/issgo/config"
	"github.com/sashabaranov/go-openai"
)

type Client struct {
	api     *openai.Client
	model   string
	verbose bool
}

func NewClient(cfg *config.Config) *Client {
	ocfg := openai.DefaultConfig(cfg.LLM.APIKey)
	ocfg.BaseURL = cfg.LLM.BaseURL
	return &Client{
		api:     openai.NewClientWithConfig(ocfg),
		model:   cfg.LLM.Model,
		verbose: cfg.Agent.Verbose,
	}
}

func (c *Client) Chat(ctx context.Context, system string, history []Message, tools []ToolDef) (*ChatResponse, error) {
	msgs := make([]openai.ChatCompletionMessage, 0, len(history)+2)

	if system != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: system,
		})
	}

	for _, h := range history {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    h.Role,
			Content: h.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: msgs,
	}

	if len(tools) > 0 {
		req.Tools = toOpenAITools(tools)
	}

	resp, err := c.api.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("chat completion: %w", err)
	}

	return toChatResponse(resp), nil
}

func toOpenAITools(tools []ToolDef) []openai.Tool {
	result := make([]openai.Tool, len(tools))
	for i, t := range tools {
		result[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return result
}

func toChatResponse(resp openai.ChatCompletionResponse) *ChatResponse {
	cr := &ChatResponse{
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	for _, choice := range resp.Choices {
		c := Choice{
			FinishReason: string(choice.FinishReason),
			Message: Message{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
		}

		for _, tc := range choice.Message.ToolCalls {
			c.ToolCalls = append(c.ToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}

		cr.Choices = append(cr.Choices, c)
	}

	return cr
}
