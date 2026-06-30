package cmd

import (
	"fmt"
	"os"

	"github.com/issgo/issgo/config"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via -ldflags.
	Version   = "dev"
	BuildTime = "unknown"

	cfgFile string
	profile string
)

var rootCmd = &cobra.Command{
	Use:   "issgo",
	Short: "IssGo — AI Agent & Automation CLI",
	Long: `IssGo is an AI-powered CLI tool that lets you describe tasks in
natural language and executes them autonomously.

Commands:
  init      Generate configuration file
  run       Execute an AI-powered task
  chat      Start interactive chat mode
  watch     Watch directory and trigger on changes
  config    Show or edit configuration
  serve     Start HTTP API server
  version   Show version information

Examples:
  issgo init
  issgo run "List all Go files and count their lines"
  issgo chat
  issgo watch ./src --on-change "format all files"
  issgo serve --port 8420

Configuration:
  Set ISSGO_LLM_API_KEY in your environment or in .issgo.yaml.
  Profiles allow quick switching between LLM providers.
`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "use a named profile from config")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig() (*config.Config, error) {
	if cfgFile != "" {
		return config.LoadWithPath(cfgFile)
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if profile != "" {
		if p, err := config.GetProfile(cfg, profile); err == nil {
			cfg = config.ProfileToConfig(*p)
		}
	}
	return cfg, nil
}
