package main

import (
	"fmt"
	"os"
	"runtime/debug"
)

var (
	GitCommit string
	GitTag    string
	Version   = "v1.0.0"
)

func isAnyOf(s string, anyOf ...string) bool {
	for _, v := range anyOf {
		if s == v {
			return true
		}
	}
	return false
}

func askingForVersion() bool {
	// Check if the first argument is "--version" or "-v"
	if len(os.Args) > 1 {
		arg := os.Args[1]
		return arg == "--version" || arg == "-v" || arg == "version" || arg == "-version" || arg == "v"
	}
	return false
}

func printVersion() {
	fmt.Println("anchor-go cli")
	fmt.Println("Version:", Version)
	fmt.Printf("Tag/Branch: %q\n", GitTag)
	fmt.Printf("Commit: %q\n", GitCommit)
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("More info:\n")
		for _, setting := range info.Settings {
			if isAnyOf(setting.Key,
				"-compiler",
				"GOARCH",
				"GOOS",
				"GOAMD64",
				"vcs",
				"vcs.revision",
				"vcs.time",
				"vcs.modified",
			) {
				fmt.Printf("  %q: %q\n", setting.Key, setting.Value)
			}
		}
	}
}
