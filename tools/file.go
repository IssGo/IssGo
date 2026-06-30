package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/issgo/issgo/internal/utils"
)

type FileTool struct{}

func (f *FileTool) Name() string        { return "file" }
func (f *FileTool) Description() string { return "Read, write, list, delete, copy, move, and inspect files/directories." }

func (f *FileTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type": "string",
				"enum": []string{"read", "write", "append", "list", "delete", "copy", "move", "exists", "stat", "mkdir"},
				"description": "File operation to perform.",
			},
			"path":     map[string]any{"type": "string", "description": "Target file or directory path."},
			"dest":     map[string]any{"type": "string", "description": "Destination path for copy/move."},
			"content":  map[string]any{"type": "string", "description": "Content for write/append."},
			"encoding": map[string]any{"type": "string", "description": "File encoding hint."},
		},
		"required": []string{"action", "path"},
	}
}

type fileArgs struct {
	Action   string `json:"action"`
	Path     string `json:"path"`
	Dest     string `json:"dest"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func (f *FileTool) Execute(_ context.Context, raw json.RawMessage) Result {
	var args fileArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return ResultErr("invalid args: " + err.Error())
	}
	args.Path = utils.ExpandPath(args.Path)

	switch args.Action {
	case "read":
		data, err := os.ReadFile(args.Path)
		if err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(string(data))

	case "write":
		dir := filepath.Dir(args.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return ResultErr(err.Error())
		}
		if err := os.WriteFile(args.Path, []byte(args.Content), 0o644); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(fmt.Sprintf("wrote %d bytes to %s", len(args.Content), args.Path))

	case "append":
		dir := filepath.Dir(args.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return ResultErr(err.Error())
		}
		fi, err := os.OpenFile(args.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return ResultErr(err.Error())
		}
		defer fi.Close()
		n, err := fi.WriteString(args.Content)
		if err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(fmt.Sprintf("appended %d bytes to %s", n, args.Path))

	case "list":
		entries, err := os.ReadDir(args.Path)
		if err != nil {
			return ResultErr(err.Error())
		}
		var lines []string
		for _, e := range entries {
			info, _ := e.Info()
			tag := ""
			if e.IsDir() {
				tag = "/"
			}
			lines = append(lines, fmt.Sprintf("%s  %s%s  %s  %s",
				info.Mode(), utils.HumanBytes(info.Size()),
				tag, info.ModTime().Format(time.RFC3339), e.Name()))
		}
		return ResultOk(strings.Join(lines, "\n"))

	case "delete":
		if err := os.RemoveAll(args.Path); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk("deleted " + args.Path)

	case "copy":
		data, err := os.ReadFile(args.Path)
		if err != nil {
			return ResultErr(err.Error())
		}
		os.MkdirAll(filepath.Dir(args.Dest), 0o755)
		if err := os.WriteFile(args.Dest, data, 0o644); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(fmt.Sprintf("copied %s -> %s", args.Path, args.Dest))

	case "move":
		os.MkdirAll(filepath.Dir(args.Dest), 0o755)
		if err := os.Rename(args.Path, args.Dest); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(fmt.Sprintf("moved %s -> %s", args.Path, args.Dest))

	case "exists":
		ok := utils.FileExists(args.Path)
		return ResultOk(fmt.Sprintf("%v", ok))

	case "stat":
		info, err := os.Stat(args.Path)
		if err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(fmt.Sprintf("Name: %s\nSize: %s\nMode: %s\nModTime: %s\nIsDir: %v",
			info.Name(), utils.HumanBytes(info.Size()),
			info.Mode(), info.ModTime().Format(time.RFC3339), info.IsDir()))

	case "mkdir":
		if err := os.MkdirAll(args.Path, 0o755); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk("created directory " + args.Path)

	default:
		return ResultErr("unknown action: " + args.Action)
	}
}
