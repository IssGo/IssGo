package agent

import (
	"strings"

	"github.com/issgo/issgo/llm"
)

// Memory manages conversation history for the agent.
type Memory struct {
	messages []llm.Message
	maxTurns int
}

func NewMemory(maxTurns int) *Memory {
	if maxTurns <= 0 {
		maxTurns = 20
	}
	return &Memory{
		messages: make([]llm.Message, 0, maxTurns*2),
		maxTurns: maxTurns,
	}
}

func (m *Memory) Add(role, content string) {
	m.messages = append(m.messages, llm.Message{Role: role, Content: content})
	m.trim()
}

func (m *Memory) History() []llm.Message {
	return m.messages
}

func (m *Memory) Last() *llm.Message {
	if len(m.messages) == 0 {
		return nil
	}
	return &m.messages[len(m.messages)-1]
}

func (m *Memory) Summarize() string {
	var sb strings.Builder
	for _, msg := range m.messages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (m *Memory) Clear() {
	m.messages = m.messages[:0]
}

// trim keeps the conversation within maxTurns by removing oldest turns.
func (m *Memory) trim() {
	maxMsgs := m.maxTurns * 2
	for len(m.messages) > maxMsgs {
		m.messages = m.messages[2:]
	}
}
