package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/issgo/issgo/agent"
	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/internal/logger"
	"github.com/issgo/issgo/server"
	"github.com/spf13/cobra"
)

var (
	servePort int
	serveHost string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP API server",
	Long: `Start an HTTP API server that exposes IssGo functionality via REST endpoints.

Endpoints:
  GET  /api/v1/health   — Health check
  POST /api/v1/run      — Execute a task (body: {"task":"..."})
  POST /api/v1/stream   — Execute with SSE streaming
  GET  /api/v1/tools    — List available tools

Examples:
  issgo serve
  issgo serve --port 8080
  issgo serve --host 0.0.0.0 --port 8420`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 0, "Port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "", "Host to bind to")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	logger.Init(logger.Config{Verbose: cfg.Agent.Verbose})
	defer logger.Sync()

	if servePort > 0 {
		cfg.Server.Port = servePort
	}
	if serveHost != "" {
		cfg.Server.Host = serveHost
	}

	cfg.Server.Enabled = true

	if errs := config.Validate(cfg); len(errs) > 0 {
		return fmt.Errorf(config.FormatErrors(errs))
	}

	ag := agent.New(cfg)
	srv := server.New(cfg, ag)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down server...")
		srv.Shutdown(cmd.Context())
	}()

	fmt.Printf("IssGo API Server v%s\n", Version)
	return srv.Start()
}
