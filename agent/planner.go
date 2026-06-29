package agent

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/prompts"
)

// PlanStep represents a single step in the execution plan.
type PlanStep struct {
	Description string `json:"description"`
	Tool        string `json:"tool,omitempty"`
}

// Plan is a sequence of steps to accomplish a task.
type Plan struct {
	Goal  string     `json:"goal"`
	Steps []PlanStep `json:"steps"`
}

type Planner struct {
	client *llm.Client
	tools  []llm.ToolDef
}

func NewPlanner(client *llm.Client, tools []llm.ToolDef) *Planner {
	return &Planner{client: client, tools: tools}
}

func (p *Planner) CreatePlan(ctx context.Context, request string) (*Plan, error) {
	var toolNames []string
	for _, t := range p.tools {
		toolNames = append(toolNames, "- "+t.Name+": "+t.Description)
	}

	tmpl, err := template.New("planner").Parse(prompts.PlannerPrompt)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"Request": request,
		"Tools":   strings.Join(toolNames, "\n"),
	})
	if err != nil {
		return nil, err
	}

	history := []llm.Message{{Role: "user", Content: buf.String()}}
	resp, err := p.client.Chat(ctx, "", history, nil)
	if err != nil {
		return nil, err
	}

	plan := &Plan{Goal: request}
	if len(resp.Choices) > 0 {
		lines := strings.Split(strings.TrimSpace(resp.Choices[0].Message.Content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			plan.Steps = append(plan.Steps, PlanStep{Description: line})
		}
	}

	return plan, nil
}
