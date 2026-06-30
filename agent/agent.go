package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/tools"

	"github.com/issgo/issgo/internal/logger"
)

// ─── Agent ─────────────────────────────────────────────────────

type Agent struct {
	cfg       *config.Config
	client    *llm.Client
	registry  *tools.Registry
	memory    *Memory
	executor  *Executor
	planner   *Planner
	reflector *Reflector
	safety    *Safety
}

func New(cfg *config.Config) *Agent {
	client := llm.NewClient(cfg)
	registry := tools.NewRegistry(cfg)
	memory := NewMemory(cfg.Agent.MaxSteps)

	var reflector *Reflector
	if cfg.Agent.Reflector {
		reflector = NewReflector(client)
	}

	safety := NewSafety(client, cfg.Agent.AllowApprove)

	execOpts := ExecutorOptions{
		MaxSteps:     cfg.Agent.MaxSteps,
		MaxRetries:   cfg.Agent.MaxRetries,
		Streaming:    cfg.Agent.Streaming,
		Verbose:      cfg.Agent.Verbose,
		AllowApprove: cfg.Agent.AllowApprove,
		Reflector:    reflector,
		Safety:       safety,
	}

	executor := NewExecutor(client, registry, memory, execOpts)
	planner := NewPlanner(client, registry.List())

	return &Agent{
		cfg:       cfg,
		client:    client,
		registry:  registry,
		memory:    memory,
		executor:  executor,
		planner:   planner,
		reflector: reflector,
		safety:    safety,
	}
}

// Run executes a task and returns the final response.
func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	logger.Log.Infow("agent: starting task", "task", task)

	result, err := a.executor.Run(ctx, task)
	if err != nil {
		logger.Log.Errorw("agent: task failed", "task", task, "error", err)
		return "", fmt.Errorf("agent: %w", err)
	}

	logger.Log.Infow("agent: task completed", "task", task, "result_len", len(result))
	return result, nil
}

// RunWithSignalTrap wraps Run with SIGINT handling.
func (a *Agent) RunWithSignalTrap(ctx context.Context, task string) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nInterrupted. Shutting down gracefully...")
		cancel()
	}()

	return a.Run(ctx, task)
}

// ─── Accessors ─────────────────────────────────────────────────

func (a *Agent) Memory() *Memory         { return a.memory }
func (a *Agent) Client() *llm.Client     { return a.client }
func (a *Agent) Registry() *tools.Registry { return a.registry }
func (a *Agent) Config() *config.Config  { return a.cfg }
func (a *Agent) Planner() *Planner       { return a.planner }
