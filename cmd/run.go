package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/issgo/issgo/agent"
	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/internal/logger"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [task description]",
	Short: "Execute an AI-powered task",
	Long: `Run an AI agent to autonomously complete a task described in natural language.

Examples:
  issgo run "List all Go files and count their lines"
  issgo run "Create a new directory called backups and copy all .yaml files into it"
  issgo run "Search for TODO comments in the codebase"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTask,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runTask(cmd *cobra.Command, args []string) error {
	task := strings.Join(args, " ")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger.Init(cfg.Agent.Verbose)
	defer logger.Sync()

	if cfg.LLM.APIKey == "" {
		return fmt.Errorf(
			"no API key configured.\n\nSet it in .issgo.yaml or via environment variable:\n  export ISSGO_LLM_API_KEY=your-key-here",
		)
	}

	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("\n%s %s\n\n", bold("Task:"), cyan(task))

	ag := agent.New(cfg)

	// Trap Ctrl+C for graceful cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nInterrupted. Shutting down...")
		cancel()
	}()

	result, err := ag.Run(ctx, task)
	if err != nil {
		return fmt.Errorf("agent run: %w", err)
	}

	fmt.Printf("\n%s\n\n%s\n", bold("Result:"), result)
	return nil
}
