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
	watchOnce  bool
)

var watchCmd = &cobra.Command{
	Use:   "watch [directory]",
	Short: "Watch a directory and trigger on changes",
	Long: `Watch a directory recursively for file changes, then execute an AI action.

Examples:
  issgo watch . --on-change "run tests"
  issgo watch ./src --on-change "format all Go files"
  issgo watch ./config --debounce 2000 --on-change "restart the server"
  issgo watch . --on-change "lint changed files" --once`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringVarP(&onChange, "on-change", "c", "", "AI action to perform on file change (required)")
	watchCmd.Flags().IntVarP(&debounceMs, "debounce", "d", 500, "Debounce delay in milliseconds")
	watchCmd.Flags().BoolVar(&watchOnce, "once", false, "Run once on first change and exit")
	watchCmd.MarkFlagRequired("on-change")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	watchDir = "."
	if len(args) > 0 {
		watchDir = args[0]
	}
	absDir, _ := filepath.Abs(watchDir)

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	logger.Init(logger.Config{Verbose: cfg.Agent.Verbose})
	defer logger.Sync()

	if errs := config.Validate(cfg); len(errs) > 0 {
		return fmt.Errorf(config.FormatErrors(errs))
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == ".git" || base == "node_modules" || base == "vendor" {
			return filepath.SkipDir
		}
		watcher.Add(path)
		return nil
	})

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("\n%s %s\n%s %s\n\n", green("Watching:"), absDir, green("On change:"), onChange)

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
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			base := filepath.Base(event.Name)
			if len(base) > 0 && base[0] == '.' {
				continue
			}

			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(time.Duration(debounceMs)*time.Millisecond, func() {
				fmt.Printf("\n%s %s\n\n", color.New(color.Bold).Sprint("Executing:"), onChange)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
				defer cancel()
				result, err := ag.Run(ctx, onChange)
				if err != nil {
					fmt.Printf("  Error: %v\n", err)
				} else {
					fmt.Printf("  %s\n\n", result)
				}
				if watchOnce {
					fmt.Println("Exiting (--once).")
					os.Exit(0)
				}
			})

		case <-sigCh:
			fmt.Println("\nStopping watcher...")
			return nil
		}
	}
}
