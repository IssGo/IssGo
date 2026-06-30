package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/issgo/issgo/agent"
	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/internal/logger"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start interactive chat mode",
	Long: `Start an interactive chat session with the AI agent.

In chat mode, each line you type is sent as a task. The agent's memory
persists across messages, allowing multi-turn conversations.

Type /exit, /quit, or Ctrl+C to exit.
Type /clear to reset conversation memory.
Type /history to show conversation history.
Type /save <name> to save the session.
Type /load <name> to load a session.`,
	RunE: runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

func runChat(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	logger.Init(logger.Config{Verbose: cfg.Agent.Verbose})
	defer logger.Sync()

	if errs := config.Validate(cfg); len(errs) > 0 {
		return fmt.Errorf("%s", config.FormatErrors(errs))
	}

	ag := agent.New(cfg)
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("\n%s\n", bold("IssGo Chat Mode"))
	fmt.Print("Type your tasks or commands. /exit to quit.\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(cyan("> "))
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			switch {
			case input == "/exit" || input == "/quit":
				fmt.Println("Goodbye!")
				return nil
			case input == "/clear":
				ag.Memory().Clear()
				fmt.Println("Memory cleared.")
				continue
			case input == "/history":
				for _, m := range ag.Memory().History() {
					fmt.Printf("  [%s] %s\n", m.Role, m.Content)
				}
				continue
			default:
				fmt.Printf("Unknown command: %s\n", input)
				continue
			}
		}

		// Execute task
		result, err := ag.Run(cmd.Context(), input)
		if err != nil {
			fmt.Printf("  %s: %v\n", color.New(color.FgRed).Sprint("Error"), err)
			continue
		}
		fmt.Printf("\n%s\n\n", green(result))
	}

	return nil
}
