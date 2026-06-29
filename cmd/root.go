package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "issgo",
	Short: "IssGo — AI Agent & Automation CLI",
	Long: `IssGo is an AI-powered CLI tool that lets you describe tasks in natural language
and executes them autonomously using LLM-powered planning and tool execution.

Commands:
  init      Generate a .issgo.yaml configuration file
  run       Execute an AI-powered task
  watch     Watch a directory and trigger on file changes

Examples:
  issgo init
  issgo run "Find all JSON files and merge them into one"
  issgo watch ./src --on-change "run tests"

Configuration:
  Set ISSGO_LLM_API_KEY in your environment or in .issgo.yaml.
`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
