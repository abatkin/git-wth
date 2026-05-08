package main

import (
	"embed"
	"fmt"
	"git-wth/ui"
	"git-wth/util"
	"os"
	"slices"
	"strings"

	"git-wth/git"
)

//go:embed text/*.txt
var textFS embed.FS

var (
	help  string
	key   string
	usage string
)

func init() {
	help = loadText("text/help.txt")
	key = loadText("text/key.txt")
	usage = loadText("text/usage.txt")
}

func loadText(path string) string {
	b, err := textFS.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

type Branch struct {
	Name             string
	Remote           string
	RemoteURL        string
	RemoteMergepoint string
	LocalBranch      string
	RemoteBranch     string
	Ignore           bool
}

type App struct {
	git    git.Git
	opts   Options
	config Config
	ui     ui.ColorManager
}

func main() {
	if err := run(os.Args[1:], git.Cli{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, g git.Git) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	if opts.Help {
		return nil
	}

	config, err := loadConfig()
	if err != nil {
		return err
	}
	app := &App{git: g, opts: opts, config: config}
	app.ui.Enabled = app.wantColor()

	if opts.DumpConfig {
		fmt.Print(config.toYAML())
		return nil
	}

	branches, err := app.indexBranches()
	if err != nil {
		return err
	}

	showDirty := len(opts.Branches) == 0
	targetNames := opts.Branches
	if len(targetNames) == 0 {
		current, err := g.SymbolicRefHead()
		if err != nil {
			return err
		}
		targetNames = []string{current}
	}

	var targets []*Branch
	for _, target := range targetNames {
		name := strings.TrimPrefix(target, "heads/")
		branch := branches[name]
		if branch == nil {
			return fmt.Errorf("Error: can't find branch %q.", name)
		}
		targets = append(targets, branch)
	}

	for _, target := range targets {
		if err := app.show(target); err != nil {
			return err
		}
		if opts.ShowRelations || target.RemoteBranch == "" {
			if err := app.showRelations(target, branches); err != nil {
				return err
			}
		}
	}

	modified := false
	uncommitted := false
	if showDirty {
		modified, err = g.HasModifiedFiles()
		if err != nil {
			return err
		}
		uncommitted, err = g.HasStagedChanges()
		if err != nil {
			return err
		}
	}

	if opts.Key {
		fmt.Println()
		fmt.Println(key)
	}
	if modified || uncommitted {
		fmt.Println()
	}
	if modified {
		fmt.Println(app.ui.Red("NOTE") + ": working directory contains modified files.")
	}
	if uncommitted {
		fmt.Println(app.ui.Red("NOTE") + ": staging area contains staged but uncommitted files.")
	}

	return nil
}

func (a *App) indexBranches() (map[string]*Branch, error) {
	remoteLines, err := a.git.ConfigRegexp(`^remote\..*\.url`)
	if err != nil {
		return nil, err
	}
	remotes := map[string]string{}
	for _, line := range remoteLines {
		if strings.HasPrefix(line, "remote.") {
			rest := strings.TrimPrefix(line, "remote.")
			nameURL := strings.SplitN(rest, ".url ", 2)
			if len(nameURL) == 2 {
				remotes[nameURL[0]] = nameURL[1]
			}
		}
	}

	branchLines, err := a.git.ConfigRegexp(`^branch\.`)
	if err != nil {
		return nil, err
	}
	branches := map[string]*Branch{}
	for _, line := range branchLines {
		switch {
		case strings.Contains(line, ".remote "):
			name, value, ok := parseBranchConfigLine(line, ".remote ")
			if !ok {
				continue
			}
			branch := ensureBranch(branches, name)
			branch.Remote = value
			branch.RemoteURL = remotes[value]
		case strings.Contains(line, ".merge "):
			name, value, ok := parseBranchConfigLine(line, ".merge ")
			if !ok {
				continue
			}
			value = strings.TrimPrefix(value, "refs/")
			value = strings.TrimPrefix(value, "heads/")
			branch := ensureBranch(branches, name)
			branch.RemoteMergepoint = value
		}
	}

	refs, err := a.git.ShowRef()
	if err != nil {
		return nil, err
	}
	remoteBranches := map[string]bool{}
	for _, ref := range refs {
		if name, ok := strings.CutPrefix(ref.Name, "heads/"); ok {
			if name == "HEAD" {
				continue
			}
			branch := ensureBranch(branches, name)
			branch.Name = name
			branch.LocalBranch = ref.Name
			continue
		}
		remoteRef, ok := strings.CutPrefix(ref.Name, "remotes/")
		if !ok {
			continue
		}
		remote, branchName, ok := strings.Cut(remoteRef, "/")
		if !ok {
			continue
		}
		remoteBranches[remote+"/"+branchName] = true
		if branchName == "HEAD" {
			continue
		}
		ignore := !(a.opts.All || remote == "origin")
		mapName := branchName
		if existing := branches[branchName]; existing == nil || existing.Remote != remote {
			mapName = remote + "/" + branchName
		}
		branch := ensureBranch(branches, mapName)
		branch.Name = mapName
		branch.Remote = remote
		branch.RemoteBranch = remote + "/" + branchName
		branch.RemoteURL = remotes[remote]
		branch.Ignore = ignore
	}

	for _, branch := range branches {
		if branch.Remote == "" || branch.RemoteMergepoint == "" {
			continue
		}
		if branch.Remote == "." {
			branch.RemoteBranch = branch.RemoteMergepoint
		} else {
			remoteBranch := branch.Remote + "/" + branch.RemoteMergepoint
			if remoteBranches[remoteBranch] {
				branch.RemoteBranch = remoteBranch
			}
		}
	}

	return branches, nil
}

func parseBranchConfigLine(line, marker string) (string, string, bool) {
	if !strings.HasPrefix(line, "branch.") {
		return "", "", false
	}
	rest := strings.TrimPrefix(line, "branch.")
	name, value, ok := strings.Cut(rest, marker)
	return name, value, ok
}

func ensureBranch(branches map[string]*Branch, name string) *Branch {
	if branches[name] == nil {
		branches[name] = &Branch{Name: name}
	}
	return branches[name]
}

func (a *App) show(b *Branch) error {
	haveBoth := b.LocalBranch != "" && b.RemoteBranch != ""
	var pushc, pullc []string
	oosync := false
	var err error
	if haveBoth {
		pushc, err = a.commitsBetween(b.RemoteBranch, b.LocalBranch)
		if err != nil {
			return err
		}
		pullc, err = a.commitsBetween(b.LocalBranch, b.RemoteBranch)
		if err != nil {
			return err
		}
		oosync = len(pushc) > 0 && len(pullc) > 0
	}

	if b.LocalBranch != "" {
		fmt.Println("Local branch: " + a.ui.Green(strings.TrimPrefix(b.LocalBranch, "heads/")))
		if haveBoth {
			if len(pushc) == 0 {
				fmt.Println(a.widget(true, false, false, false) + " in sync with remote")
			} else {
				action := "push"
				if oosync {
					action = "push after rebase / merge"
				}
				fmt.Println(a.widget(false, false, false, false) + " NOT in sync with remote (you should " + action + ")")
				if !a.opts.Short {
					a.showCommits(pushc, "    ")
				}
			}
		}
	}

	if b.RemoteBranch != "" {
		fmt.Printf("Remote branch: %s (%s)\n", a.ui.Cyan(b.RemoteBranch), b.RemoteURL)
		if haveBoth {
			if len(pullc) == 0 {
				fmt.Println(a.widget(true, false, false, false) + " in sync with local")
			} else {
				action := "rebase / merge"
				if len(pushc) == 0 {
					action = "merge"
				}
				fmt.Println(a.widget(false, false, false, false) + " NOT in sync with local (you should " + action + ")")
				if !a.opts.Short {
					a.showCommits(pullc, "    ")
				}
			}
		}
	}
	if oosync {
		fmt.Println()
		fmt.Println(a.ui.Red("WARNING") + ": local and remote branches have diverged. A merge will occur unless you rebase.")
	}
	return nil
}

func (a *App) showRelations(b *Branch, allBranches map[string]*Branch) error {
	var ibs []*Branch
	var fbs []*Branch
	for _, branch := range allBranches {
		if a.isIntegrationBranch(branch.LocalBranch) || a.isIntegrationBranch(branch.RemoteBranch) {
			ibs = append(ibs, branch)
		} else {
			fbs = append(fbs, branch)
		}
	}

	if a.isIntegrationBranch(b.LocalBranch) {
		if len(fbs) > 0 {
			fmt.Println()
			fmt.Println("Feature branches:")
		}
		for _, br := range fbs {
			if a.shouldIgnore(br) {
				continue
			}
			localOnly := br.RemoteBranch == ""
			remoteOnly := br.LocalBranch == ""
			name := br.Name
			switch {
			case localOnly:
				name = a.ui.Purple(name)
			case remoteOnly:
				name = a.ui.Cyan(name)
			default:
				name = a.ui.Green(name)
			}
			head := br.LocalBranch
			if remoteOnly {
				head = br.RemoteBranch
			}
			remoteAhead := []string{}
			if b.RemoteBranch != "" {
				var err error
				remoteAhead, err = a.commitsBetween(b.RemoteBranch, head)
				if err != nil {
					return err
				}
			}
			localAhead := []string{}
			if b.LocalBranch != "" {
				var err error
				localAhead, err = a.commitsBetween(b.LocalBranch, head)
				if err != nil {
					return err
				}
			}
			localOnlyText := ""
			if localOnly {
				localOnlyText = "(local-only) "
			}
			if len(localAhead) == 0 && len(remoteAhead) == 0 {
				fmt.Printf("%s %s %sis merged in\n", a.widget(true, remoteOnly, localOnly, false), name, localOnlyText)
			} else if len(localAhead) == 0 {
				fmt.Printf("%s %s merged in (only locally)\n", a.widget(true, remoteOnly, localOnly, true), name)
			} else {
				behind, err := a.commitsBetween(head, util.FirstNonEmpty(br.LocalBranch, br.RemoteBranch))
				if err != nil {
					return err
				}
				ahead := localAhead
				if remoteOnly {
					ahead = remoteAhead
				}
				fmt.Printf("%s %s %sis NOT merged in (%s)\n", a.widget(false, remoteOnly, localOnly, false), name, localOnlyText, aheadBehindString(ahead, behind))
				if !a.opts.Short {
					a.showCommits(ahead, "    ")
				}
			}
		}
	} else {
		if len(ibs) > 0 {
			fmt.Println()
			fmt.Println("Integration branches:")
		}
		sortBranchesByName(ibs)
		for _, br := range ibs {
			if a.shouldIgnore(br) {
				continue
			}
			localOnly := br.RemoteBranch == ""
			remoteOnly := br.LocalBranch == ""
			name := br.Name
			if remoteOnly {
				name = a.ui.Cyan(name)
			} else {
				name = a.ui.Green(name)
			}
			ahead, err := a.commitsBetween(br.Name, util.FirstNonEmpty(b.LocalBranch, b.RemoteBranch))
			if err != nil {
				return err
			}
			if len(ahead) == 0 {
				fmt.Printf("%s merged into %s\n", a.widget(true, localOnly, false, false), name)
			} else {
				fmt.Printf("%s NOT merged into %s (%s ahead)\n", a.widget(false, localOnly, false, false), name, util.Pluralize(len(ahead), "commit"))
				if !a.opts.Short {
					a.showCommits(ahead, "    ")
				}
			}
		}
	}
	return nil
}

func sortBranchesByName(branches []*Branch) {
	slices.SortFunc(branches, func(a, b *Branch) int {
		return strings.Compare(a.Name, b.Name)
	})
}

func (a *App) shouldIgnore(branch *Branch) bool {
	return branch.Ignore || util.Contains(a.config.Ignore, branch.LocalBranch) || util.Contains(a.config.Ignore, branch.RemoteBranch)
}

func (a *App) isIntegrationBranch(branch string) bool {
	return util.Contains(a.config.IntegrationBranches, branch)
}

func (a *App) commitsBetween(from, to string) ([]string, error) {
	format := "- %s [" + a.ui.Yellow("%h") + "]"
	if a.opts.Long {
		format += " (" + a.ui.Purple("%ae") + "; %ar)"
	}
	return a.git.Log(format, from+".."+to)
}

func (a *App) showCommits(commits []string, prefix string) {
	if len(commits) == 0 {
		fmt.Println(prefix + " none")
		return
	}
	max := a.config.MaxCommits
	if a.opts.AllCommits {
		max = len(commits)
	}
	if max == len(commits)-1 {
		max--
	}
	if max < 0 {
		max = 0
	}
	if max > len(commits) {
		max = len(commits)
	}
	for _, commit := range commits[:max] {
		fmt.Println(prefix + commit)
	}
	if len(commits) > max {
		fmt.Println(a.ui.Grey(fmt.Sprintf("%s... and %d more (use -A to see all).", prefix, len(commits)-max)))
	}
}

func aheadBehindString(ahead, behind []string) string {
	parts := []string{}
	if len(ahead) > 0 {
		parts = append(parts, util.Pluralize(len(ahead), "commit")+" ahead")
	}
	if len(behind) > 0 {
		parts = append(parts, util.Pluralize(len(behind), "commit")+" behind")
	}
	return strings.Join(parts, "; ")
}

func (a *App) widget(mergedIn bool, remoteOnly bool, localOnly bool, localOnlyMerge bool) string {
	left, right := "[", "]"
	switch {
	case remoteOnly:
		left, right = "{", "}"
	case localOnly:
		left, right = "(", ")"
	}
	middle := " "
	if mergedIn && localOnlyMerge {
		middle = a.ui.Green("~")
	} else if mergedIn {
		middle = a.ui.Green("x")
	}
	return left + middle + right
}

func (a *App) wantColor() bool {
	want, _ := a.git.Config(colorConfigKey)
	if want == "" {
		want, _ = a.git.Config(legacyColorConfigKey)
	}
	if want == "" {
		want, _ = a.git.Config(defaultColorConfigKey)
	}
	switch strings.TrimSpace(want) {
	case "true", "yes", "always":
		return true
	case "auto":
		return ui.StdoutIsTerminal()
	default:
		return false
	}
}
