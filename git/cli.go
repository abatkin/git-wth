package git

import (
	"os/exec"
	"strings"
)

type Cli struct{}

func (Cli) Config(key string) (string, error) {
	out, err := gitOutput("config", key)
	if err != nil {
		return "", nil
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

func (Cli) ConfigRegexp(pattern string) ([]string, error) {
	out, err := gitOutput("config", "--get-regexp", pattern)
	if err != nil {
		return nil, nil
	}
	return splitLines(string(out)), nil
}

func (Cli) Log(format string, revisionRange string) ([]string, error) {
	out, err := gitOutput("log", "--pretty=format:"+format, revisionRange)
	if err != nil {
		return nil, err
	}
	return splitLines(string(out)), nil
}

func (Cli) ShowRef() ([]Ref, error) {
	out, err := gitOutput("show-ref")
	if err != nil {
		return nil, err
	}
	var refs []Ref
	for _, line := range splitLines(string(out)) {
		sha, ref, ok := strings.Cut(line, " refs/")
		if !ok {
			continue
		}
		refs = append(refs, Ref{SHA: sha, Name: ref})
	}
	return refs, nil
}

func (Cli) SymbolicRefHead() (string, error) {
	out, err := gitOutput("symbolic-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/"), nil
}

func (Cli) HasModifiedFiles() (bool, error) {
	out, err := gitOutput("ls-files", "-m")
	if err != nil {
		return false, err
	}
	return len(out) > 0, nil
}

func (Cli) HasStagedChanges() (bool, error) {
	out, err := gitOutput("diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	return len(out) > 0, nil
}

func gitOutput(args ...string) ([]byte, error) {
	return exec.Command("git", args...).Output()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
