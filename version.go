package main

import (
	"fmt"
	"runtime/debug"
)

var (
	Version   = "unknown"
	BuildTime = ""
)

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return Version
	}

	commit := ""
	date := ""
	modified := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.time":
			date = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}

	if BuildTime != "" {
		date = BuildTime
	}

	res := fmt.Sprintf("git-wth %s", Version)
	if commit != "" {
		short := commit[:min(len(commit), 7)]
		if modified {
			short += "-dirty"
		}
		res += fmt.Sprintf(" (%s)", short)
	}
	if date != "" {
		res += fmt.Sprintf(" built on %s", date)
	}

	return res
}
