package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/prompts"
	"github.com/issgo/issgo/tools"

	"github.com/issgo/issgo/internal/logger"
)

// ─── Executor ──────────────────────────────────────────────────

// ProgressEvent describes a single step in task execution.
type ProgressEvent struct {
	Step     int    `json:"step"`
	Action   string `json:"action"` // "tool_call", "tool_result", "done", "error", "max_steps"
	Tool     string `json:"tool,omitempty"`
	Details  string `json:"details,omitempty"`
	MaxSteps int    `json:"max_steps"`
}

// ProgressFunc is called at each execution step.
type ProgressFunc func(ProgressEvent)

type Executor struct {
	client   *llm.Client
	registry *tools.Registry
	memory   *Memory
	options  ExecutorOptions
}

type ExecutorOptions struct {
	MaxSteps     int
	MaxRetries   int
	Streaming    bool
	Verbose      bool
	AllowApprove bool
	Reflector    *Reflector
	Safety       *Safety
	OnProgress   ProgressFunc
}

func NewExecutor(client *llm.Client, registry *tools.Registry, memory *Memory, opts ExecutorOptions) *Executor {
	if opts.MaxSteps <= 0 {
		opts.MaxSteps = 30
	}
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}
	return &Executor{
		client:   client,
		registry: registry,
		memory:   memory,
		options:  opts,
	}
}

// Run executes the task using the ReAct pattern.
func (e *Executor) Run(ctx context.Context, task string) (string, error) {
	e.memory.Add("user", task)
	toolDefs := e.registry.List()
	sysPrompt := e.buildSystemPrompt()

	cyan := color.New(color.FgCyan).SprintFunc()
	if e.options.Verbose {
		fmt.Printf("\n%s\n", cyan("Available tools:"))
		for _, t := range toolDefs {
			fmt.Printf("  - %s: %s\n", t.Name, t.Description)
		}
	}

	for step := 0; step < e.options.MaxSteps; step++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		resp, err := e.client.Chat(ctx, sysPrompt, e.memory.History(), toolDefs)
		if err != nil {
			logger.Log.Errorw("chat error", "step", step, "error", err)
			// Retry
			if step < e.options.MaxSteps-1 {
				continue
			}
			return "", fmt.Errorf("step %d: %w", step, err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("step %d: empty response from LLM", step)
		}

		choice := resp.Choices[0]

		// Handle tool calls
		if len(choice.ToolCalls) > 0 {
			e.memory.AddAssistantToolCalls(choice.ToolCalls)

			if e.options.Verbose {
				for _, tc := range choice.ToolCalls {
					fmt.Printf("  [tool:%s] %s\n", tc.Name, tc.Arguments)
				}
			}

			for _, tc := range choice.ToolCalls {
				// Safety pre-check for shell commands BEFORE execution
				if tc.Name == "shell" && e.options.AllowApprove && e.options.Safety != nil {
					cmd, _ := tools.MustGetArg(json.RawMessage(tc.Arguments), "command")
					if cmd != "" {
						if blocked := e.options.Safety.EvaluateCommand(ctx, cmd, task); blocked {
							e.memory.AddToolResult(tc.Name, "BLOCKED: Command was evaluated as unsafe")
							if e.options.Verbose {
								fmt.Printf("  [result:%s] ✗ BLOCKED by safety check\n", tc.Name)
							}
							continue
						}
					}
				}

				result := e.registry.Execute(ctx, tc.Name, json.RawMessage(tc.Arguments))

				output := result.Output
				if !result.Success {
					output = "ERROR: " + result.Error
					if result.Output != "" {
						output += "\n" + result.Output
					}
				}

				if e.options.Verbose {
					if result.Success {
						fmt.Printf("  [result:%s] ✓ (%d chars)\n", tc.Name, len(result.Output))
					} else {
						fmt.Printf("  [result:%s] ✗ %s\n", tc.Name, result.Error)
					}
				}

				e.memory.AddToolResult(tc.Name, output)

				if e.options.OnProgress != nil {
					status := "ok"
					if !result.Success {
						status = "error"
					}
					e.options.OnProgress(ProgressEvent{
						Step:     step + 1,
						MaxSteps: e.options.MaxSteps,
						Action:   "tool_result",
						Tool:     tc.Name,
						Details:  status,
					})
				}
			}
			continue
		}

		// No tool calls — LLM is done
		content := strings.TrimSpace(choice.Message.Content)
		e.memory.AddAssistant(content)

		// Run reflector if enabled
		if e.options.Reflector != nil && !e.options.Reflector.HasRun() {
			e.options.Reflector.Reflect(ctx, task, e.memory)
			if e.options.Reflector.NeedsReplan() {
				content += "\n\n---\n⚠ Self-review suggests the plan may need adjustment: " + e.options.Reflector.Feedback()
			} else if e.options.Reflector.IsBlocked() {
				content += "\n\n---\n⛔ Self-review indicates this task may be blocked: " + e.options.Reflector.Feedback()
			}
		}

		return content, nil
	}

	// Exhausted steps — ask for summary
	logger.Log.Warnw("max steps reached", "max", e.options.MaxSteps)
	e.memory.Add("user", "You have reached the maximum number of steps. Please summarize what you've accomplished so far in a few sentences and explain what remains to be done.")

	resp, err := e.client.Chat(ctx, sysPrompt, e.memory.History(), nil)
	if err == nil && len(resp.Choices) > 0 {
		summary := strings.TrimSpace(resp.Choices[0].Message.Content)
		e.memory.AddAssistant(summary)
		return summary + "\n\n---\n⚠ Max steps reached. Task may be incomplete.", nil
	}

	return "⚠ Task did not complete within the maximum step limit.", nil
}

func (e *Executor) buildSystemPrompt() string {
	var toolDescs []string
	for _, t := range e.registry.List() {
		toolDescs = append(toolDescs, fmt.Sprintf("- **%s**: %s", t.Name, t.Description))
	}

	return prompts.RenderSystem(prompts.SystemVars{
		OS:               runtime.GOOS,
		Arch:             runtime.GOARCH,
		WorkingDir:       ".",
		CurrentTime:      time.Now().Format(time.RFC3339),
		Shell:            "bash",
		ToolsDescription: strings.Join(toolDescs, "\n"),
	})
}
