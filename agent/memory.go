// Package agent implements the AI agent loop, planning, memory, and execution.
package agent

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"

	"github.com/issgo/issgo/llm"
)

// ─── Memory ────────────────────────────────────────────────────

type Memory struct {
	mu          sync.RWMutex
	messages    []llm.Message
	maxTurns    int
	summary     string
	summarizeAt int // trigger summarization after this many turns
}

func NewMemory(maxTurns int) *Memory {
	if maxTurns <= 0 {
		maxTurns = 30
	}
	return &Memory{
		messages:    make([]llm.Message, 0, maxTurns*2),
		maxTurns:    maxTurns,
		summarizeAt: maxTurns / 2,
	}
}

func (m *Memory) Add(role, content string) {
	m.mu.Lock()
	m.messages = append(m.messages, llm.Message{Role: role, Content: content})
	m.trim()
	m.mu.Unlock()
}

func (m *Memory) AddToolResult(name, result string) {
	m.Add("tool", fmt.Sprintf("[%s] %s", name, result))
}

func (m *Memory) AddAssistantToolCalls(calls []llm.ToolCall) {
	m.mu.Lock()
	m.messages = append(m.messages, llm.AssistantToolCalls(calls))
	m.mu.Unlock()
}

func (m *Memory) AddAssistant(content string) {
	if content != "" {
		m.Add("assistant", content)
	}
}

func (m *Memory) History() []llm.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]llm.Message, len(m.messages))
	copy(out, m.messages)
	return out
}

func (m *Memory) Last() *llm.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.messages) == 0 {
		return nil
	}
	last := m.messages[len(m.messages)-1]
	return &last
}

func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}

func (m *Memory) Summary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.summary != "" {
		return m.summary
	}
	return m.summarize()
}

func (m *Memory) SetSummary(s string) {
	m.mu.Lock()
	m.summary = s
	m.mu.Unlock()
}

func (m *Memory) Clear() {
	m.mu.Lock()
	m.messages = m.messages[:0]
	m.summary = ""
	m.mu.Unlock()
}

func (m *Memory) Hash() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := sha256.New()
	for _, msg := range m.messages {
		fmt.Fprintf(h, "%s:%s\n", msg.Role, msg.Content)
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

func (m *Memory) TurnCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	turns := 0
	for _, msg := range m.messages {
		if msg.Role == "user" || msg.Role == "assistant" && len(msg.Content) > 0 {
			turns++
		}
	}
	return turns / 2
}

func (m *Memory) trim() {
	maxMsgs := m.maxTurns * 2
	if len(m.messages) <= maxMsgs {
		// Check if we should trigger summarization
		if m.summary == "" && len(m.messages) >= m.summarizeAt*2 {
			m.summary = m.summarize()
		}
		return
	}

	// Drop oldest turns and update summary
	drop := len(m.messages) - maxMsgs + 4
	if drop > 0 {
		m.summary = m.summarize()
		m.messages = m.messages[drop:]
	}
}

func (m *Memory) summarize() string {
	var sb strings.Builder
	for _, msg := range m.messages {
		if msg.Role == "assistant" && len(msg.Content) > 0 {
			sb.WriteString(msg.Content)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m *Memory) ToSnapshot() *MemorySnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	msgs := make([]llm.Message, len(m.messages))
	copy(msgs, m.messages)
	return &MemorySnapshot{
		Messages:    msgs,
		Summary:     m.summary,
		MaxTurns:    m.maxTurns,
		SummarizeAt: m.summarizeAt,
	}
}

func (m *Memory) FromSnapshot(s *MemorySnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]llm.Message, len(s.Messages))
	copy(m.messages, s.Messages)
	m.summary = s.Summary
	m.maxTurns = s.MaxTurns
	m.summarizeAt = s.SummarizeAt
}

type MemorySnapshot struct {
	Messages    []llm.Message `json:"messages"`
	Summary     string        `json:"summary"`
	MaxTurns    int           `json:"max_turns"`
	SummarizeAt int           `json:"summarize_at"`
}
