# git-wth

`git-wth` is Git What The Heck: a Go rewrite of `git-wtf`.

It displays the state of a Git repository in a readable, easy-to-scan format.
It's useful for getting a summary of how a branch relates to a remote, and for
wrangling many topic branches.

git-wth can show you:
- How a branch relates to the remote repo, if it's a tracking branch.
- How a branch relates to integration branches, if it's a feature branch.
- How a branch relates to the feature branches, if it's an integration
  branch.

git-wth is best used before a git push, or between a git fetch and a git
merge. Be sure to set color.ui to auto or yes for maximum viewing pleasure.

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
example `heads/main`. Remote branches must use `remotes/<remote>/<branch>`.

`versions` is also accepted as a legacy alias for `integration-branches`.

## History and Licensing

git-wth is a Go port of `git-wtf`, originally written in Ruby by
William Morgan (2008–2009). This project continues under the original
GPL-3.0-or-later license.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

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
