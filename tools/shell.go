package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ShellTool struct{}

func (s *ShellTool) Name() string        { return "shell" }
func (s *ShellTool) Description() string { return "Execute a shell command and return its output." }

func (s *ShellTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The shell command to execute.",
			},
			"working_dir": map[string]any{
				"type":        "string",
				"description": "Working directory for the command.",
			},
		},
		"required": []string{"command"},
	}
}

type shellArgs struct {
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir"`
}

func (s *ShellTool) Execute(ctx context.Context, raw json.RawMessage) Result {
	var args shellArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("invalid args: %v", err)}
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", args.Command)
	if args.WorkingDir != "" {
		cmd.Dir = args.WorkingDir
	}

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return Result{Success: false, Error: "command timed out after 60s"}
	}

	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Success: false, Output: output, Error: err.Error()}
	}

	if output == "" {
		output = "(no output)"
	}
	return Result{Success: true, Output: output}
}
