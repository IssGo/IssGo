package config

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func Validate(cfg *Config) []error {
	var errs []error

	if cfg.LLM.APIKey == "" {
		errs = append(errs, &ValidationError{
			Field:   "llm.api_key",
			Message: "API key is required. Set it in config or via ISSGO_LLM_API_KEY environment variable.",
		})
	}

	if cfg.LLM.Model == "" {
		errs = append(errs, &ValidationError{
			Field:   "llm.model",
			Message: "Model name is required.",
		})
	}

	if cfg.LLM.BaseURL == "" {
		errs = append(errs, &ValidationError{
			Field:   "llm.base_url",
			Message: "Base URL is required.",
		})
	}

	if cfg.LLM.Temperature < 0 || cfg.LLM.Temperature > 2 {
		errs = append(errs, &ValidationError{
			Field:   "llm.temperature",
			Message: "Temperature must be between 0.0 and 2.0.",
		})
	}

	if cfg.LLM.MaxTokens < 1 {
		errs = append(errs, &ValidationError{
			Field:   "llm.max_tokens",
			Message: "Max tokens must be at least 1.",
		})
	}

	if cfg.LLM.TimeoutSecs < 1 {
		cfg.LLM.TimeoutSecs = 120
	}

	if cfg.Agent.MaxSteps < 1 {
		cfg.Agent.MaxSteps = 30
	}

	if cfg.Server.Enabled {
		if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
			errs = append(errs, &ValidationError{
				Field:   "server.port",
				Message: "Port must be between 1 and 65535.",
			})
		}
	}

	return errs
}

func FormatErrors(errs []error) string {
	var sb strings.Builder
	sb.WriteString("Configuration errors:\n")
	for _, e := range errs {
		sb.WriteString(fmt.Sprintf("  - %v\n", e))
	}
	return sb.String()
}
