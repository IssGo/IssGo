# Contributing to IssGo

Thanks for your interest! Here's everything you need to know.

## Quick Links

- [Bug Report](https://github.com/VyCen/issgo/issues/new?template=bug_report.yml)
- [Feature Request](https://github.com/VyCen/issgo/issues/new?template=feature_request.yml)
- [Tool Request](https://github.com/VyCen/issgo/issues/new?template=tool_request.yml)
- [Discussions](https://github.com/VyCen/issgo/discussions)

## Development Setup

```bash
git clone https://github.com/VyCen/issgo.git
cd issgo
go mod tidy
go build -o issgo .
export ISSGO_LLM_API_KEY=sk-xxx
./issgo run "say hello"
```

## Project Structure

| Directory | Purpose |
|-----------|---------|
| `cmd/` | CLI commands (cobra) |
| `agent/` | Agent loop, planning, memory, safety |
| `tools/` | Tool interface + implementations |
| `llm/` | LLM clients (OpenAI, Ollama) |
| `config/` | Configuration loading |
| `server/` | HTTP API server |
| `internal/` | Shared utilities |

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- `gofmt -s -w .` before every commit
- Package documentation starts with `// Package xxx does yyy.`
- Functions over 50 lines should be split

## Adding a New Tool

1. Create `tools/yourtool.go`
2. Implement the `Tool` interface:
   ```go
   type YourTool struct{}
   func (t *YourTool) Name() string        { return "yourtool" }
   func (t *YourTool) Description() string { return "What this tool does" }
   func (t *YourTool) Schema() any         { return map[string]any{...} }
   func (t *YourTool) Execute(ctx context.Context, args json.RawMessage) Result { ... }
   ```
3. Register in `tools/tools.go` NewRegistry
4. Add to `config/config.go` ToolsConfig
5. Add tests
6. Update README tool list

## Commit Convention

```
type: short description

- feat: new feature
- fix: bug fix
- tool: new or updated tool
- docs: documentation
- refactor: code restructuring
- test: adding or updating tests
```

## Getting Help

Open a [Discussion](https://github.com/VyCen/issgo/discussions) or comment on an existing issue.
