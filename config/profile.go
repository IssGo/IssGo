package config

import "fmt"

type Profile struct {
	Name     string `mapstructure:"name"      yaml:"name"`
	Provider string `mapstructure:"provider"  yaml:"provider"`
	Model    string `mapstructure:"model"     yaml:"model"`
	APIKey   string `mapstructure:"api_key"   yaml:"api_key"`
	BaseURL  string `mapstructure:"base_url"  yaml:"base_url"`
}

func ListProfiles(cfg *Config) []string {
	names := make([]string, 0, len(cfg.Profiles))
	for _, p := range cfg.Profiles {
		names = append(names, p.Name)
	}
	return names
}

func GetProfile(cfg *Config, name string) (*Profile, error) {
	for _, p := range cfg.Profiles {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("profile %q not found", name)
}

func ProfileToConfig(p Profile) *Config {
	cfg := DefaultConfig()
	if p.Provider != "" {
		cfg.LLM.Provider = p.Provider
	}
	if p.Model != "" {
		cfg.LLM.Model = p.Model
	}
	if p.APIKey != "" {
		cfg.LLM.APIKey = p.APIKey
	}
	if p.BaseURL != "" {
		cfg.LLM.BaseURL = p.BaseURL
	}
	return cfg
}
