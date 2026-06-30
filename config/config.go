package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ─── Sub-configs ───────────────────────────────────────────────

type LLMConfig struct {
	Provider     string            `mapstructure:"provider"      yaml:"provider"`
	Model        string            `mapstructure:"model"         yaml:"model"`
	APIKey       string            `mapstructure:"api_key"       yaml:"api_key"`
	BaseURL      string            `mapstructure:"base_url"      yaml:"base_url"`
	Temperature  float64           `mapstructure:"temperature"   yaml:"temperature"`
	MaxTokens    int               `mapstructure:"max_tokens"    yaml:"max_tokens"`
	TimeoutSecs  int               `mapstructure:"timeout_secs"  yaml:"timeout_secs"`
	Headers      map[string]string `mapstructure:"headers"       yaml:"headers"`
	RetryCount   int               `mapstructure:"retry_count"   yaml:"retry_count"`
}

type ToolsConfig struct {
	Shell     bool   `mapstructure:"shell"      yaml:"shell"`
	File      bool   `mapstructure:"file"       yaml:"file"`
	Web       bool   `mapstructure:"web"        yaml:"web"`
	Browser   bool   `mapstructure:"browser"    yaml:"browser"`
	Git       bool   `mapstructure:"git"        yaml:"git"`
	Search    bool   `mapstructure:"search"     yaml:"search"`
	Plugins   bool   `mapstructure:"plugins"    yaml:"plugins"`
	PluginsDir string `mapstructure:"plugins_dir" yaml:"plugins_dir"`
}

type AgentConfig struct {
	MaxSteps      int    `mapstructure:"max_steps"      yaml:"max_steps"`
	AllowApprove  bool   `mapstructure:"allow_approve"  yaml:"allow_approve"`
	Verbose       bool   `mapstructure:"verbose"        yaml:"verbose"`
	Streaming     bool   `mapstructure:"streaming"      yaml:"streaming"`
	Reflector     bool   `mapstructure:"reflector"      yaml:"reflector"`
	MaxRetries    int    `mapstructure:"max_retries"     yaml:"max_retries"`
	SystemPrompt  string `mapstructure:"system_prompt"   yaml:"system_prompt"`
	SessionDir    string `mapstructure:"session_dir"     yaml:"session_dir"`
	MaxSessionAge string `mapstructure:"max_session_age" yaml:"max_session_age"`
}

type ServerConfig struct {
	Enabled  bool   `mapstructure:"enabled"  yaml:"enabled"`
	Host     string `mapstructure:"host"     yaml:"host"`
	Port     int    `mapstructure:"port"     yaml:"port"`
	WSEnable bool   `mapstructure:"ws"       yaml:"ws"`
	AuthToken string `mapstructure:"auth_token" yaml:"auth_token"`
}

type Config struct {
	LLM      LLMConfig      `mapstructure:"llm"      yaml:"llm"`
	Tools    ToolsConfig    `mapstructure:"tools"    yaml:"tools"`
	Agent    AgentConfig    `mapstructure:"agent"    yaml:"agent"`
	Server   ServerConfig   `mapstructure:"server"   yaml:"server"`
	Profiles []Profile      `mapstructure:"profiles" yaml:"profiles"`
	Active   string         `mapstructure:"active"   yaml:"active"`
}

// ─── Defaults ──────────────────────────────────────────────────

func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider:    "deepseek",
			Model:       "deepseek-chat",
			BaseURL:     "https://api.deepseek.com",
			Temperature: 0.7,
			MaxTokens:   4096,
			TimeoutSecs: 120,
			RetryCount:  3,
		},
		Tools: ToolsConfig{
			Shell:      true,
			File:       true,
			Web:        true,
			Browser:    false,
			Git:        true,
			Search:     true,
			Plugins:    false,
			PluginsDir: "~/.issgo/plugins",
		},
		Agent: AgentConfig{
			MaxSteps:      30,
			AllowApprove:  true,
			Verbose:       false,
			Streaming:     true,
			Reflector:     true,
			MaxRetries:    3,
			SessionDir:    "~/.issgo/sessions",
			MaxSessionAge: "168h",
		},
		Server: ServerConfig{
			Enabled: false,
			Host:    "127.0.0.1",
			Port:    8420,
			WSEnable: true,
		},
	}
}

// ─── Load ──────────────────────────────────────────────────────

func Load() (*Config, error) {
	return LoadWithPath("")
}

func LoadWithPath(explicitPath string) (*Config, error) {
	cfg := DefaultConfig()
	v := viper.New()

	v.SetConfigName(".issgo")
	v.SetConfigType("yaml")

	if explicitPath != "" {
		v.SetConfigFile(explicitPath)
	} else {
		v.AddConfigPath(".")
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(home)
		}
	}

	v.SetEnvPrefix("ISSGO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && explicitPath == "" {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// expand paths
	cfg.Tools.PluginsDir = expand(cfg.Tools.PluginsDir)
	cfg.Agent.SessionDir = expand(cfg.Agent.SessionDir)

	// Apply active profile overlay
	if cfg.Active != "" && len(cfg.Profiles) > 0 {
		for _, p := range cfg.Profiles {
			if p.Name == cfg.Active {
				cfg.applyProfile(p)
				break
			}
		}
	}

	return cfg, nil
}

// ─── Helpers ───────────────────────────────────────────────────

func expand(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (c *Config) applyProfile(p Profile) {
	if p.Provider != "" {
		c.LLM.Provider = p.Provider
	}
	if p.Model != "" {
		c.LLM.Model = p.Model
	}
	if p.APIKey != "" {
		c.LLM.APIKey = p.APIKey
	}
	if p.BaseURL != "" {
		c.LLM.BaseURL = p.BaseURL
	}
}

// ─── Write default config file ─────────────────────────────────

func WriteDefault(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	v := viper.New()
	v.Set("llm", DefaultConfig().LLM)
	v.Set("tools", DefaultConfig().Tools)
	v.Set("agent", DefaultConfig().Agent)
	v.Set("server", DefaultConfig().Server)
	return v.WriteConfigAs(path)
}
