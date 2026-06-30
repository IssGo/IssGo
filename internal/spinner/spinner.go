// Package spinner provides a terminal spinner for long-running operations.
package spinner

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	mu      sync.Mutex
	msg     string
	active  bool
	stopCh  chan struct{}
	stopped chan struct{}
}

func New(msg string) *Spinner {
	return &Spinner{msg: msg}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		return
	}
	s.active = true
	s.stopCh = make(chan struct{})
	s.stopped = make(chan struct{})
	go s.run()
}

func (s *Spinner) run() {
	tick := time.NewTicker(80 * time.Millisecond)
	defer tick.Stop()
	defer close(s.stopped)
	i := 0
	for {
		select {
		case <-s.stopCh:
			fmt.Fprintf(os.Stderr, "\r\033[K")
			return
		case <-tick.C:
			s.mu.Lock()
			fmt.Fprintf(os.Stderr, "\r\033[K%s %s", frames[i%len(frames)], s.msg)
			s.mu.Unlock()
			i++
		}
	}
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		return
	}
	close(s.stopCh)
	<-s.stopped
	s.active = false
}

func (s *Spinner) Update(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

func (s *Spinner) Success(msg string) {
	s.Stop()
	fmt.Fprintf(os.Stderr, "\r\033[K✓ %s\n", msg)
}

func (s *Spinner) Fail(msg string) {
	s.Stop()
	fmt.Fprintf(os.Stderr, "\r\033[K✗ %s\n", msg)
}
