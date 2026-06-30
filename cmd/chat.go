package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
Type /load <id> to load a session.
Type /sessions to list saved sessions.`,
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
			parts := strings.SplitN(input, " ", 2)
			cmdName := parts[0]
			cmdArg := ""
			if len(parts) > 1 {
				cmdArg = strings.TrimSpace(parts[1])
			}
			switch cmdName {
			case "/exit", "/quit":
				fmt.Println("Goodbye!")
				return nil
			case "/clear":
				ag.ResetSession()
				fmt.Println("Memory cleared.")
				continue
			case "/history":
				for _, m := range ag.Memory().History() {
					content := m.Content
					if len(content) > 120 {
						content = content[:120] + "..."
					}
					fmt.Printf("  [%s] %s\n", m.Role, content)
				}
				continue
			case "/save":
				if cmdArg == "" {
					fmt.Println("Usage: /save <name>")
					continue
				}
				if err := ag.SaveSession(cmdArg); err != nil {
					fmt.Printf("  Save failed: %v\n", err)
				} else {
					fmt.Printf("  Session saved as %q.\n", cmdArg)
				}
				continue
			case "/load":
				if cmdArg == "" {
					fmt.Println("Usage: /load <id>")
					continue
				}
				if err := ag.LoadSession(cmdArg); err != nil {
					fmt.Printf("  Load failed: %v\n", err)
				} else {
					fmt.Printf("  Session %q loaded.\n", cmdArg)
				}
				continue
			case "/sessions":
				list, err := ag.ListSessions()
				if err != nil {
					fmt.Printf("  Error: %v\n", err)
					continue
				}
				if len(list) == 0 {
					fmt.Println("  No saved sessions.")
				} else {
					fmt.Println("  Saved sessions:")
					for _, s := range list {
						fmt.Printf("    %s — %s (%s)\n", s.ID, s.Task, s.UpdatedAt.Format("2006-01-02 15:04"))
					}
				}
				continue
			default:
				fmt.Printf("Unknown command: %s\n", input)
				fmt.Println("Available: /exit, /clear, /history, /save <name>, /load <id>, /sessions")
				continue
			}
		}

		// Execute task
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
		result, err := ag.Run(ctx, input)
		cancel()
		if err != nil {
			fmt.Printf("  %s: %v\n", color.New(color.FgRed).Sprint("Error"), err)
			continue
		}
		fmt.Printf("\n%s\n\n", green(result))
	}

	return nil
}
