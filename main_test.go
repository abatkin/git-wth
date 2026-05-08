package main

import (
	"git-wth/git"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseOptions(t *testing.T) {
	opts, err := parseOptions([]string{"--long", "-A", "--relations", "heads/topic"})
	if err != nil {
		t.Fatalf("parseOptions returned error: %v", err)
	}
	if !opts.Long || !opts.AllCommits || !opts.ShowRelations {
		t.Fatalf("expected long, all commits, and relations flags: %+v", opts)
	}
	if got, want := opts.Branches, []string{"heads/topic"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("branches = %v, want %v", got, want)
	}
}

func TestParseOptionsUnknownLongFlag(t *testing.T) {
	if _, err := parseOptions([]string{"--wat"}); err == nil {
		t.Fatal("expected unknown long flag to return an error")
	}
}

func TestParseSimpleYAML(t *testing.T) {
	parsed := parseSimpleYAML(`---
integration-branches:
- heads/main
- heads/release
ignore: [heads/tmp, "origin/wip"]
max_commits: 12
`)

	if got, want := parsed.stringList("integration-branches"), []string{"heads/main", "heads/release"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("integration branches = %v, want %v", got, want)
	}
	if got, want := parsed.stringList("ignore"), []string{"heads/tmp", "origin/wip"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ignore = %v, want %v", got, want)
	}
	if got, ok := parsed.intValue("max_commits"); !ok || got != 12 {
		t.Fatalf("max_commits = %d, %v; want 12, true", got, ok)
	}
}

func TestLoadConfigPrefersWthConfigFile(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Chdir cleanup returned error: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, legacyConfigFileName), []byte("max_commits: 2\n"), 0o644); err != nil {
		t.Fatalf("WriteFile legacy config returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte("max_commits: 7\n"), 0o644); err != nil {
		t.Fatalf("WriteFile config returned error: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if got, want := config.MaxCommits, 7; got != want {
		t.Fatalf("MaxCommits = %d, want %d", got, want)
	}
}

func TestLoadConfigFallsBackToLegacyWtfConfigFile(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Chdir cleanup returned error: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, legacyConfigFileName), []byte("max_commits: 3\n"), 0o644); err != nil {
		t.Fatalf("WriteFile legacy config returned error: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if got, want := config.MaxCommits, 3; got != want {
		t.Fatalf("MaxCommits = %d, want %d", got, want)
	}
}

func TestAheadBehindString(t *testing.T) {
	got := aheadBehindString([]string{"a", "b"}, []string{"c"})
	want := "2 commits ahead; 1 commit behind"
	if got != want {
		t.Fatalf("aheadBehindString = %q, want %q", got, want)
	}
}

func TestWantColorPrefersWthConfig(t *testing.T) {
	app := &App{git: fakeGit{config: map[string]string{
		colorConfigKey:       "false",
		legacyColorConfigKey: "true",
	}}}
	if app.wantColor() {
		t.Fatal("wantColor returned true, want false from color.wth")
	}
}

func TestWantColorFallsBackToLegacyWtfConfig(t *testing.T) {
	app := &App{git: fakeGit{config: map[string]string{
		legacyColorConfigKey: "true",
	}}}
	if !app.wantColor() {
		t.Fatal("wantColor returned false, want true from color.wtf")
	}
}

type fakeGit struct {
	config map[string]string
}

func (f fakeGit) Config(key string) (string, error) {
	return f.config[key], nil
}

func (f fakeGit) ConfigRegexp(pattern string) ([]string, error) {
	return nil, nil
}

func (f fakeGit) Log(format string, revisionRange string) ([]string, error) {
	return nil, nil
}

func (f fakeGit) ShowRef() ([]git.Ref, error) {
	return nil, nil
}

func (f fakeGit) SymbolicRefHead() (string, error) {
	return "", nil
}

func (f fakeGit) HasModifiedFiles() (bool, error) {
	return false, nil
}

func (f fakeGit) HasStagedChanges() (bool, error) {
	return false, nil
}
