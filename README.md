# git-wth

`git-wth` is Git What The Heck: a Go rewrite of `git-wtf`.

It displays the state of a Git repository in a readable, easy-to-scan format.
Use it to see how a branch relates to its remote, how feature branches relate to
integration branches, and how integration branches relate to feature branches.

## Usage

```sh
git wth [branch+] [options]
```

If no branch is specified, `git-wth` uses the current branch.

Options:

- `-l`, `--long`: include author info and date for each commit
- `-a`, `--all`: show all branches across all remotes, not just `origin`
- `-A`, `--all-commits`: show all commits, not just the configured maximum
- `-s`, `--short`: do not show commits
- `-k`, `--key`: show the output key
- `-r`, `--relations`: show relation to feature and integration branches
- `--dump-config`: print the current configuration and exit

## Configuration

`git-wth` looks for `.git-wthrc` starting in the current directory and then
recursively up to the filesystem root. If `.git-wthrc` is not found, it falls
back to the legacy `.git-wtfrc` filename so it can be used as a drop-in
replacement for `git-wtf`.

To start a configuration file:

```sh
git wth --dump-config > .git-wthrc
```

The config file is a small YAML file with these keys:

- `integration-branches`: branches treated as integration branches
- `ignore`: branches to hide
- `max_commits`: number of commits to display when `--all-commits` is not used

Local branches referenced in config files must be prefixed with `heads/`, for
example `heads/master`. Remote branches must use `remotes/<remote>/<branch>`.

`versions` is also accepted as a legacy alias for `integration-branches`.

## Compatibility With git-wtf

This project aims to be usable as a drop-in replacement for `git-wtf` where
practical. Compatibility support currently includes:

- fallback loading of `.git-wtfrc` when `.git-wthrc` is not present
- fallback reading of `color.wtf` when `color.wth` is not set
- support for the legacy `versions` config key

Backwards-incompatible behavior changes, command-line differences, config key
changes, and other functionality differences from `git-wtf` should be
documented in this README.

## Development

Run the test suite:

```sh
go test ./...
```

Build the CLI:

```sh
go build -o git-wth .
```

Check help output:

```sh
./git-wth --help
```
