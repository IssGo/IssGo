package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/issgo/issgo/internal/safe"
)

type ShellTool struct{}

func (s *ShellTool) Name() string        { return "shell" }
func (s *ShellTool) Description() string { return "Execute shell commands with a 120s timeout. Returns stdout+stderr." }

func (s *ShellTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command":     map[string]any{"type": "string", "description": "Shell command to execute."},
			"working_dir": map[string]any{"type": "string", "description": "Working directory."},
			"env":         map[string]any{"type": "object", "description": "Environment variables."},
			"timeout_sec": map[string]any{"type": "number", "description": "Override default timeout (max 120)."},
		},
		"required": []string{"command"},
	}
}

type shellArgs struct {
	Command    string            `json:"command"`
	WorkingDir string            `json:"working_dir"`
	Env        map[string]string `json:"env"`
	TimeoutSec float64           `json:"timeout_sec"`
}

func (s *ShellTool) Execute(ctx context.Context, raw json.RawMessage) Result {
	var args shellArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return ResultErr("invalid args: " + err.Error())
	}

	// Safety check
	audit := safe.Audit(args.Command)
	if audit.IsDangerous {
		return ResultErr("dangerous command blocked: matches pattern " + audit.Pattern)
	}

	timeout := 120 * time.Second
	if args.TimeoutSec > 0 && args.TimeoutSec <= 120 {
		timeout = time.Duration(args.TimeoutSec) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", args.Command)
	if args.WorkingDir != "" {
		cmd.Dir = args.WorkingDir
	}
	if len(args.Env) > 0 {
		for k, v := range args.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return Result{Success: false, Output: string(out), Error: "command timed out"}
	}

	output := string(out)
	// Truncate excessively long output
	if len(output) > 50000 {
		output = output[:50000] + "\n... (truncated)"
	}

	// Sanitize non-printable chars
	output = sanitizeOutput(output)

	if err != nil {
		return Result{Success: false, Output: output, Error: err.Error()}
	}
	if output == "" {
		output = "(no output)"
	}
	return ResultOk(output)
}

func sanitizeOutput(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == utf8.RuneError || (r < 32 && r != '\n' && r != '\r' && r != '\t') {
			b.WriteString(fmt.Sprintf("\\x%02x", r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
