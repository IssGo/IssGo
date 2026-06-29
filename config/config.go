package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type LLMConfig struct {
	Provider string `mapstructure:"provider" yaml:"provider"`
	Model    string `mapstructure:"model"    yaml:"model"`
	APIKey   string `mapstructure:"api_key"  yaml:"api_key"`
	BaseURL  string `mapstructure:"base_url" yaml:"base_url"`
}

type ToolsConfig struct {
	Shell   bool `mapstructure:"shell"   yaml:"shell"`
	File    bool `mapstructure:"file"    yaml:"file"`
	Web     bool `mapstructure:"web"     yaml:"web"`
	Browser bool `mapstructure:"browser" yaml:"browser"`
}

type AgentConfig struct {
	MaxSteps     int  `mapstructure:"max_steps"     yaml:"max_steps"`
	AllowApprove bool `mapstructure:"allow_approve" yaml:"allow_approve"`
	Verbose      bool `mapstructure:"verbose"       yaml:"verbose"`
}

type Config struct {
	LLM   LLMConfig   `mapstructure:"llm"   yaml:"llm"`
	Tools ToolsConfig `mapstructure:"tools" yaml:"tools"`
	Agent AgentConfig `mapstructure:"agent" yaml:"agent"`
}

func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider: "deepseek",
			Model:    "deepseek-chat",
			BaseURL:  "https://api.deepseek.com",
		},
		Tools: ToolsConfig{
			Shell:   true,
			File:    true,
			Web:     true,
			Browser: false,
		},
		Agent: AgentConfig{
			MaxSteps:     20,
			AllowApprove: true,
			Verbose:      false,
		},
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()

	v.SetConfigName(".issgo")
	v.SetConfigType("yaml")

	// 当前目录
	v.AddConfigPath(".")

	// 用户主目录
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(home)
	}

	v.SetEnvPrefix("ISSGO")
	v.AutomaticEnv()
	v.BindEnv("llm.api_key", "ISSGO_LLM_API_KEY")
	v.BindEnv("llm.model", "ISSGO_LLM_MODEL")
	v.BindEnv("llm.provider", "ISSGO_LLM_PROVIDER")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func WriteDefault(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	v := viper.New()
	v.Set("llm", DefaultConfig().LLM)
	v.Set("tools", DefaultConfig().Tools)
	v.Set("agent", DefaultConfig().Agent)
	return v.WriteConfigAs(path)
}
