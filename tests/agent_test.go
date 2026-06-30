package tests

import (
	"context"
	"testing"

	"github.com/issgo/issgo/agent"
	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/tools"
	"github.com/stretchr/testify/assert"
)

// ─── Agent Tests ───────────────────────────────────────────────

func TestAgentCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	ag := agent.New(cfg)
	assert.NotNil(t, ag)
	assert.NotNil(t, ag.Memory())
	assert.NotNil(t, ag.Client())
	assert.NotNil(t, ag.Config())
}

func TestMemoryOperations(t *testing.T) {
	m := agent.NewMemory(10)
	assert.Equal(t, 0, m.Len())

	m.Add("user", "hello")
	m.Add("assistant", "hi there")
	assert.Equal(t, 2, m.Len())

	last := m.Last()
	assert.NotNil(t, last)
	assert.Equal(t, "assistant", last.Role)

	history := m.History()
	assert.Len(t, history, 2)
}

func TestMemoryTrimming(t *testing.T) {
	m := agent.NewMemory(2) // only 2 turns = 4 messages
	for i := 0; i < 10; i++ {
		m.Add("user", "msg")
		m.Add("assistant", "reply")
	}
	assert.LessOrEqual(t, m.Len(), 8)
}

func TestMemoryClear(t *testing.T) {
	m := agent.NewMemory(10)
	m.Add("user", "test")
	m.Clear()
	assert.Equal(t, 0, m.Len())
}

func TestMemorySnapshot(t *testing.T) {
	m := agent.NewMemory(10)
	m.Add("user", "task")
	m.Add("assistant", "result")

	snap := m.ToSnapshot()
	assert.Equal(t, 2, len(snap.Messages))

	m2 := agent.NewMemory(10)
	m2.FromSnapshot(snap)
	assert.Equal(t, 2, m2.Len())
}

func TestPlannerCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	client := llm.NewClient(cfg)
	defs := []llm.ToolDef{
		{Name: "shell", Description: "run commands"},
	}
	planner := agent.NewPlanner(client, defs)
	assert.NotNil(t, planner)
}

// ─── Tool Tests ────────────────────────────────────────────────

func TestRegistryCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	reg := tools.NewRegistry(cfg)
	assert.NotNil(t, reg)
	assert.True(t, reg.Count() >= 2, "at least file+shell tools should be registered")
}

func TestRegistryList(t *testing.T) {
	cfg := config.DefaultConfig()
	reg := tools.NewRegistry(cfg)
	defs := reg.List()
	assert.NotEmpty(t, defs)
}

func TestToolExecution(t *testing.T) {
	ft := &tools.FileTool{}
	result := ft.Execute(context.Background(), []byte(`{"action":"exists","path":"/tmp"}`))
	assert.True(t, result.Success)
}

func TestShellSafety(t *testing.T) {
	st := &tools.ShellTool{}
	// rm -rf / should be blocked
	result := st.Execute(context.Background(), []byte(`{"command":"rm -rf /"}`))
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "dangerous")
}

func TestToolNotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	reg := tools.NewRegistry(cfg)
	result := reg.Execute(context.Background(), "nonexistent", []byte(`{}`))
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "unknown tool")
}

// ─── Config Tests ──────────────────────────────────────────────

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	assert.Equal(t, "deepseek", cfg.LLM.Provider)
	assert.Equal(t, "deepseek-chat", cfg.LLM.Model)
	assert.True(t, cfg.Tools.Shell)
	assert.False(t, cfg.Tools.Browser)
	assert.True(t, cfg.Agent.Reflector)
}

func TestConfigValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.LLM.APIKey = ""
	errs := config.Validate(cfg)
	assert.NotEmpty(t, errs)

	cfg.LLM.APIKey = "sk-test"
	cfg.LLM.Temperature = 5.0
	errs = config.Validate(cfg)
	assert.NotEmpty(t, errs)
}

func TestProfileAPI(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Profiles = []config.Profile{
		{Name: "test", Provider: "openai", Model: "gpt-4"},
	}
	p, err := config.GetProfile(cfg, "test")
	assert.NoError(t, err)
	assert.Equal(t, "openai", p.Provider)

	_, err = config.GetProfile(cfg, "missing")
	assert.Error(t, err)
}

// ─── LLM Type Tests ────────────────────────────────────────────

func TestMessageConstructors(t *testing.T) {
	u := llm.UserMsg("hi")
	assert.Equal(t, "user", u.Role)
	assert.Equal(t, "hi", u.Content)

	sys := llm.SystemMsg("system prompt")
	assert.Equal(t, "system", sys.Role)

	assistant := llm.AssistantMsg("response")
	assert.Equal(t, "assistant", assistant.Role)
}

// ─── Safe Tests ────────────────────────────────────────────────

func TestSafeAudit(t *testing.T) {
	// This would need importing internal/safe
	// Skip for now
}
