# IssGo

> **AI Agent + 自动化 CLI 工具** — 用自然语言下达任务，Agent 自动规划并执行。

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 目录

- [介绍](#介绍)
- [安装](#安装)
  - [前置条件](#前置条件)
  - [从源码编译](#从源码编译)
  - [使用 go install](#使用-go-install)
- [快速开始](#快速开始)
- [命令参考](#命令参考)
  - [issgo init](#issgo-init)
  - [issgo run](#issgo-run)
  - [issgo watch](#issgo-watch)
  - [issgo chat](#issgo-chat)
  - [issgo serve](#issgo-serve)
  - [issgo config](#issgo-config)
- [配置](#配置)
  - [配置文件格式](#配置文件格式)
  - [配置加载优先级](#配置加载优先级)
  - [环境变量](#环境变量)
  - [Profiles](#profiles多配置切换)
- [架构](#架构)
  - [项目结构](#项目结构)
  - [核心流程](#核心流程)
  - [ReAct 循环](#react-循环)
- [内置工具](#内置工具)
  - [file — 文件操作](#file--文件操作)
  - [shell — Shell 命令](#shell--shell-命令)
  - [web — HTTP 请求](#web--http-请求)
  - [browser — 浏览器自动化](#browser--浏览器自动化)
  - [git — Git 操作](#git--git-操作)
  - [search — 文件内容搜索](#search--文件内容搜索)
  - [plugins — 插件系统](#plugins--插件系统)
- [LLM 提供商](#llm-提供商)
  - [DeepSeek（默认）](#deepseek默认)
  - [OpenAI](#openai)
  - [其他兼容服务](#其他兼容服务)
- [使用示例](#使用示例)
  - [代码分析](#代码分析)
  - [文件批处理](#文件批处理)
  - [自动化工作流](#自动化工作流)
  - [Web 抓取](#web-抓取)
  - [watch 模式](#watch-模式)
- [开发](#开发)
  - [依赖管理](#依赖管理)
  - [运行测试](#运行测试)
  - [技术栈](#技术栈)
- [FAQ](#faq)
- [License](#license)

---

## 介绍

IssGo 是一个运行在本地机器上的 AI Agent 命令行工具。它接收自然语言描述的任务，自动分解为步骤，调用内置工具（文件操作、Shell 命令、HTTP 请求、无头浏览器）执行，并将结果汇总返回。

**核心理念：** 你只需要描述"做什么"，IssGo 负责"怎么做"。

```
$ issgo run "找出项目中所有 TODO 注释，按文件分组，保存到 todos.md"

Task: 找出项目中所有 TODO 注释，按文件分组，保存到 todos.md

  [plan] 1. 搜索所有 Go 文件中的 TODO 注释
  [step 1] grep -rn "TODO" --include="*.go" . → 发现 23 条
  [step 2] 按文件分组排序
  [step 3] 写入 todos.md

Result: 已在 12 个文件中找到 23 条 TODO 注释，结果保存到 todos.md。
```

---

## 安装

### 前置条件

- **Go** 1.24 或更高版本
- **DeepSeek API Key**（或其他 OpenAI 兼容 API 密钥）
- （可选）**Chrome / Chromium** — 使用 `browser` 工具时需要

### 从源码编译

```bash
git clone https://github.com/IssGo/IssGo.git
cd issgo
go mod tidy
go build -o issgo .
sudo mv issgo /usr/local/bin/
```

### 使用 go install

```bash
go install github.com/IssGo/IssGo@v2026.06.30_2
```

---

## 快速开始

```bash
# 1. 初始化配置文件
issgo init

# 2. 编辑 .issgo.yaml，填入 API Key
vim .issgo.yaml

# 3. 或者直接使用环境变量
export ISSGO_LLM_API_KEY="sk-your-deepseek-key"

# 4. 运行你的第一个任务
issgo run "统计当前目录下每种文件类型的数量"
```

---

## 命令参考

### issgo init

在当前目录生成 `.issgo.yaml` 配置文件。

```bash
issgo init                # 生成默认配置
issgo init --force        # 覆盖已有配置（别名 -f）
```

### issgo run

执行 AI 任务。任务描述可以是任意自然语言。

```bash
issgo run "列出所有 JSON 文件并验证格式是否正确"
issgo run "将 logs/ 下超过 7 天的文件移动到 archive/"
issgo run "从 package.json 中提取所有依赖名称和版本"
```

**可选参数：**

| 参数 | 描述 |
|------|------|
| `--verbose` / `-v` | 输出详细执行日志（包含每一步的完整 LLM 响应） |
| `--no-spinner` | 禁用进度动画 |
| `--profile` / `-p` | 使用指定的 profile 配置 |

### issgo watch

递归监听目录变化，变化时自动触发 AI 任务。

```bash
issgo watch . --on-change "运行 go vet 并修复所有问题"
issgo watch ./src --on-change "格式化修改过的文件" --debounce 1000
```

**参数：**

| 参数 | 必需 | 默认值 | 描述 |
|------|------|--------|------|
| `--on-change` / `-c` | 是 | — | 文件变化时执行的 AI 任务 |
| `--debounce` / `-d` | 否 | `500` | 防抖延迟（毫秒），合并短时间内的多次变化 |
| `--once` | 否 | `false` | 首次变化后执行一次即退出 |

> 监听范围：默认 `.`，可指定任意目录。自动递归监听所有子目录。

### issgo chat

启动交互式聊天模式，Agent 记忆跨消息持久化。

```bash
issgo chat

# 支持的命令：
#   /exit, /quit  — 退出
#   /clear        — 清空对话记忆
#   /history      — 查看对话历史
#   /save <name>  — 保存会话
#   /load <name>  — 加载会话
```

### issgo serve

启动 HTTP API 服务器，通过 REST 接口暴露 Agent 能力。

```bash
issgo serve                  # 默认 127.0.0.1:8420
issgo serve --port 8080      # 指定端口
issgo serve --host 0.0.0.0   # 允许外部访问

# API 端点：
#   GET  /api/v1/health       — 健康检查
#   POST /api/v1/run          — 执行任务
#   POST /api/v1/stream       — SSE 流式执行
#   GET  /api/v1/tools        — 列出可用工具
```

### issgo config

查看当前配置和可用 profiles。

```bash
issgo config                  # 摘要视图
issgo config --all            # 完整配置（JSON）
```

---

## 配置

### 配置文件格式

```yaml
# .issgo.yaml

llm:
  provider: deepseek                  # deepseek | openai | custom
  model: deepseek-chat                # 模型名称
  api_key: ""                         # API 密钥（留空则从环境变量读取）
  base_url: https://api.deepseek.com  # API 地址

tools:
  shell: true                         # 允许 Shell 命令执行
  file: true                          # 允许文件操作
  web: true                           # 允许 HTTP 请求
  browser: false                      # 允许无头浏览器（需 Chrome）
  git: true                           # 允许 Git 操作
  search: true                        # 允许文件内容搜索
  plugins: false                      # 允许第三方插件

agent:
  max_steps: 30                       # 单次任务最大工具调用次数
  allow_approve: true                 # 危险操作前询问用户
  verbose: false                      # 详细调试输出
  streaming: true                     # 流式输出 LLM 响应
  reflector: true                     # 任务完成后自我评估
  max_retries: 3                      # LLM 调用失败重试次数
```

### 配置加载优先级

```
环境变量 (ISSGO_*)  >  ./.issgo.yaml  >  ~/.issgo.yaml  >  内置默认值
```

### 环境变量

| 环境变量 | 对应配置项 | 示例 |
|----------|-----------|------|
| `ISSGO_LLM_API_KEY` | `llm.api_key` | `sk-xxxxxxxx` |
| `ISSGO_LLM_MODEL` | `llm.model` | `deepseek-chat` |
| `ISSGO_LLM_PROVIDER` | `llm.provider` | `deepseek` |

> 推荐将 API Key 放在环境变量中，而非明文写入配置文件：
> ```bash
> export ISSGO_LLM_API_KEY="sk-xxxxxxxx"
> ```

### Profiles（多配置切换）

Profiles 允许预置多套 LLM 配置，通过 `--profile` / `-p` 参数或 `active` 字段快速切换：

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

active: deepseek    # 设为空字符串使用默认配置
```

```bash
issgo run "你的任务" --profile ollama
```

---

## 架构

### 项目结构

```
issgo/
├── main.go                    # 程序入口
├── go.mod
├── cmd/
│   ├── root.go                # CLI 根命令（Cobra）
│   ├── init.go                # issgo init
│   ├── run.go                 # issgo run
│   ├── chat.go                # issgo chat（交互模式）
│   ├── serve.go               # issgo serve（HTTP API）
│   ├── config.go              # issgo config
│   ├── watch.go               # issgo watch（fsnotify）
│   └── version.go             # issgo version
├── agent/
│   ├── agent.go               # Agent 门面：组装各组件
│   ├── planner.go             # 任务规划器（自然语言 → 步骤列表）
│   ├── executor.go            # 执行器（ReAct 循环）
│   ├── reflector.go           # 自省器（评估进展，决定是否重规划）
│   ├── safety.go              # 安全审查（危险命令检测 + LLM 确认）
│   ├── memory.go              # 对话历史管理（自动归约 + 快照）
│   ├── session.go             # 会话持久化（保存/加载/列表）
│   └── stream.go              # 流式执行器
├── tools/
│   ├── tools.go               # Tool 接口 + Registry
│   ├── file.go                # 文件工具（10 种操作）
│   ├── shell.go               # Shell 工具（安全审查 + 输出清理）
│   ├── web.go                 # HTTP 工具（resty）
│   ├── browser.go             # 浏览器工具（chromedp）
│   ├── git.go                 # Git 工具
│   ├── search.go              # 搜索工具（正则 + glob）
│   └── plugin.go              # 插件系统
├── llm/
│   ├── provider.go            # Provider 接口 + 数据类型
│   ├── client.go              # LLM 客户端（重试 + 缓存 + 多 provider）
│   ├── openai.go              # OpenAI 兼容 provider
│   ├── ollama.go              # Ollama provider（本地模型）
│   └── cache.go               # LRU 响应缓存
├── server/
│   ├── server.go              # HTTP 服务器
│   ├── handler.go             # API 路由处理
│   └── middleware.go           # 日志 / CORS / 认证中间件
├── config/
│   ├── config.go              # 配置加载（viper）
│   ├── profile.go             # Profile 管理
│   └── validate.go            # 配置校验
├── prompts/
│   └── templates.go           # System / Planner / Reflector / Safety / Memory 模板
└── internal/
    ├── safe/
    │   └── safe.go            # 静态命令安全审计
    ├── diff/
    │   └── diff.go            # 文本差异计算
    ├── logger/
    │   └── logger.go          # 结构化日志（zap）
    ├── spinner/
    │   └── spinner.go         # 终端旋转动画
    ├── progress/
    │   └── progress.go        # 进度条
    └── utils/
        └── utils.go           # 通用工具函数
```

### 核心流程

```
用户输入: "列出所有 Go 文件并统计行数"
    │
    ▼
┌──────────┐
│   CLI    │  cobra 解析命令 → 加载配置 → 初始化日志
└────┬─────┘
     │
     ▼
┌──────────┐
│  Agent   │  创建 LLM Client + Tool Registry + Memory + Executor
└────┬─────┘
     │
     ▼
┌──────────┐
│ Executor │  进入 ReAct 循环
└────┬─────┘
     │
     ├─→ LLM: "需要先列出文件" → tool_call: shell(find . -name "*.go")
     │       ← Tool 返回文件列表
     │
     ├─→ LLM: "统计行数" → tool_call: shell(wc -l *.go)
     │       ← Tool 返回统计结果
     │
     └─→ LLM: 生成最终回答 → 返回用户
```

### ReAct 循环

IssGo 使用经典的 **ReAct（Reasoning + Acting）** 模式：

1. **Reason**：LLM 分析当前状态，决定下一步
2. **Act**：调用 Tool 执行具体操作
3. **Observe**：收集 Tool 返回结果
4. **Repeat**：直到任务完成或达到 `max_steps` 上限

```
Step 0: LLM 分析任务 → 返回 tool_call（或最终回复）
Step 1: 执行 tool_call → 结果注入对话历史
Step 2: LLM 再次分析 → 返回 tool_call（或最终回复）
...
Step N: LLM 返回最终回复 → 任务完成
```

---

## 内置工具

IssGo 内置 7 个工具，可在 `.issgo.yaml` 中独立开关。

### file — 文件操作

| 操作 | 描述 |
|------|------|
| `read` | 读取文件内容 |
| `write` | 写入文件（自动创建父目录） |
| `append` | 追加内容到文件 |
| `list` | 列出目录内容 |
| `delete` | 删除文件或目录（递归） |
| `copy` | 复制文件 |
| `move` | 移动文件 |
| `exists` | 检查文件/目录是否存在 |
| `stat` | 获取文件详细信息 |
| `mkdir` | 创建目录 |

```json
// LLM 会生成类似这样的 tool_call：
{
  "name": "file",
  "arguments": {
    "action": "read",
    "path": "/home/user/project/go.mod"
  }
}
```

### shell — Shell 命令

通过 `bash -c` 执行任意 Shell 命令，120 秒超时自动终止。

```bash
# 示例：LLM 会调用 shell 工具执行
find . -name "*.md" | head -20
grep -rn "TODO" --include="*.go" .
cat /etc/os-release
```

**安全提示：** 在生产环境中使用前，请仔细审查 LLM 生成的命令。配置中设置 `allow_approve: true` 可在执行危险操作前请求用户确认。

### web — HTTP 请求

支持标准 HTTP 方法，基于 [resty](https://github.com/go-resty/resty) 实现。

```json
{
  "name": "web",
  "arguments": {
    "url": "https://api.github.com/repos/IssGo/IssGo",
    "method": "GET",
    "headers": { "Accept": "application/json" }
  }
}
```

### browser — 浏览器自动化

基于 [chromedp](https://github.com/chromedp/chromedp) 实现的无头 Chrome 控制。

| 操作 | 描述 |
|------|------|
| `navigate` | 导航到 URL，返回页面标题 |
| `screenshot` | 截图（返回字节数） |
| `content` | 提取页面或指定 CSS 选择器的 HTML |

> 需要安装 Chrome/Chromium。默认禁用，在配置中设置 `tools.browser: true` 开启。

### git — Git 操作

支持常用 Git 子命令：`status`、`diff`、`log`、`branch`、`add`、`commit`、`pull`、`push`、`checkout`、`stash`、`tag` 等。

```json
{
  "name": "git",
  "arguments": {
    "command": "status",
    "repo_path": "/path/to/repo"
  }
}
```

> 默认禁用 force push 到 main/master 分支，确保安全。

### search — 文件内容搜索

支持正则表达式、字面量匹配和 glob 文件过滤的递归搜索。

```json
{
  "name": "search",
  "arguments": {
    "pattern": "TODO",
    "path": "./src",
    "glob": "*.go",
    "literal": true
  }
}
```

### plugins — 插件系统

将可执行脚本放入 `~/.issgo/plugins/` 目录，脚本支持 `--issgo-manifest` 参数输出 JSON 清单即可被识别为插件工具。

> 默认禁用，在配置中设置 `tools.plugins: true` 并指定 `plugins_dir` 开启。

---

## LLM 提供商

### DeepSeek（默认）

```yaml
llm:
  provider: deepseek
  model: deepseek-chat
  base_url: https://api.deepseek.com
  api_key: "sk-xxxxxxxx"
```

DeepSeek 兼容 OpenAI API 格式，推荐使用 `deepseek-chat`（v3）或 `deepseek-reasoner`（r1）。

### OpenAI

```yaml
llm:
  provider: openai
  model: gpt-4o
  base_url: https://api.openai.com/v1
  api_key: "sk-xxxxxxxx"
```

### 其他兼容服务

任何兼容 OpenAI Chat Completions API 的服务都可以使用。常见兼容服务：

| 服务 | base_url | 推荐模型 |
|------|----------|---------|
| DeepSeek | `https://api.deepseek.com` | `deepseek-chat` |
| OpenAI | `https://api.openai.com/v1` | `gpt-4o`, `gpt-4o-mini` |
| Ollama（本地） | `http://localhost:11434/v1` | `llama3`, `qwen2.5` |
| Groq | `https://api.groq.com/openai/v1` | `llama-3.1-70b` |
| vLLM（自部署） | `http://your-host:8000/v1` | 任意 |

> 如果使用本地模型（Ollama / vLLM），确保模型支持 Function Calling / Tool Use。

---

## 使用示例

### 代码分析

```bash
# 查找未使用的导入
issgo run "检查项目中所有 Go 文件的 import 是否有未使用的"

# 统计代码量
issgo run "统计项目代码行数，按语言分类，输出 Markdown 表格"

# 代码审查
issgo run "审查 cmd/ 目录下的代码，检查错误处理是否完整"
```

### 文件批处理

```bash
# 格式转换
issgo run "将 data/ 下所有 CSV 转换为 JSON，保持原文件名"

# 重命名
issgo run "将所有 .jpg 文件重命名为小写并用下划线替换空格"

# 查找替换
issgo run "在 src/ 下所有 .ts 文件中，将 require() 改为 import"
```

### 自动化工作流

```bash
# 系统信息
issgo run "检查磁盘使用率、内存和 CPU 负载，生成健康报告"

# Git 操作
issgo run "创建新分支 feature/auth，从 main 创建，推送到 origin"

# 日志分析
issgo run "分析 nginx access.log，统计 Top 10 访问 IP 和状态码分布"
```

### Web 抓取

```bash
# 需要先开启 browser 工具：
issgo run "打开 https://news.ycombinator.com，提取首页所有文章标题和链接，保存为 news.md"
```

### watch 模式

```bash
# 自动测试
issgo watch . --on-change "运行 go test ./... 并修复失败的测试"

# 自动格式化
issgo watch ./src --on-change "用 gofmt 格式化所有修改过的 Go 文件" --debounce 2000

# 自动文档
issgo watch ./api --on-change "根据代码变化更新 API.md 文档"
```

---

## 开发

### 依赖管理

```bash
# 下载依赖
go mod tidy

# 查看依赖树
go mod graph

# 更新依赖
go get -u ./...
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 带覆盖率
go test -cover ./...

# 详细输出
go test -v ./...
```

### 技术栈

| 组件 | 技术 | 用途 |
|------|------|------|
| CLI 框架 | [cobra](https://github.com/spf13/cobra) | 命令解析与路由 |
| 配置管理 | [viper](https://github.com/spf13/viper) | YAML + 环境变量 |
| 日志 | [zap](https://github.com/uber-go/zap) | 结构化日志 |
| LLM 客户端 | [go-openai](https://github.com/sashabaranov/go-openai) | OpenAI API 调用 |
| HTTP 客户端 | [resty](https://github.com/go-resty/resty) | HTTP 请求工具 |
| 浏览器自动化 | [chromedp](https://github.com/chromedp/chromedp) | 无头 Chrome |
| 文件监听 | [fsnotify](https://github.com/fsnotify/fsnotify) | 目录变化监听 |
| 终端着色 | [color](https://github.com/fatih/color) | CLI 输出美化 |

---

## FAQ

<details>
<summary><b>Q: 如何安全地存储 API Key？</b></summary>

推荐使用环境变量：

```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
export ISSGO_LLM_API_KEY="sk-xxxxxxxx"
```

也可以放在 `.issgo.yaml` 中，但注意不要将文件提交到 Git 仓库。
</details>

<details>
<summary><b>Q: max_steps 应该设置多少？</b></summary>

默认 30 步对大多数任务足够。复杂任务可以调大到 50。设置太小可能导致任务未完成就被截断。
</details>

<details>
<summary><b>Q: shell 工具有哪些安全限制？</b></summary>

默认启用了 120 秒超时。所有命令在当前工作目录执行。建议在配置中开启 `allow_approve: true`，在执行 `rm -rf`、`git push --force` 等危险命令前会请求确认。
</details>

<details>
<summary><b>Q: 支持哪些操作系统？</b></summary>

Linux、macOS、Windows（通过 Git Bash 或 WSL）。`file` 和 `shell` 工具在所有平台可用；`browser` 工具需要额外安装 Chrome。
</details>

<details>
<summary><b>Q: 可以使用本地模型吗？</b></summary>

可以。运行 Ollama 或其他兼容服务后，配置指向本地地址：

```yaml
llm:
  base_url: http://localhost:11434/v1
  model: qwen2.5:14b
  api_key: "ollama"   # Ollama 不需要真实 key，但必填
```

注意：本地模型需要支持 Function Calling。
</details>

<details>
<summary><b>Q: 如何调试 LLM 调用过程？</b></summary>

在配置文件中设置 `agent.verbose: true`，或在命令行中加 `-v`：

```bash
issgo run "你的任务" -v
```

这会在终端打印完整的 Prompt、LLM 响应和 Tool 调用详情。
</details>

---

## License

[MIT](LICENSE)

---

<p align="center">
  <sub>Made with ❤️ by the IssGo team</sub>
</p>
