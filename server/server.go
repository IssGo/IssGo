// Package server provides an HTTP API for IssGo.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/issgo/issgo/config"
)

type Server struct {
	cfg    *config.Config
	agent  AgentRunner
	srv    *http.Server
}

// AgentRunner is the interface the server needs from the agent.
type AgentRunner interface {
	Run(ctx context.Context, task string) (string, error)
	ListTools() []string
}

func New(cfg *config.Config, agent AgentRunner) *Server {
	s := &Server{cfg: cfg, agent: agent}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/run", s.handleRun)
	mux.HandleFunc("/api/v1/stream", s.handleStream)
	mux.HandleFunc("/api/v1/tools", s.handleTools)

	var handler http.Handler = mux
	handler = LoggingMiddleware(handler)
	handler = CORSMiddleware(handler)
	if cfg.Server.AuthToken != "" {
		handler = AuthMiddleware(cfg.Server.AuthToken)(handler)
	}

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	fmt.Printf("IssGo API server listening on http://%s:%d\n", s.cfg.Server.Host, s.cfg.Server.Port)
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return fmt.Sprintf("http://%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
}
