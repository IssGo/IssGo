package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/prompts"
)

// ─── Planner ───────────────────────────────────────────────────

type PlanStep struct {
	Number      int    `json:"number"`
	Description string `json:"description"`
	Tool        string `json:"tool,omitempty"`
	Args        string `json:"args,omitempty"`
	Status      string `json:"status"` // pending, active, done, failed, skipped
}

type Plan struct {
	Goal      string     `json:"goal"`
	Steps     []PlanStep `json:"steps"`
	CreatedAt string     `json:"created_at"`
}

type Planner struct {
	client *llm.Client
	tools  []llm.ToolDef
}

func NewPlanner(client *llm.Client, toolDefs []llm.ToolDef) *Planner {
	return &Planner{client: client, tools: toolDefs}
}

func (p *Planner) CreatePlan(ctx context.Context, request string) (*Plan, error) {
	var toolNames []string
	for _, t := range p.tools {
		toolNames = append(toolNames, fmt.Sprintf("- **%s**: %s", t.Name, t.Description))
	}

	planPrompt := prompts.RenderPlanner(prompts.PlannerVars{
		Tools:   strings.Join(toolNames, "\n"),
		Request: request,
	})

	history := []llm.Message{llm.UserMsg(planPrompt)}
	resp, err := p.client.Chat(ctx, "", history, nil)
	if err != nil {
		return nil, fmt.Errorf("planner: %w", err)
	}

	plan := &Plan{Goal: request}
	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		lines := strings.Split(strings.TrimSpace(content), "\n")
		stepNum := 1
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Strip leading numbering
			for strings.HasPrefix(line, fmt.Sprintf("%d.", stepNum)) ||
				strings.HasPrefix(line, fmt.Sprintf("%d)", stepNum)) ||
				strings.HasPrefix(line, fmt.Sprintf("%d ", stepNum)) {
				idx := strings.Index(line, " ")
				if idx > 0 {
					line = strings.TrimSpace(line[idx+1:])
				}
				break
			}
			plan.Steps = append(plan.Steps, PlanStep{
				Number:      stepNum,
				Description: line,
				Status:      "pending",
			})
			stepNum++
		}
	}

	// If LLM returned no structured steps, create a single-step plan
	if len(plan.Steps) == 0 {
		plan.Steps = append(plan.Steps, PlanStep{
			Number:      1,
			Description: request,
			Status:      "pending",
		})
	}

	return plan, nil
}

func (p *Plan) CurrentStep() *PlanStep {
	for i := range p.Steps {
		if p.Steps[i].Status == "pending" || p.Steps[i].Status == "active" {
			return &p.Steps[i]
		}
	}
	return nil
}

func (p *Plan) MarkDone(stepNum int) {
	for i := range p.Steps {
		if p.Steps[i].Number == stepNum {
			p.Steps[i].Status = "done"
			return
		}
	}
}

func (p *Plan) MarkFailed(stepNum int) {
	for i := range p.Steps {
		if p.Steps[i].Number == stepNum {
			p.Steps[i].Status = "failed"
			return
		}
	}
}

func (p *Plan) IsComplete() bool {
	for _, s := range p.Steps {
		if s.Status == "pending" || s.Status == "active" {
			return false
		}
	}
	return len(p.Steps) > 0
}

func (p *Plan) Format() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Goal: %s\n\n", p.Goal))
	for _, s := range p.Steps {
		icon := "  "
		switch s.Status {
		case "done":
			icon = "✓ "
		case "failed":
			icon = "✗ "
		case "active":
			icon = "→ "
		}
		sb.WriteString(fmt.Sprintf("%s%d. %s\n", icon, s.Number, s.Description))
	}
	return sb.String()
}
