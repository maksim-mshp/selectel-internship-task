package loglint

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	overrides, err := decodeConfigOverrides(data)
	if err != nil {
		return cfg, fmt.Errorf("%s: %w", path, err)
	}
	cfg = applyOverrides(cfg, overrides)
	if err := validateConfig(cfg); err != nil {
		return cfg, fmt.Errorf("%s: %w", path, err)
	}
	compiled, err := compilePatterns(cfg.Patterns)
	if err != nil {
		return cfg, fmt.Errorf("%s: %w", path, err)
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

func decodeConfigOverrides(data []byte) (ConfigOverrides, error) {
	var overrides ConfigOverrides
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&overrides); err != nil {
		return overrides, err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return overrides, fmt.Errorf("unexpected extra document")
		}
		return overrides, err
	}
	return overrides, nil
}

func validateConfig(cfg Config) error {
	if cfg.Sensitive && len(cfg.Patterns) == 0 {
		return fmt.Errorf("sensitive enabled but patterns is empty")
	}
	for i, p := range cfg.Patterns {
		if strings.TrimSpace(p) == "" {
			return fmt.Errorf("patterns[%d] is empty", i)
		}
	}
	return nil
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
