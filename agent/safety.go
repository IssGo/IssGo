package agent

import (
	"context"
	"strings"

	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/prompts"

	"github.com/issgo/issgo/internal/safe"
)

// ─── Safety ────────────────────────────────────────────────────

type Safety struct {
	client   *llm.Client
	approval bool
}

func NewSafety(client *llm.Client, requireApproval bool) *Safety {
	return &Safety{client: client, approval: requireApproval}
}

// EvaluateCommand checks if a shell command is safe.
// Returns true if the command is BLOCKED.
func (s *Safety) EvaluateCommand(ctx context.Context, cmd string, request string) bool {
	// First, static analysis
	audit := safe.Audit(cmd)
	if audit.IsDangerous {
		return true // blocked
	}

	if !s.approval {
		return false
	}

	// If approval is required and command is risky, ask LLM
	if audit.IsWarning {
		return s.askLLM(ctx, cmd, request)
	}

	return false
}

func (s *Safety) askLLM(ctx context.Context, cmd, request string) bool {
	prompt := prompts.RenderSafety(prompts.SafetyVars{
		Command:    cmd,
		WorkingDir: ".",
		Request:    request,
	})

	history := []llm.Message{llm.UserMsg(prompt)}
	resp, err := s.client.Chat(ctx, "", history, nil)
	if err != nil {
		// Err on the side of caution
		return true
	}

	if len(resp.Choices) == 0 {
		return true
	}

	verdict := strings.ToUpper(strings.TrimSpace(resp.Choices[0].Message.Content))
	return strings.HasPrefix(verdict, "BLOCK")
}

// StaticCheck performs a fast static safety check (no LLM).
func StaticCheck(cmd string) bool {
	return safe.Audit(cmd).IsDangerous
}
