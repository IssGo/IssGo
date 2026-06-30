package cmd

import (
	"fmt"

	"github.com/issgo/issgo/internal/utils"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or edit configuration",
	Long:  `Show current configuration values or list available profiles.`,
	RunE: runConfig,
}

var configShowAll bool

func init() {
	configCmd.Flags().BoolVarP(&configShowAll, "all", "a", false, "Show full configuration")
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if configShowAll {
		fmt.Println(utils.PrettyJSON(cfg))
		return nil
	}

	fmt.Println("IssGo Configuration:")
	fmt.Println("─────────────────────")
	fmt.Printf("LLM Provider:  %s\n", cfg.LLM.Provider)
	fmt.Printf("Model:         %s\n", cfg.LLM.Model)
	fmt.Printf("Base URL:      %s\n", cfg.LLM.BaseURL)
	fmt.Printf("API Key:       %s\n", maskKey(cfg.LLM.APIKey))
	fmt.Printf("Temperature:   %.1f\n", cfg.LLM.Temperature)
	fmt.Printf("Max Tokens:    %d\n", cfg.LLM.MaxTokens)
	fmt.Println()
	fmt.Printf("Max Steps:     %d\n", cfg.Agent.MaxSteps)
	fmt.Printf("Verbose:       %v\n", cfg.Agent.Verbose)
	fmt.Printf("Approval:      %v\n", cfg.Agent.AllowApprove)
	fmt.Printf("Reflector:     %v\n", cfg.Agent.Reflector)
	fmt.Println()

	// Profiles
	if len(cfg.Profiles) > 0 {
		fmt.Println("Profiles:")
		for _, p := range cfg.Profiles {
			active := ""
			if p.Name == cfg.Active {
				active = " *active*"
			}
			fmt.Printf("  - %s (%s / %s)%s\n", p.Name, p.Provider, p.Model, active)
		}
		fmt.Println()
	}

	fmt.Printf("Enabled Tools: shell=%v file=%v web=%v browser=%v git=%v search=%v\n",
		cfg.Tools.Shell, cfg.Tools.File, cfg.Tools.Web, cfg.Tools.Browser, cfg.Tools.Git, cfg.Tools.Search)
	fmt.Printf("Config file:   %s\n", cfgFile)
	fmt.Printf("Profile:       %s\n", profile)

	return nil
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
