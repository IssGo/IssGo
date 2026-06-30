package tools

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

type GitTool struct{}

func (g *GitTool) Name() string { return "git" }
func (g *GitTool) Description() string {
	return "Execute git operations: status, diff, log, branch, commit, etc."
}

func (g *GitTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"enum":        []string{"status", "diff", "log", "branch", "add", "commit", "pull", "push", "checkout", "stash", "tag", "remote", "show", "blame", "describe", "rev-parse"},
				"description": "Git subcommand.",
			},
			"repo_path": map[string]any{"type": "string", "description": "Path to git repo (default: cwd)."},
			"args":      map[string]any{"type": "string", "description": "Additional arguments."},
		},
		"required": []string{"command"},
	}
}

type gitArgs struct {
	Command  string `json:"command"`
	RepoPath string `json:"repo_path"`
	Args     string `json:"args"`
}

var allowedGitCommands = map[string]bool{
	"status": true, "diff": true, "log": true, "branch": true,
	"add": true, "commit": true, "pull": true, "push": true,
	"checkout": true, "stash": true, "tag": true, "remote": true,
	"show": true, "blame": true, "describe": true, "rev-parse": true,
}

func (g *GitTool) Execute(ctx context.Context, raw json.RawMessage) Result {
	var args gitArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return ResultErr("invalid args: " + err.Error())
	}

	if !allowedGitCommands[args.Command] {
		return ResultErr("git command not allowed: " + args.Command)
	}

	dir := args.RepoPath
	if dir == "" {
		dir = "."
	}

	cmdArgs := []string{args.Command}
	if args.Args != "" {
		cmdArgs = append(cmdArgs, strings.Fields(args.Args)...)
	}

	// Safety: disallow force push to main/master
	if args.Command == "push" && strings.Contains(args.Args, "--force") {
		if strings.Contains(args.Args, "main") || strings.Contains(args.Args, "master") {
			return ResultErr("force push to main/master is blocked for safety")
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	output := string(out)
	if len(output) > 20000 {
		output = output[:20000] + "\n... (truncated)"
	}
	if output == "" {
		output = "(no output)"
	}

	if err != nil {
		return Result{Success: false, Output: output, Error: err.Error()}
	}

	return ResultOk(output)
}
