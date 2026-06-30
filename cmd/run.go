package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/issgo/issgo/agent"
	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/internal/logger"
	"github.com/issgo/issgo/internal/spinner"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [task description]",
	Short: "Execute an AI-powered task",
	Long: `Run an AI agent to autonomously complete a task described in natural language.

Examples:
  issgo run "List all Go files and count their lines"
  issgo run "Create a backup of all .yaml files"
  issgo run "Search for TODO comments in the codebase"
  issgo run "Analyze this directory and generate a summary report"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTask,
}

var (
	runVerbose bool
	runNoSpinner bool
)

func init() {
	runCmd.Flags().BoolVarP(&runVerbose, "verbose", "v", false, "Verbose output (show all LLM and tool interactions)")
	runCmd.Flags().BoolVar(&runNoSpinner, "no-spinner", false, "Disable the progress spinner")
	rootCmd.AddCommand(runCmd)
}

func runTask(cmd *cobra.Command, args []string) error {
	task := strings.Join(args, " ")

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if runVerbose {
		cfg.Agent.Verbose = true
	}

	logger.Init(logger.Config{Verbose: cfg.Agent.Verbose})
	defer logger.Sync()

	if errs := config.Validate(cfg); len(errs) > 0 {
		return fmt.Errorf("%s", config.FormatErrors(errs))
	}

	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("\n%s %s\n\n", bold("Task:"), cyan(task))

	ag := agent.New(cfg)

	var sp *spinner.Spinner
	if !runNoSpinner {
		sp = spinner.New("Thinking...")
		sp.Start()
	}

	ctx := context.Background()
	result, err := ag.RunWithSignalTrap(ctx, task)

	if sp != nil {
		if err != nil {
			sp.Fail("Failed")
		} else {
			sp.Success("Done")
		}
	}

	if err != nil {
		return fmt.Errorf("agent run: %w", err)
	}

	fmt.Printf("\n%s\n\n%s\n", green("Result:"), result)
	return nil
}
