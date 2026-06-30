package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/issgo/issgo/internal/utils"
)

// ─── Plugin system ─────────────────────────────────────────────

// PluginTool wraps an external script/program as a Tool.
type PluginTool struct {
	name        string
	description string
	schema      any
	execPath    string
}

func LoadPlugins(dir string) ([]*PluginTool, error) {
	dir = utils.ExpandPath(dir)
	if !utils.FileExists(dir) {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read plugin dir: %w", err)
	}

	var plugins []*PluginTool
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		info, err := e.Info()
		if err != nil || info.Mode()&0o111 == 0 {
			continue // not executable
		}

		pt := probePlugin(path)
		if pt != nil {
			plugins = append(plugins, pt)
		}
	}
	return plugins, nil
}

func probePlugin(path string) *PluginTool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--issgo-manifest")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var manifest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Schema      any    `json:"schema"`
	}
	if err := json.Unmarshal(out, &manifest); err != nil {
		// Try reading manifest from stdout line-by-line
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			if err := json.Unmarshal(scanner.Bytes(), &manifest); err == nil {
				break
			}
		}
	}

	if manifest.Name == "" {
		return nil
	}

	return &PluginTool{
		name:        manifest.Name,
		description: manifest.Description,
		schema:      manifest.Schema,
		execPath:    path,
	}
}

func (p *PluginTool) Name() string        { return p.name }
func (p *PluginTool) Description() string { return p.description }
func (p *PluginTool) Schema() any         { return p.schema }

func (p *PluginTool) Execute(ctx context.Context, args json.RawMessage) Result {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.execPath)
	cmd.Stdin = strings.NewReader(string(args))
	out, err := cmd.CombinedOutput()

	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Success: false, Output: output, Error: err.Error()}
	}

	var result Result
	if json.Unmarshal(out, &result) != nil {
		result = ResultOk(output)
	}
	return result
}
