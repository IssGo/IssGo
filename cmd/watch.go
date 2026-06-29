package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/issgo/issgo/agent"
	"github.com/issgo/issgo/config"
	"github.com/issgo/issgo/internal/logger"
	"github.com/spf13/cobra"
)

var (
	watchDir   string
	onChange   string
	debounceMs int
)

var watchCmd = &cobra.Command{
	Use:   "watch [directory]",
	Short: "Watch a directory and trigger on changes",
	Long: `Watch a directory recursively for file changes and execute an action when changes are detected.

Examples:
  issgo watch . --on-change "run tests"
  issgo watch ./src --on-change "format all go files"
  issgo watch ./config --debounce 1000 --on-change "restart the server"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringVarP(&onChange, "on-change", "c", "", "Action to perform on file change (required)")
	watchCmd.Flags().IntVar(&debounceMs, "debounce", 500, "Debounce delay in milliseconds")
	watchCmd.MarkFlagRequired("on-change")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	watchDir = "."
	if len(args) > 0 {
		watchDir = args[0]
	}

	absDir, err := filepath.Abs(watchDir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger.Init(cfg.Agent.Verbose)
	defer logger.Sync()

	if cfg.LLM.APIKey == "" {
		return fmt.Errorf("no API key configured")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	// Add directory and subdirectories
	if err := filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walk directory: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Printf("\n%s %s\n", green("Watching:"), absDir)
	fmt.Printf("%s %s\n\n", green("On change:"), onChange)

	ag := agent.New(cfg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var debounce *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Ignore hidden files and temp files
			base := filepath.Base(event.Name)
			if len(base) > 0 && base[0] == '.' {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			fmt.Printf("  %s %s\n", bold("Changed:"), event.Name)

			if debounce != nil {
				debounce.Stop()
			}

			debounce = time.AfterFunc(time.Duration(debounceMs)*time.Millisecond, func() {
				fmt.Printf("\n%s\n\n", bold("Executing: "+onChange))
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()

				result, err := ag.Run(ctx, onChange)
				if err != nil {
					fmt.Printf("  Error: %v\n", err)
				} else {
					fmt.Printf("  Result: %s\n", result)
				}
				fmt.Println()
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logger.Log.Errorw("watch error", "error", err)

		case <-sigCh:
			fmt.Println("\nStopping watcher...")
			return nil
		}
	}
}
