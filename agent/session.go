package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/issgo/issgo/internal/utils"
)

// ─── Session ───────────────────────────────────────────────────

type Session struct {
	ID        string            `json:"id"`
	Task      string            `json:"task"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Memory    *MemorySnapshot   `json:"memory"`
	Status    string            `json:"status"` // running, completed, failed, interrupted
	Result    string            `json:"result,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

type SessionManager struct {
	dir string
}

func NewSessionManager(dir string) *SessionManager {
	dir = utils.ExpandPath(dir)
	os.MkdirAll(dir, 0o700)
	return &SessionManager{dir: dir}
}

func (sm *SessionManager) Create(task string, memory *Memory) *Session {
	s := &Session{
		ID:        utils.RandomID(8),
		Task:      task,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Memory:    memory.ToSnapshot(),
		Status:    "running",
	}
	sm.Save(s)
	return s
}

func (sm *SessionManager) Save(s *Session) {
	s.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	if err := os.WriteFile(sm.path(s.ID), data, 0o600); err != nil {
		return
	}
}

func (sm *SessionManager) Load(id string) (*Session, error) {
	data, err := os.ReadFile(sm.path(id))
	if err != nil {
		return nil, fmt.Errorf("load session %s: %w", id, err)
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode session %s: %w", id, err)
	}
	return &s, nil
}

func (sm *SessionManager) List() ([]*Session, error) {
	entries, err := os.ReadDir(sm.dir)
	if err != nil {
		return nil, err
	}
	var sessions []*Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		s, err := sm.Load(strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (sm *SessionManager) Delete(id string) error {
	return os.Remove(sm.path(id))
}

func (sm *SessionManager) path(id string) string {
	return filepath.Join(sm.dir, id+".json")
}
