package agent

import (
	"context"
	"fmt"

	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/tools"

	"github.com/issgo/issgo/internal/logger"
)

// Agent ties together the planner, executor, memory, and tools.
type Agent struct {
	cfg      *config.Config
	client   *llm.Client
	registry *tools.Registry
	memory   *Memory
	executor *Executor
}

func New(cfg *config.Config) *Agent {
	client := llm.NewClient(cfg)
	registry := tools.NewRegistry(cfg)
	memory := NewMemory(cfg.Agent.MaxSteps)
	executor := NewExecutor(client, registry, memory)

	return &Agent{
		cfg:      cfg,
		client:   client,
		registry: registry,
		memory:   memory,
		executor: executor,
	}
}

// Run executes a user task and returns the final response.
func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	logger.Log.Infow("starting task", "task", task)

	result, err := a.executor.Run(ctx, task, a.cfg.Agent.MaxSteps)
	if err != nil {
		logger.Log.Errorw("task failed", "task", task, "error", err)
		return "", fmt.Errorf("agent execution: %w", err)
	}

	logger.Log.Infow("task completed", "task", task)
	return result, nil
}
