package main

import (
	"git-wth/git"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestLoadConfigNormalizesRemoteBranchNames(t *testing.T) {
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

	configData := `integration-branches:
- heads/main
- remotes/upstream/main
- origin/stable
ignore: [remotes/origin/topic, heads/tmp, origin/wip]
`
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte(configData), 0o644); err != nil {
		t.Fatalf("WriteFile config returned error: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if got, want := config.IntegrationBranches, []string{"heads/main", "upstream/main", "origin/stable"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("IntegrationBranches = %v, want %v", got, want)
	}
	if got, want := config.Ignore, []string{"origin/topic", "heads/tmp", "origin/wip"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Ignore = %v, want %v", got, want)
	}
}

func TestLoadConfigNormalizesLegacyVersionsBranchNames(t *testing.T) {
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

	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte("versions: [remotes/upstream/main]\n"), 0o644); err != nil {
		t.Fatalf("WriteFile config returned error: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if got, want := config.IntegrationBranches, []string{"upstream/main"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("IntegrationBranches = %v, want %v", got, want)
	}
}

func TestAheadBehindString(t *testing.T) {
	got := aheadBehindString([]string{"a", "b"}, []string{"c"})
	want := "2 commits ahead; 1 commit behind"
	if got != want {
		t.Fatalf("aheadBehindString = %q, want %q", got, want)
	}
}

func TestShowRelationsIntegrationLocalOnlyFeatureHasNoBehindCount(t *testing.T) {
	logCalls := []string{}
	app := &App{
		git: fakeGit{
			logs: map[string][]string{
				"heads/main..heads/topic": {"topic ahead"},
			},
			logCalls: &logCalls,
		},
		opts: Options{Short: true},
		config: Config{
			IntegrationBranches: []string{"heads/main"},
		},
	}
	main := &Branch{Name: "main", LocalBranch: "heads/main"}
	topic := &Branch{Name: "topic", LocalBranch: "heads/topic"}

	output := captureStdout(t, func() {
		if err := app.showRelations(main, map[string]*Branch{"main": main, "topic": topic}); err != nil {
			t.Fatalf("showRelations returned error: %v", err)
		}
	})

	if containsString(logCalls, "heads/topic..heads/topic") {
		t.Fatalf("Log calls included self-comparison: %v", logCalls)
	}
	if strings.Contains(output, "behind") {
		t.Fatalf("output contains behind count, want none:\n%s", output)
	}
}

func TestShowRelationsIntegrationRemoteOnlyFeatureUsesRemoteAheadCommits(t *testing.T) {
	logCalls := []string{}
	app := &App{
		git: fakeGit{
			logs: map[string][]string{
				"origin/main..origin/topic": {"remote ahead 1", "remote ahead 2"},
				"heads/main..origin/topic":  {"local ahead"},
			},
			logCalls: &logCalls,
		},
		opts: Options{Short: true},
		config: Config{
			IntegrationBranches: []string{"heads/main"},
		},
	}
	main := &Branch{Name: "main", LocalBranch: "heads/main", RemoteBranch: "origin/main"}
	topic := &Branch{Name: "topic", RemoteBranch: "origin/topic"}

	output := captureStdout(t, func() {
		if err := app.showRelations(main, map[string]*Branch{"main": main, "topic": topic}); err != nil {
			t.Fatalf("showRelations returned error: %v", err)
		}
	})

	if containsString(logCalls, "origin/topic..origin/topic") {
		t.Fatalf("Log calls included self-comparison: %v", logCalls)
	}
	if !strings.Contains(output, "2 commits ahead") {
		t.Fatalf("output = %q, want remote-ahead count", output)
	}
	if strings.Contains(output, "behind") {
		t.Fatalf("output contains behind count, want none:\n%s", output)
	}
}

func TestShowRelationsIntegrationLocalAndRemoteFeatureBehindUsesRemote(t *testing.T) {
	logCalls := []string{}
	app := &App{
		git: fakeGit{
			logs: map[string][]string{
				"heads/main..heads/topic":   {"topic ahead"},
				"heads/topic..origin/topic": {"topic behind"},
			},
			logCalls: &logCalls,
		},
		opts: Options{Short: true},
		config: Config{
			IntegrationBranches: []string{"heads/main"},
		},
	}
	main := &Branch{Name: "main", LocalBranch: "heads/main"}
	topic := &Branch{Name: "topic", LocalBranch: "heads/topic", RemoteBranch: "origin/topic"}

	output := captureStdout(t, func() {
		if err := app.showRelations(main, map[string]*Branch{"main": main, "topic": topic}); err != nil {
			t.Fatalf("showRelations returned error: %v", err)
		}
	})

	if !containsString(logCalls, "heads/topic..origin/topic") {
		t.Fatalf("Log calls = %v, want heads/topic..origin/topic", logCalls)
	}
	if containsString(logCalls, "heads/topic..heads/topic") {
		t.Fatalf("Log calls included self-comparison: %v", logCalls)
	}
	if !strings.Contains(output, "1 commit behind") {
		t.Fatalf("output = %q, want behind count", output)
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
	config   map[string]string
	logs     map[string][]string
	logCalls *[]string
}

func (f fakeGit) Config(key string) (string, error) {
	return f.config[key], nil
}

func (f fakeGit) ConfigRegexp(pattern string) ([]string, error) {
	return nil, nil
}

func (f fakeGit) Log(format string, revisionRange string) ([]string, error) {
	if f.logCalls != nil {
		*f.logCalls = append(*f.logCalls, revisionRange)
	}
	return f.logs[revisionRange], nil
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

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe returned error: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("Close writer returned error: %v", err)
	}
	os.Stdout = oldStdout
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close reader returned error: %v", err)
	}
	return string(out)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
