package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

const (
	configFileName        = ".git-wthrc"
	legacyConfigFileName  = ".git-wtfrc"
	colorConfigKey        = "color.wth"
	legacyColorConfigKey  = "color.wtf"
	defaultColorConfigKey = "color.ui"
)

type Config struct {
	IntegrationBranches []string
	Ignore              []string
	MaxCommits          int
}

type configYAML struct {
	IntegrationBranches *branchList `yaml:"integration-branches"`
	Versions            *branchList `yaml:"versions"`
	Ignore              *branchList `yaml:"ignore"`
	MaxCommits          *int        `yaml:"max_commits"`
}

type branchList []string

func defaultConfig() Config {
	return Config{
		IntegrationBranches: []string{"heads/main", "heads/master", "heads/next", "heads/edge"},
		Ignore:              []string{},
		MaxCommits:          5,
	}
}

func loadConfig() (Config, error) {
	config := defaultConfig()
	path, err := findConfigFile()
	if err != nil || path == "" {
		return config, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	parsed, err := parseConfigYAML(data)
	if err != nil {
		return config, err
	}
	if parsed.IntegrationBranches != nil {
		config.IntegrationBranches = normalizeConfigBranchNames(*parsed.IntegrationBranches)
	} else if parsed.Versions != nil {
		config.IntegrationBranches = normalizeConfigBranchNames(*parsed.Versions)
	}
	if parsed.Ignore != nil {
		config.Ignore = normalizeConfigBranchNames(*parsed.Ignore)
	}
	if parsed.MaxCommits != nil {
		config.MaxCommits = *parsed.MaxCommits
	}
	return config, nil
}

func normalizeConfigBranchNames(branches []string) []string {
	normalized := make([]string, 0, len(branches))
	for _, branch := range branches {
		normalized = append(normalized, normalizeConfigBranchName(branch))
	}
	return normalized
}

func normalizeConfigBranchName(branch string) string {
	return strings.TrimPrefix(branch, "remotes/")
}

func findConfigFile() (string, error) {
	path, err := findFile(configFileName)
	if err != nil || path != "" {
		return path, err
	}
	return findFile(legacyConfigFileName)
}

func findFile(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func parseConfigYAML(data []byte) (configYAML, error) {
	var parsed configYAML
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return configYAML{}, err
	}
	return parsed, nil
}

func (l *branchList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		if value.Tag == "!!null" || value.Value == "" {
			*l = branchList{}
			return nil
		}
		*l = branchList{value.Value}
		return nil
	}
	var branches []string
	if err := value.Decode(&branches); err != nil {
		return err
	}
	*l = branches
	return nil
}

func (c Config) toYAML() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("integration-branches:\n")
	for _, branch := range c.IntegrationBranches {
		fmt.Fprintf(&b, "- %s\n", branch)
	}
	b.WriteString("ignore:\n")
	for _, branch := range c.Ignore {
		fmt.Fprintf(&b, "- %s\n", branch)
	}
	fmt.Fprintf(&b, "max_commits: %d\n", c.MaxCommits)
	return b.String()
}
