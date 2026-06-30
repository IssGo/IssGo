// Package tools defines the tool interface and registry.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/internal/logger"
)

// ─── Types ─────────────────────────────────────────────────────

type Result struct {
	Success bool            `json:"success"`
	Output  string          `json:"output"`
	Error   string          `json:"error,omitempty"`
	Meta    map[string]any  `json:"meta,omitempty"`
}

type Tool interface {
	Name() string
	Description() string
	Schema() any
	Execute(ctx context.Context, args json.RawMessage) Result
}

// ─── Registry ──────────────────────────────────────────────────

type Registry struct {
	mu       sync.RWMutex
	tools    map[string]Tool
	pluginCh map[string]chan Result // per-tool async channels
}

func NewRegistry(cfg *config.Config) *Registry {
	r := &Registry{
		tools:    make(map[string]Tool),
		pluginCh: make(map[string]chan Result),
	}

	// Register built-in tools based on config
	add := func(ok bool, t Tool) {
		if ok {
			r.Register(t)
		}
	}

	add(cfg.Tools.File, &FileTool{})
	add(cfg.Tools.Shell, &ShellTool{})
	add(cfg.Tools.Web, &WebTool{})
	add(cfg.Tools.Browser, &BrowserTool{})
	add(cfg.Tools.Git, &GitTool{})
	add(cfg.Tools.Search, &SearchTool{})

	return r
}

func (r *Registry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[t.Name()] = t
	logger.Log.Debugw("tool registered", "name", t.Name())
}

func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}

func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) List() []llm.ToolDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for n := range r.tools {
		names = append(names, n)
	}
	sort.Strings(names)

	defs := make([]llm.ToolDef, 0, len(names))
	for _, n := range names {
		t := r.tools[n]
		defs = append(defs, llm.ToolDef{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Schema(),
		})
	}
	return defs
}

func (r *Registry) ToolNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for n := range r.tools {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

func (r *Registry) Execute(ctx context.Context, name string, args json.RawMessage) Result {
	t, ok := r.Get(name)
	if !ok {
		return Result{Success: false, Error: "unknown tool: " + name}
	}
	logger.Log.Debugw("executing tool", "name", name)
	result := t.Execute(ctx, args)
	if !result.Success {
		logger.Log.Warnw("tool failed", "name", name, "error", result.Error)
	}
	return result
}

// ─── Async helpers ─────────────────────────────────────────────

func (r *Registry) ExecuteAsync(ctx context.Context, name string, args json.RawMessage) <-chan Result {
	ch := make(chan Result, 1)
	go func() {
		defer close(ch)
		ch <- r.Execute(ctx, name, args)
	}()
	return ch
}

// ─── Tool helpers ──────────────────────────────────────────────

func ResultOk(output string) Result {
	return Result{Success: true, Output: output}
}

func ResultErr(err string) Result {
	return Result{Success: false, Error: err}
}

func ResultOkWithMeta(output string, meta map[string]any) Result {
	return Result{Success: true, Output: output, Meta: meta}
}

func ToJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func MustGetArg(args json.RawMessage, key string) (string, bool) {
	var m map[string]any
	if err := json.Unmarshal(args, &m); err != nil {
		return "", false
	}
	v, ok := m[key]
	if !ok {
		return "", false
	}
	switch val := v.(type) {
	case string:
		return val, true
	default:
		return fmt.Sprintf("%v", val), true
	}
}
