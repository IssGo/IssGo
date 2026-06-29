package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/prompts"
	"github.com/issgo/issgo/tools"
)

type Executor struct {
	client   *llm.Client
	registry *tools.Registry
	memory   *Memory
}

func NewExecutor(client *llm.Client, registry *tools.Registry, memory *Memory) *Executor {
	return &Executor{client: client, registry: registry, memory: memory}
}

func (e *Executor) Run(ctx context.Context, task string, maxSteps int) (string, error) {
	sysPrompt := e.buildSystemPrompt()
	toolDefs := e.registry.List()

	e.memory.Add("user", task)

	for step := 0; step < maxSteps; step++ {
		resp, err := e.client.Chat(ctx, sysPrompt, e.memory.History(), toolDefs)
		if err != nil {
			return "", fmt.Errorf("step %d: %w", step, err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("step %d: no response from LLM", step)
		}

		choice := resp.Choices[0]

		// If the LLM made tool calls, execute them.
		if len(choice.ToolCalls) > 0 {
			for _, tc := range choice.ToolCalls {
				e.memory.Add("assistant", fmt.Sprintf("[tool:%s] %s", tc.Name, tc.Arguments))
				result := e.registry.Execute(ctx, tc.Name, json.RawMessage(tc.Arguments))
				output := result.Output
				if !result.Success {
					output = "ERROR: " + result.Error
					if result.Output != "" {
						output += "\n" + result.Output
					}
				}
				e.memory.Add("tool", fmt.Sprintf("[%s] %s", tc.Name, output))
			}
			continue
		}

		// No tool calls — LLM is done. Return the content.
		content := strings.TrimSpace(choice.Message.Content)
		if content != "" {
			e.memory.Add("assistant", content)
		}
		return content, nil
	}

	// If we exhausted steps, ask the LLM to summarize.
	resp, err := e.client.Chat(ctx, sysPrompt, append(e.memory.History(),
		llm.Message{Role: "user", Content: "Summarize what you've done so far in a few sentences."},
	), nil)
	if err == nil && len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "task did not complete within the maximum number of steps", nil
}

func (e *Executor) buildSystemPrompt() string {
	tmpl, err := template.New("system").Parse(prompts.SystemPrompt)
	if err != nil {
		return prompts.SystemPrompt
	}

	var buf bytes.Buffer
	_ = tmpl.Execute(&buf, map[string]string{
		"WorkingDir":  ".",
		"CurrentTime": time.Now().Format(time.RFC3339),
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
	})

	return buf.String()
}
