package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/prompts"

	"github.com/issgo/issgo/internal/logger"
)

// ─── Reflector ─────────────────────────────────────────────────

type Reflector struct {
	client   *llm.Client
	ran      bool
	verdict  string // CONTINUE, REPLAN, COMPLETE, BLOCKED
	feedback string
}

func NewReflector(client *llm.Client) *Reflector {
	return &Reflector{client: client}
}

func (r *Reflector) HasRun() bool { return r.ran }

func (r *Reflector) Verdict() string { return r.verdict }

func (r *Reflector) Feedback() string { return r.feedback }

// Reflect analyzes the agent's recent actions and decides if replanning is needed.
func (r *Reflector) Reflect(ctx context.Context, task string, memory *Memory) {
	r.ran = true

	actions := memory.Summary()
	if actions == "" {
		r.verdict = "CONTINUE"
		return
	}

	prompt := prompts.RenderReflector(prompts.ReflectorVars{
		Task:    task,
		Actions: actions,
	})

	history := []llm.Message{llm.UserMsg(prompt)}
	resp, err := r.client.Chat(ctx, "", history, nil)
	if err != nil {
		logger.Log.Warnw("reflector: chat failed", "error", err)
		r.verdict = "CONTINUE"
		return
	}

	if len(resp.Choices) == 0 {
		r.verdict = "CONTINUE"
		return
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	upper := strings.ToUpper(content)

	switch {
	case strings.Contains(upper, "CONTINUE"):
		r.verdict = "CONTINUE"
	case strings.Contains(upper, "REPLAN"):
		r.verdict = "REPLAN"
	case strings.Contains(upper, "COMPLETE"):
		r.verdict = "COMPLETE"
	case strings.Contains(upper, "BLOCKED"):
		r.verdict = "BLOCKED"
	default:
		r.verdict = "CONTINUE"
	}

	r.feedback = content
	logger.Log.Infow("reflector verdict", "verdict", r.verdict)
}

// NeedsReplan returns true if the reflector thinks the approach is wrong.
func (r *Reflector) NeedsReplan() bool { return r.verdict == "REPLAN" }

// IsBlocked returns true if the task cannot proceed.
func (r *Reflector) IsBlocked() bool { return r.verdict == "BLOCKED" }

func (r *Reflector) String() string {
	return fmt.Sprintf("Reflector(%s): %s", r.verdict, r.feedback)
}
