package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileTool struct{}

func (f *FileTool) Name() string        { return "file" }
func (f *FileTool) Description() string { return "Read, write, list, and delete files on the local filesystem." }

func (f *FileTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"enum":        []string{"read", "write", "list", "delete", "exists"},
				"description": "The file operation to perform.",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "The file or directory path.",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Content to write (required for write action).",
			},
		},
		"required": []string{"action", "path"},
	}
}

type fileArgs struct {
	Action  string `json:"action"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (f *FileTool) Execute(ctx context.Context, raw json.RawMessage) Result {
	var args fileArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("invalid args: %v", err)}
	}

	switch args.Action {
	case "read":
		data, err := os.ReadFile(args.Path)
		if err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		return Result{Success: true, Output: string(data)}

	case "write":
		dir := filepath.Dir(args.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		if err := os.WriteFile(args.Path, []byte(args.Content), 0o644); err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		return Result{Success: true, Output: fmt.Sprintf("wrote %d bytes to %s", len(args.Content), args.Path)}

	case "list":
		entries, err := os.ReadDir(args.Path)
		if err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		var names []string
		for _, e := range entries {
			suffix := ""
			if e.IsDir() {
				suffix = "/"
			}
			names = append(names, e.Name()+suffix)
		}
		return Result{Success: true, Output: strings.Join(names, "\n")}

	case "delete":
		if err := os.RemoveAll(args.Path); err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		return Result{Success: true, Output: "deleted " + args.Path}

	case "exists":
		_, err := os.Stat(args.Path)
		exists := err == nil
		return Result{Success: true, Output: fmt.Sprintf("%v", exists)}

	default:
		return Result{Success: false, Error: "unknown action: " + args.Action}
	}
}
