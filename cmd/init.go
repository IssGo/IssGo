package cmd

import (
	"fmt"
	"os"

	"github.com/issgo/issgo/internal/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a .issgo.yaml configuration file",
	Long: `Generate a default .issgo.yaml configuration file in the current directory.

If a config file already exists, use --force to overwrite it.`,
	RunE: runInit,
}

var forceOverwrite bool

func init() {
	initCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Overwrite existing config file")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	path := ".issgo.yaml"

	if utils.FileExists(path) && !forceOverwrite {
		return fmt.Errorf("%s already exists. Use --force to overwrite", path)
	}

	if err := os.WriteFile(path, []byte(defaultConfigYAML), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("✓ Created %s\n", path)
	fmt.Println("\nEdit this file to set your API key and preferences, then run:")
	fmt.Println("  issgo run \"your task here\"")
	return nil
}

const defaultConfigYAML = `# IssGo Configuration
# See https://github.com/issgo/issgo for full documentation.

llm:
  provider: deepseek
  model: deepseek-chat
  api_key: ""           # Set your API key here, or use ISSGO_LLM_API_KEY env var
  base_url: https://api.deepseek.com

tools:
  shell: true           # Enable shell command execution
  file: true            # Enable file operations
  web: true             # Enable HTTP requests
  browser: false        # Enable headless browser (requires Chrome)

agent:
  max_steps: 20         # Max tool-calling iterations per task
  allow_approve: true   # Prompt user before executing dangerous actions
  verbose: false        # Enable detailed debug output
`
