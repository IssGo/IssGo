# IssGo

[![中文](https://img.shields.io/badge/lang-中文-red.svg)](README.md) [![English](https://img.shields.io/badge/lang-English-blue.svg)](README-en.md)

> **AI Agent + Automation CLI Tool** — Describe tasks in natural language, the Agent autonomously plans and executes them.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Build from Source](#build-from-source)
  - [Using go install](#using-go-install)
- [Quick Start](#quick-start)
- [Commands](#commands)
  - [issgo init](#issgo-init)
  - [issgo run](#issgo-run)
  - [issgo watch](#issgo-watch)
  - [issgo chat](#issgo-chat)
  - [issgo serve](#issgo-serve)
  - [issgo config](#issgo-config)
- [Configuration](#configuration)
  - [Config File Format](#config-file-format)
  - [Config Load Priority](#config-load-priority)
  - [Environment Variables](#environment-variables)
  - [Profiles](#profiles)
- [Architecture](#architecture)
  - [Project Structure](#project-structure)
  - [Core Flow](#core-flow)
  - [ReAct Loop](#react-loop)
- [Built-in Tools](#built-in-tools)
  - [file — File Operations](#file--file-operations)
  - [shell — Shell Commands](#shell--shell-commands)
  - [web — HTTP Requests](#web--http-requests)
  - [browser — Browser Automation](#browser--browser-automation)
  - [git — Git Operations](#git--git-operations)
  - [search — File Content Search](#search--file-content-search)
  - [plugins — Plugin System](#plugins--plugin-system)
- [LLM Providers](#llm-providers)
  - [DeepSeek (Default)](#deepseek-default)
  - [OpenAI](#openai)
  - [Other Compatible Services](#other-compatible-services)
- [Usage Examples](#usage-examples)
  - [Code Analysis](#code-analysis)
  - [File Batch Processing](#file-batch-processing)
  - [Automation Workflows](#automation-workflows)
  - [Web Scraping](#web-scraping)
  - [Watch Mode](#watch-mode)
- [Development](#development)
  - [Dependency Management](#dependency-management)
  - [Running Tests](#running-tests)
  - [Tech Stack](#tech-stack)
- [FAQ](#faq)
- [License](#license)

---

## Introduction

IssGo is an AI Agent CLI tool that runs on your local machine. It accepts tasks described in natural language, automatically breaks them down into steps, calls built-in tools (file operations, shell commands, HTTP requests, headless browser) to execute, and returns the summarized results.

**Core Philosophy:** You only describe "what to do", IssGo handles "how to do it".

```
$ issgo run "Find all TODO comments in the project, group by file, save to todos.md"

Task: Find all TODO comments in the project, group by file, save to todos.md

  [plan] 1. Search all Go files for TODO comments
  [step 1] grep -rn "TODO" --include="*.go" . → Found 23
  [step 2] Group by file and sort
  [step 3] Write to todos.md

Result: Found 23 TODO comments across 12 files, saved to todos.md.
```

---

## Installation

### Prerequisites

- **Go** 1.24 or higher
- **DeepSeek API Key** (or any OpenAI-compatible API key)
- (Optional) **Chrome / Chromium** — required for the `browser` tool

### Build from Source

```bash
git clone https://github.com/IssGo/IssGo.git
cd issgo
go mod tidy
go build -o issgo .
sudo mv issgo /usr/local/bin/
```

### Using go install

```bash
go install github.com/IssGo/IssGo@v2026.07.01
```

---

## Quick Start

```bash
# 1. Initialize config file
issgo init

# 2. Edit .issgo.yaml to set your API Key
vim .issgo.yaml

# 3. Or use environment variables directly
export ISSGO_LLM_API_KEY="sk-your-deepseek-key"

# 4. Run your first task
issgo run "Count the number of files of each type in the current directory"
```

---

## Commands

### issgo init

Generate a `.issgo.yaml` configuration file in the current directory.

```bash
issgo init                # Generate default config
issgo init --force        # Overwrite existing config (alias -f)
```

### issgo run

Execute an AI task. The task description can be any natural language.

```bash
issgo run "List all JSON files and verify they are well-formed"
issgo run "Move files older than 7 days from logs/ to archive/"
issgo run "Extract all dependency names and versions from package.json"
```

**Optional Arguments:**

| Flag | Description |
|------|-------------|
| `--verbose` / `-v` | Output detailed execution logs |
| `--no-spinner` | Disable progress spinner |
| `--no-cache` | Skip LLM response cache |
| `--profile` / `-p` | Use a specific profile configuration |

### issgo watch

Recursively watch a directory for changes, then trigger an AI action automatically.

```bash
issgo watch . --on-change "run go vet and fix all issues"
issgo watch ./src --on-change "format modified files" --debounce 1000
```

**Arguments:**

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--on-change` / `-c` | Yes | — | AI task to run on file change |
| `--debounce` / `-d` | No | `500` | Debounce delay (ms) |
| `--once` | No | `false` | Exit after the first change |

### issgo chat

Start an interactive chat session with persistent memory.

```bash
issgo chat

# Supported commands:
#   /exit, /quit  — Exit
#   /clear        — Clear conversation memory
#   /history      — Show conversation history
#   /save <name>  — Save session
#   /load <id>    — Load a session
#   /sessions     — List saved sessions
```

### issgo serve

Start an HTTP API server to expose Agent capabilities via REST.

```bash
issgo serve                  # Default 127.0.0.1:8420
issgo serve --port 8080      # Custom port
issgo serve --host 0.0.0.0   # Allow external access

# API Endpoints:
#   GET  /api/v1/health       — Health check
#   POST /api/v1/run          — Execute a task
#   POST /api/v1/stream       — SSE streaming execution
#   GET  /api/v1/tools        — List available tools
```

### issgo config

View current configuration and available profiles.

```bash
issgo config                  # Summary view
issgo config --all            # Full config (JSON)
```

---

## Configuration

### Config File Format

```yaml
# .issgo.yaml

llm:
  provider: deepseek                  # deepseek | openai | custom
  model: deepseek-chat                # Model name
  api_key: ""                         # API key (or use ISSGO_LLM_API_KEY env var)
  base_url: https://api.deepseek.com  # API base URL

tools:
  shell: true                         # Allow shell command execution
  file: true                          # Allow file operations
  web: true                           # Allow HTTP requests
  browser: false                      # Allow headless browser (requires Chrome)
  git: true                           # Allow git operations
  search: true                        # Allow file content search
  plugins: false                      # Allow third-party plugins

agent:
  max_steps: 30                       # Max tool calls per task
  allow_approve: true                 # Ask user before dangerous actions
  verbose: false                      # Detailed debug output
  streaming: true                     # Stream LLM responses
  reflector: true                     # Self-evaluation after task
  max_retries: 3                      # LLM call retry count
```

### Config Load Priority

```
Env Vars (ISSGO_*)  >  ./.issgo.yaml  >  ~/.issgo.yaml  >  Built-in defaults
```

### Environment Variables

| Variable | Config Key | Example |
|----------|-----------|---------|
| `ISSGO_LLM_API_KEY` | `llm.api_key` | `sk-xxxxxxxx` |
| `ISSGO_LLM_MODEL` | `llm.model` | `deepseek-chat` |
| `ISSGO_LLM_PROVIDER` | `llm.provider` | `deepseek` |

> It is recommended to set your API Key in environment variables rather than plaintext in config file:
> ```bash
> export ISSGO_LLM_API_KEY="sk-xxxxxxxx"
> ```

### Profiles

Profiles allow you to preset multiple LLM configurations and switch quickly via `--profile` / `-p` or the `active` field:

```yaml
profiles:
  - name: deepseek
    provider: deepseek
    model: deepseek-chat
    base_url: https://api.deepseek.com

  - name: openai
    provider: openai
    model: gpt-4o
    base_url: https://api.openai.com/v1

  - name: ollama
    provider: ollama
    model: qwen2.5:14b
    base_url: http://localhost:11434

active: deepseek    # Set to empty string to use default config
```

```bash
issgo run "your task" --profile ollama
```

---

## Architecture

### Project Structure

```
issgo/
├── main.go                    # Entry point
├── go.mod
├── cmd/
│   ├── root.go                # CLI root command (Cobra)
│   ├── init.go                # issgo init
│   ├── run.go                 # issgo run
│   ├── chat.go                # issgo chat (interactive mode)
│   ├── serve.go               # issgo serve (HTTP API)
│   ├── config.go              # issgo config
│   ├── watch.go               # issgo watch (fsnotify)
│   └── version.go             # issgo version
├── agent/
│   ├── agent.go               # Agent facade
│   ├── planner.go             # Task planner
│   ├── executor.go            # Executor (ReAct loop)
│   ├── reflector.go           # Self-reflection module
│   ├── safety.go              # Safety review
│   ├── memory.go              # Conversation history management
│   ├── session.go             # Session persistence
│   └── stream.go              # Streaming executor
├── tools/
│   ├── tools.go               # Tool interface + Registry
│   ├── file.go                # File tool (10 operations)
│   ├── shell.go               # Shell tool (security + output cleanup)
│   ├── web.go                 # HTTP tool (resty)
│   ├── browser.go             # Browser tool (chromedp)
│   ├── git.go                 # Git tool
│   ├── search.go              # Search tool (regex + glob)
│   └── plugin.go              # Plugin system
├── llm/
│   ├── provider.go            # Provider interface + types
│   ├── client.go              # LLM client (retry + cache + multi-provider)
│   ├── openai.go              # OpenAI-compatible provider
│   ├── ollama.go              # Ollama provider (local models)
│   └── cache.go               # LRU response cache
├── server/
│   ├── server.go              # HTTP server
│   ├── handler.go             # API route handlers
│   └── middleware.go           # Logging / CORS / Auth middleware
├── config/
│   ├── config.go              # Config loader (viper)
│   ├── profile.go             # Profile management
│   └── validate.go            # Config validation
├── prompts/
│   └── templates.go           # System / Planner / Reflector / Safety / Memory templates
└── internal/
    ├── safe/
    │   └── safe.go            # Static command safety audit
    ├── diff/
    │   └── diff.go            # Text diff computation
    ├── logger/
    │   └── logger.go          # Structured logging (zap)
    ├── spinner/
    │   └── spinner.go         # Terminal spinner
    ├── progress/
    │   └── progress.go        # Progress bar
    └── utils/
        └── utils.go           # Utility functions
```

### Core Flow

```
User Input: "List all Go files and count lines"
    │
    ▼
┌──────────┐
│   CLI    │  cobra parses command → loads config → initializes logger
└────┬─────┘
     │
     ▼
┌──────────┐
│  Agent   │  Creates LLM Client + Tool Registry + Memory + Executor
└────┬─────┘
     │
     ▼
┌──────────┐
│ Executor │  Enters ReAct loop
└────┬─────┘
     │
     ├─→ LLM: "Need to list files first" → tool_call: shell(find . -name "*.go")
     │       ← Tool returns file list
     │
     ├─→ LLM: "Count lines" → tool_call: shell(wc -l *.go)
     │       ← Tool returns count
     │
     └─→ LLM: Generates final response → returns to user
```

### ReAct Loop

IssGo uses the classic **ReAct (Reasoning + Acting)** pattern:

1. **Reason**: LLM analyzes the current state and decides next action
2. **Act**: Calls a Tool to perform an operation
3. **Observe**: Collects results from the Tool
4. **Repeat**: Until task completion or `max_steps` limit

---

## Built-in Tools

IssGo has 7 built-in tools, each can be toggled independently in `.issgo.yaml`.

### file — File Operations

| Operation | Description |
|-----------|-------------|
| `read` | Read file content |
| `write` | Write to file (auto-create parent dir) |
| `append` | Append content to file |
| `list` | List directory contents |
| `delete` | Delete file or directory (recursive) |
| `copy` | Copy file |
| `move` | Move file |
| `exists` | Check if file/dir exists |
| `stat` | Get detailed file info |
| `mkdir` | Create directory |

### shell — Shell Commands

Execute shell commands via `bash -c` with a 120-second timeout.

```bash
find . -name "*.md" | head -20
grep -rn "TODO" --include="*.go" .
cat /etc/os-release
```

### web — HTTP Requests

Full HTTP client based on [resty](https://github.com/go-resty/resty).

### browser — Browser Automation

Headless Chrome control via [chromedp](https://github.com/chromedp/chromedp).

| Operation | Description |
|-----------|-------------|
| `navigate` | Navigate to URL, return page title |
| `screenshot` | Take screenshot (returns byte count) |
| `content` | Extract HTML from page or CSS selector |

### git — Git Operations

Supports common git subcommands: `status`, `diff`, `log`, `branch`, `add`, `commit`, `pull`, `push`, `checkout`, `stash`, `tag`, `remote`, `show`, `blame`, `describe`, `rev-parse`.

### search — File Content Search

Recursive file content search supporting regex, literal matching, and glob file filtering.

### plugins — Plugin System

Place executable scripts in `~/.issgo/plugins/`. Scripts supporting `--issgo-manifest` to output a JSON manifest will be registered as tools.

---

## LLM Providers

### DeepSeek (Default)

```yaml
llm:
  provider: deepseek
  model: deepseek-chat
  base_url: https://api.deepseek.com
  api_key: "sk-xxxxxxxx"
```

### OpenAI

```yaml
llm:
  provider: openai
  model: gpt-4o
  base_url: https://api.openai.com/v1
  api_key: "sk-xxxxxxxx"
```

### Other Compatible Services

Any service compatible with the OpenAI Chat Completions API can be used.

---

## Usage Examples

### Code Analysis

```bash
# Find unused imports
issgo run "Check all Go files for unused imports"

# Count lines of code
issgo run "Count lines of code grouped by language, output as Markdown table"

# Code review
issgo run "Review code in cmd/ for complete error handling"
```

### File Batch Processing

```bash
# Format conversion
issgo run "Convert all CSV files in data/ to JSON, preserving filenames"

# Batch rename
issgo run "Rename all .jpg files to lowercase with underscores instead of spaces"

# Find and replace
issgo run "In all .ts files under src/, replace require() with import"
```

### Automation Workflows

```bash
# System health report
issgo run "Check disk usage, memory, and CPU load, generate a health report"

# Git branch operations
issgo run "Create new branch feature/auth from main and push to origin"

# Log analysis
issgo run "Analyze nginx access.log, report Top 10 IPs and status code distribution"
```

### Watch Mode

```bash
# Auto-test
issgo watch . --on-change "Run go test ./... and fix failing tests"

# Auto-format
issgo watch ./src --on-change "Format all modified Go files with gofmt" --debounce 2000

# Auto-docs
issgo watch ./api --on-change "Update API.md based on code changes"
```

---

## Development

### Dependency Management

```bash
# Download dependencies
go mod tidy

# View dependency tree
go mod graph

# Update dependencies
go get -u ./...
```

### Running Tests

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| CLI Framework | [cobra](https://github.com/spf13/cobra) | Command parsing & routing |
| Config Management | [viper](https://github.com/spf13/viper) | YAML + environment variables |
| Logging | [zap](https://github.com/uber-go/zap) | Structured logging |
| LLM Client | [go-openai](https://github.com/sashabaranov/go-openai) | OpenAI API calls |
| HTTP Client | [resty](https://github.com/go-resty/resty) | HTTP requests |
| Browser Automation | [chromedp](https://github.com/chromedp/chromedp) | Headless Chrome |
| File Watching | [fsnotify](https://github.com/fsnotify/fsnotify) | Directory change monitoring |
| Terminal Colors | [color](https://github.com/fatih/color) | CLI output styling |

---

## FAQ

<details>
<summary><b>Q: How to securely store API Keys?</b></summary>

Use environment variables:

```bash
# Add to ~/.bashrc or ~/.zshrc
export ISSGO_LLM_API_KEY="sk-xxxxxxxx"
```

You can also put it in `.issgo.yaml`, but avoid committing it to Git.
</details>

<details>
<summary><b>Q: What should max_steps be set to?</b></summary>

The default of 30 is enough for most tasks. Simple tasks usually take 5-10 steps. Complex multi-stage tasks can be increased to 50.
</details>

<details>
<summary><b>Q: What safety restrictions does the shell tool have?</b></summary>

Two-layer safety: ① Static analysis automatically blocks dangerous patterns (rm -rf /, fork bomb, etc.); ② LLM dynamic review for warning-level commands. Also has a 120s timeout. Enable `allow_approve: true` for confirmation before dangerous commands.
</details>

<details>
<summary><b>Q: Which operating systems are supported?</b></summary>

Linux, macOS, Windows (via Git Bash or WSL). `file` and `shell` tools work on all platforms; `browser` requires additional Chrome installation.
</details>

<details>
<summary><b>Q: Can I use local models?</b></summary>

Yes. Run Ollama or another compatible service, then point the config to your local address. Note: local models need to support Function Calling.
</details>

<details>
<summary><b>Q: How do I debug LLM calls?</b></summary>

Set `agent.verbose: true` in config or add `-v` to the command line:

```bash
issgo run "your task" -v
```

This prints the full prompt, LLM response, and tool call details.
</details>

---

## License

[MIT](LICENSE)

---

<p align="center">
  <sub>Made with ❤️ by the IssGo team</sub>
</p>
