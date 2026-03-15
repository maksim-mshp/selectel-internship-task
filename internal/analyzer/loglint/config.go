package loglint

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Lowercase         bool
	English           bool
	Special           bool
	Sensitive         bool
	SensitiveKeywords []string
}

type ConfigOverrides struct {
	Lowercase         *bool    `json:"lowercase" yaml:"lowercase"`
	English           *bool    `json:"english" yaml:"english"`
	Special           *bool    `json:"special" yaml:"special"`
	Sensitive         *bool    `json:"sensitive" yaml:"sensitive"`
	SensitiveKeywords []string `json:"sensitive_keywords" yaml:"sensitive_keywords"`
}

func DefaultConfig() Config {
	return Config{
		Lowercase:         true,
		English:           true,
		Special:           true,
		Sensitive:         true,
		SensitiveKeywords: defaultSensitiveKeywords(),
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		path = findDefaultConfigPath()
		if path == "" {
			return cfg, nil
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	var overrides ConfigOverrides
	if err := yaml.Unmarshal(data, &overrides); err != nil {
		return cfg, err
	}
	return applyOverrides(cfg, overrides), nil
}

func applyOverrides(cfg Config, overrides ConfigOverrides) Config {
	if overrides.Lowercase != nil {
		cfg.Lowercase = *overrides.Lowercase
	}
	if overrides.English != nil {
		cfg.English = *overrides.English
	}
	if overrides.Special != nil {
		cfg.Special = *overrides.Special
	}
	if overrides.Sensitive != nil {
		cfg.Sensitive = *overrides.Sensitive
	}
	if overrides.SensitiveKeywords != nil {
		cfg.SensitiveKeywords = overrides.SensitiveKeywords
	}
	return cfg
}

func findDefaultConfigPath() string {
	names := []string{".loglint.yml", ".loglint.yaml", ".loglint.json"}
	for _, name := range names {
		if _, err := os.Stat(name); err == nil {
			return filepath.Clean(name)
		}
	}
	return ""
}

func defaultSensitiveKeywords() []string {
	return []string{
		"password",
		"passwd",
		"pwd",
		"secret",
		"token",
		"access_token",
		"refresh_token",
		"id_token",
		"api_key",
		"apikey",
		"access_key",
		"private_key",
		"client_secret",
		"authorization",
		"bearer",
		"session",
		"cookie",
		"ssn",
		"cvv",
		"pin",
	}
}
