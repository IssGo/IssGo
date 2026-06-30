package cmd

import (
	"fmt"
	"os"

	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/internal/utils"
	"github.com/spf13/cobra"
)

var (
	forceOverwrite bool
	forceOutput    string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a .issgo.yaml configuration file",
	Long: `Generate a default .issgo.yaml configuration file in the current directory.

If a config file already exists, use --force to overwrite it.
Use --output to specify a custom path.`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Overwrite existing config file")
	initCmd.Flags().StringVarP(&forceOutput, "output", "o", ".issgo.yaml", "Output path for config file")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	path := forceOutput

	if utils.FileExists(path) && !forceOverwrite {
		return fmt.Errorf("%s already exists. Use --force to overwrite", path)
	}

	if err := config.WriteDefault(path); err != nil {
		// Fallback: write embedded template
		if err := os.WriteFile(path, []byte(defaultConfigTemplate), 0o644); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}

	fmt.Printf("✓ Created %s\n", path)
	fmt.Println("\nEdit this file to set your API key, then run:")
	fmt.Println("  issgo run \"your task here\"")
	fmt.Println("\nOr set via environment variables:")
	fmt.Println("  export ISSGO_LLM_API_KEY=sk-xxx")
	return nil
}

var defaultConfigTemplate = `# IssGo Configuration v2026.06.30
# Docs: https://github.com/issgo/issgo

llm:
  provider: deepseek            # deepseek | openai | ollama | custom
  model: deepseek-chat          # model name
  api_key: ""                   # API key (or use ISSGO_LLM_API_KEY env var)
  base_url: https://api.deepseek.com
  temperature: 0.7
  max_tokens: 4096
  timeout_secs: 120
  retry_count: 3

tools:
  shell: true                   # bash command execution
  file: true                    # file read/write/list/delete
  web: true                     # HTTP requests
  browser: false                # headless Chrome (requires chromium)
  git: true                     # git operations
  search: true                  # regex file search
  plugins: false                # third-party plugin scripts
  plugins_dir: ~/.issgo/plugins

agent:
  max_steps: 30                 # max tool-calling iterations per task
  allow_approve: true           # confirm before dangerous actions
  verbose: false                # detailed debug output
  streaming: true               # stream LLM responses
  reflector: true               # self-reflection after each task
  max_retries: 3
  session_dir: ~/.issgo/sessions

server:
  enabled: false
  host: 127.0.0.1
  port: 8420

# Profiles for quick switching
profiles:
  - name: deepseek
    provider: deepseek
    model: deepseek-chat
    base_url: https://api.deepseek.com
  - name: openai
    provider: openai
    model: gpt-4o
    base_url: https://api.openai.com/v1
  - name: ollama
    provider: ollama
    model: qwen2.5:14b
    base_url: http://localhost:11434

active: ""                      # which profile to use (empty = default config)
`
