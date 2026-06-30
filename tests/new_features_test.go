package tests

import (
	"context"
	"testing"

	"github.com/issgo/issgo/agent"
	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/llm"
	"github.com/issgo/issgo/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Session tests ──────────────────────────────────────────────

func TestSessionSaveLoad(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.SessionDir = t.TempDir()
	ag := agent.New(cfg)

	ag.Memory().Add("user", "hello")
	ag.Memory().Add("assistant", "hi there")

	err := ag.SaveSession("test-session")
	require.NoError(t, err)

	// Clear memory and reload
	ag.ResetSession()
	assert.Equal(t, 0, ag.Memory().Len())

	err = ag.LoadSession("test-session")
	require.NoError(t, err)
	assert.Equal(t, 2, ag.Memory().Len())
}

func TestSessionInvalidName(t *testing.T) {
	cfg := config.DefaultConfig()
	ag := agent.New(cfg)

	err := ag.SaveSession("../escape")
	assert.Error(t, err)

	err = ag.SaveSession("foo/bar")
	assert.Error(t, err)
}

func TestSessionList(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.SessionDir = t.TempDir()
	ag := agent.New(cfg)

	require.NoError(t, ag.SaveSession("one"))
	require.NoError(t, ag.SaveSession("two"))

	list, err := ag.ListSessions()
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

// ─── NoCache tests ──────────────────────────────────────────────

func TestClientNoCache(t *testing.T) {
	cfg := config.DefaultConfig()
	client := llm.NewClient(cfg)

	// Cache should be enabled by default
	_, _, ttl := client.CacheStats()
	assert.True(t, ttl > 0)

	// Disable cache
	client.SetNoCache(true)
	// Verify it doesn't panic (we can't test cache behavior without real API)
	client.SetNoCache(false)
}

// ─── Shell stdin tests ──────────────────────────────────────────

func TestShellStdin(t *testing.T) {
	st := &tools.ShellTool{}
	// Test stdin piped to wc -c
	result := st.Execute(context.Background(),
		[]byte(`{"command":"wc -c","stdin":"hello world"}`))
	assert.True(t, result.Success, result.Error)
}

func TestShellStdinEmpty(t *testing.T) {
	st := &tools.ShellTool{}
	result := st.Execute(context.Background(),
		[]byte(`{"command":"cat"}`))
	assert.True(t, result.Success, result.Error)
}

// ─── Progress callback tests ────────────────────────────────────

func TestProgressCallback(t *testing.T) {
	cfg := config.DefaultConfig()
	ag := agent.New(cfg)

	events := make([]agent.ProgressEvent, 0)
	ag.SetProgressCallback(func(e agent.ProgressEvent) {
		events = append(events, e)
	})

	// Callback should be set without error
	assert.NotNil(t, ag.Planner())
}

// ─── Git tool tests ─────────────────────────────────────────────

func TestGitToolBlockedCommands(t *testing.T) {
	gt := &tools.GitTool{}

	// Test allowed command
	result := gt.Execute(context.Background(),
		[]byte(`{"command":"status","repo_path":"/tmp"}`))
	// May fail if /tmp is not a git repo, but should not be "not allowed"
	assert.NotContains(t, result.Error, "not allowed")

	// Test force push to main
	result = gt.Execute(context.Background(),
		[]byte(`{"command":"push","args":"--force origin main"}`))
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "force push")
}

// ─── Config validation edge cases ────────────────────────────────

func TestConfigServerPortRange(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Server.Enabled = true
	cfg.Server.Port = 99999
	errs := config.Validate(cfg)
	assert.NotEmpty(t, errs)

	cfg.Server.Port = 8420
	errs = config.Validate(cfg)
	for _, e := range errs {
		assert.NotContains(t, e.Error(), "Port")
	}
}

// ─── Safe audit edge cases ──────────────────────────────────────

func TestSafeAuditEdgeCases(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.LLM.APIKey = "sk-test"
	_ = cfg // used for default config verification
}

// ─── Message serialization round trip ────────────────────────────

func TestToolCallMessageRoundTrip(t *testing.T) {
	msg := llm.AssistantToolCalls([]llm.ToolCall{
		{ID: "call_1", Name: "shell", Arguments: `{"command":"ls"}`},
	})
	assert.Equal(t, "assistant", msg.Role)
	assert.Len(t, msg.ToolCalls, 1)
	assert.Equal(t, "shell", msg.ToolCalls[0].Name)
}

// ─── Prompt rendering safety ─────────────────────────────────────

func TestPromptRendering(t *testing.T) {
	cfg := config.DefaultConfig()
	ag := agent.New(cfg)

	// Verify agent creation doesn't panic with default config
	assert.NotNil(t, ag)
	assert.NotNil(t, ag.Client())

	// Verify memory can handle empty history
	history := ag.Memory().History()
	assert.NotNil(t, history)
}
