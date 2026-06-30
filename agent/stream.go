package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/tools"
)

// ─── StreamExecutor ────────────────────────────────────────────

// StreamExecutor runs the ReAct loop with streaming responses.
type StreamExecutor struct {
	*Executor
}

func NewStreamExecutor(client *llm.Client, registry *tools.Registry, memory *Memory, opts ExecutorOptions) *StreamExecutor {
	return &StreamExecutor{Executor: NewExecutor(client, registry, memory, opts)}
}

// StreamChunk is emitted during execution.
type StreamChunk struct {
	Type    string // "text", "tool_call", "tool_result", "done", "error"
	Content string
	Tool    string
	Args    string
	Result  tools.Result
}

// RunStream runs the task and sends progress via channel.
func (se *StreamExecutor) RunStream(ctx context.Context, task string) <-chan StreamChunk {
	ch := make(chan StreamChunk, 50)

	go func() {
		defer close(ch)

		se.memory.Add("user", task)
		toolDefs := se.registry.List()
		sysPrompt := se.buildSystemPrompt()

		for step := 0; step < se.options.MaxSteps; step++ {
			select {
			case <-ctx.Done():
				ch <- StreamChunk{Type: "error", Content: ctx.Err().Error()}
				return
			default:
			}

			resp, err := se.client.Chat(ctx, sysPrompt, se.memory.History(), toolDefs)
			if err != nil {
				ch <- StreamChunk{Type: "error", Content: fmt.Sprintf("step %d: %v", step, err)}
				return
			}

			if len(resp.Choices) == 0 {
				ch <- StreamChunk{Type: "error", Content: "empty LLM response"}
				return
			}

			choice := resp.Choices[0]

			// Tool calls
			if len(choice.ToolCalls) > 0 {
				se.memory.AddAssistantToolCalls(choice.ToolCalls)

				for _, tc := range choice.ToolCalls {
					ch <- StreamChunk{
						Type: "tool_call",
						Tool: tc.Name,
						Args: tc.Arguments,
					}

					result := se.registry.Execute(ctx, tc.Name, []byte(tc.Arguments))
					output := result.Output
					if !result.Success {
						output = "ERROR: " + result.Error
						if result.Output != "" {
							output += "\n" + result.Output
						}
					}

					se.memory.AddToolResult(tc.Name, output)
					ch <- StreamChunk{
						Type:   "tool_result",
						Tool:   tc.Name,
						Result: result,
					}

					// Safety check on shell commands
					if tc.Name == "shell" && se.options.AllowApprove {
						cmd, _ := tools.MustGetArg(json.RawMessage(tc.Arguments), "command")
						if cmd != "" && se.options.Safety != nil {
							if blocked := se.options.Safety.EvaluateCommand(ctx, cmd, task); blocked {
								se.memory.AddToolResult(tc.Name, "BLOCKED: Command was evaluated as unsafe")
							}
						}
					}
				}
				continue
			}

			// Done
			content := strings.TrimSpace(choice.Message.Content)
			se.memory.AddAssistant(content)
			ch <- StreamChunk{Type: "text", Content: content}
			ch <- StreamChunk{Type: "done"}
			return
		}

		ch <- StreamChunk{Type: "error", Content: "max steps reached"}
	}()

	return ch
}
