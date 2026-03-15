package loglint

import (
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Lowercase        bool
	English          bool
	Special          bool
	Sensitive        bool
	Patterns         []string
	compiledPatterns []*regexp.Regexp
}

type ConfigOverrides struct {
	Lowercase *bool    `json:"lowercase" yaml:"lowercase"`
	English   *bool    `json:"english" yaml:"english"`
	Special   *bool    `json:"special" yaml:"special"`
	Sensitive *bool    `json:"sensitive" yaml:"sensitive"`
	Patterns  []string `json:"patterns" yaml:"patterns"`
}

func DefaultConfig() Config {
	cfg := Config{
		Lowercase: true,
		English:   true,
		Special:   true,
		Sensitive: true,
	}
	cfg.compiledPatterns, _ = compilePatterns(cfg.Patterns)
	return cfg
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
	cfg = applyOverrides(cfg, overrides)
	compiled, err := compilePatterns(cfg.Patterns)
	if err != nil {
		return cfg, err
	}
	cfg.compiledPatterns = compiled
	return cfg, nil
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
	if overrides.Patterns != nil {
		cfg.Patterns = overrides.Patterns
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

func compilePatterns(patterns []string) ([]*regexp.Regexp, error) {
	if patterns == nil {
		return nil, nil
	}
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		r, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}
