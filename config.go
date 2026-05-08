package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	parsed := parseSimpleYAML(string(data))
	if integration := parsed.stringList("integration-branches"); integration != nil {
		config.IntegrationBranches = integration
	} else if versions := parsed.stringList("versions"); versions != nil {
		config.IntegrationBranches = versions
	}
	if ignore := parsed.stringList("ignore"); ignore != nil {
		config.Ignore = ignore
	}
	if maxCommits, ok := parsed.intValue("max_commits"); ok {
		config.MaxCommits = maxCommits
	}
	return config, nil
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

type simpleYAML map[string]any

func parseSimpleYAML(input string) simpleYAML {
	result := simpleYAML{}
	currentKey := ""
	for _, raw := range strings.Split(input, "\n") {
		line := stripYAMLComment(strings.TrimSpace(raw))
		if line == "" || line == "---" {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			if currentKey != "" {
				result[currentKey] = append(result.stringList(currentKey), unquoteYAML(strings.TrimSpace(strings.TrimPrefix(line, "- "))))
			}
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		currentKey = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch {
		case value == "":
			result[currentKey] = []string{}
		case strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]"):
			result[currentKey] = parseInlineYAMLList(value)
		case isInt(value):
			n, _ := strconv.Atoi(value)
			result[currentKey] = n
		default:
			result[currentKey] = unquoteYAML(value)
		}
	}
	return result
}

func stripYAMLComment(line string) string {
	if strings.HasPrefix(line, "#") {
		return ""
	}
	return line
}

func parseInlineYAMLList(value string) []string {
	value = strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, unquoteYAML(strings.TrimSpace(part)))
	}
	return out
}

func unquoteYAML(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func isInt(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

func (y simpleYAML) stringList(key string) []string {
	value, ok := y[key]
	if !ok {
		return nil
	}
	switch v := value.(type) {
	case []string:
		return v
	case string:
		if v == "" {
			return []string{}
		}
		return []string{v}
	default:
		return nil
	}
}

func (y simpleYAML) intValue(key string) (int, bool) {
	value, ok := y[key]
	if !ok {
		return 0, false
	}
	n, ok := value.(int)
	return n, ok
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
