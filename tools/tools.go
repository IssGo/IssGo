package tools

import (
	"context"
	"encoding/json"

	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/internal/logger"
)

// Result holds the outcome of a tool execution.
type Result struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// Tool is the interface that every tool must implement.
type Tool interface {
	Name() string
	Description() string
	Schema() any
	Execute(ctx context.Context, args json.RawMessage) Result
}

// Registry holds all registered tools.
type Registry struct {
	tools map[string]Tool
}

func NewRegistry(cfg *config.Config) *Registry {
	r := &Registry{tools: make(map[string]Tool)}

	if cfg.Tools.File {
		r.Register(&FileTool{})
	}
	if cfg.Tools.Shell {
		r.Register(&ShellTool{})
	}
	if cfg.Tools.Web {
		r.Register(&WebTool{})
	}
	if cfg.Tools.Browser {
		r.Register(&BrowserTool{})
	}

	return r
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
	logger.Log.Debugw("registered tool", "name", t.Name())
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) List() []llm.ToolDef {
	defs := make([]llm.ToolDef, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, llm.ToolDef{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Schema(),
		})
	}
	return defs
}

func (r *Registry) Execute(ctx context.Context, name string, args json.RawMessage) Result {
	t, ok := r.Get(name)
	if !ok {
		return Result{Success: false, Error: "unknown tool: " + name}
	}
	logger.Log.Debugw("executing tool", "name", name, "args", string(args))
	return t.Execute(ctx, args)
}
