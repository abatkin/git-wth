package main

import (
	"fmt"
	"strings"
)

type Options struct {
	Help          bool
	Long          bool
	Short         bool
	All           bool
	AllCommits    bool
	DumpConfig    bool
	Key           bool
	ShowRelations bool
	Version       bool
	Branches      []string
}

func parseOptions(args []string) (Options, error) {
	var opts Options
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			opts.Help = true
		case "--long", "-l":
			opts.Long = true
		case "--short", "-s":
			opts.Short = true
		case "--all", "-a":
			opts.All = true
		case "--all-commits", "-A":
			opts.AllCommits = true
		case "--dump-config":
			opts.DumpConfig = true
		case "--key", "-k":
			opts.Key = true
		case "--relations", "-r":
			opts.ShowRelations = true
		case "--version":
			opts.Version = true
		default:
			if strings.HasPrefix(arg, "--") {
				return opts, fmt.Errorf("Error: unknown argument %s", arg)
			}
			opts.Branches = append(opts.Branches, arg)
		}
	}
	return opts, nil
}
