package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const programName = "bump-version"

const preCommiHookPath = ".git/hooks/pre-commit"

// Program exit status codes.
const (
	ExitSuccess = 0 // may be warnings but no error occurred
	ExitFailure = 1 // an error occurred
	ExitUsage   = 2 // invalid usage of the program
)

func main() {
	configFilename := "" // not determined yet

	forceMode := false

	args := os.Args[1:] // get rid of executable name from args

	// Process all options
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "-config":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "%s: option -config needs filename argument\n", programName)
				os.Exit(ExitUsage)
			}
			configFilename = args[1]
			args = args[2:]
		case "-force":
			forceMode = true
			args = args[1:]
		default:
			fmt.Fprintf(os.Stderr, "%s: unknown option %s\n\n", programName, args[0])
			printHelpAndExit()
		}
	}

	// If no command specified, default to command `bump`
	if len(args) < 1 {
		cfg := fatalOnErr2(loadConfig(configFilename))
		fatalOnErr(bumpVersion(&cfg))
		return
	}

	command := args[0]

	// Dispatch commands
	switch command {
	case "bump":
		cfg := fatalOnErr2(loadConfig(configFilename))
		fatalOnErr(bumpVersion(&cfg))

	case "preview-changelog":
		cfg := fatalOnErr2(loadConfig(configFilename))

		currentVersion := fatalOnErr2(getCurrentVersionByGitTag(cfg.VersionTagFormat))

		curTag := versionToGitTag(currentVersion, cfg.VersionTagFormat)

		commitStats := fatalOnErr2(processCommitMessages(curTag, cfg.IgnoreInvalidCommits, cfg.AllowedCommitKinds))

		newVersion := currentVersion
		newVersion.Increment(commitStats.HasBreakingChange, len(commitStats.Features) != 0)

		newChangeLog := generateNewChangelogHead(currentVersion, newVersion, commitStats)
		fmt.Println(newChangeLog.String())

	case "add-hook":
		// Grant the permission to execute
		execPerm := os.FileMode(0o755)

		// Create hooks directory if missing
		if err := os.MkdirAll(filepath.Dir(preCommiHookPath), execPerm); err != nil {
			fatalf("could not create the directory for the hook: %v", err)
		}

		hookScript := fmt.Sprintf(`#!/bin/sh
%s lint-commit "$COMMIT_MESSAGE"
`, programName)

		// If the hook file already exists, ask the user to confirm overwriting it
		if !forceMode && fileExists(preCommiHookPath) {
			question := "The hook file already exists. Do you want to overwrite it?"
			if !askYesNo(os.Stdin, os.Stdout, question, false) {
				fmt.Println("Canceled.")
				return
			}
		}

		// Write the hook file
		if err := os.WriteFile(preCommiHookPath, []byte(hookScript), execPerm); err != nil {
			fatalf("could not write the hook file: %v", err)
		}

		// Ensure executable permission
		if err := os.Chmod(preCommiHookPath, execPerm); err != nil {
			fatalf("could not ensure the hook is executable: %v", err)
		}

		fmt.Println("Git Hook added")

	case "remove-hook":
		if err := os.RemoveAll(preCommiHookPath); err != nil {
			fatalf("could not remove the hook file: %v", err)
		}

		fmt.Println("Git Hook removed")

	case "lint-commit":
		cfg := fatalOnErr2(loadConfig(configFilename))

		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "%s: lint-commit needs a commit message as its argument\n", programName)
			os.Exit(ExitUsage)
		}

		if _, err := parseConventionalCommit(args[1], false, cfg.AllowedCommitKinds); err != nil {
			fatalf("invalid commit message: %v", err)
			os.Exit(ExitFailure)
		}

	case "lint", "lint-all":
		cfg := fatalOnErr2(loadConfig(configFilename))

		var gitCommits []string

		if command == "lint-all" {
			gitCommits = fatalOnErr2(getGitCommitLines(""))
		} else {
			currentVersion := fatalOnErr2(getCurrentVersionByGitTag(cfg.VersionTagFormat))
			curTag := versionToGitTag(currentVersion, cfg.VersionTagFormat)
			gitCommits = fatalOnErr2(getGitCommitLines(curTag))
		}

		hasInvalidCommits := false

		// Iterate over the chosen commits and validate them
		for _, message := range gitCommits {
			_, err := parseConventionalCommit(message, true, cfg.AllowedCommitKinds)
			if err != nil {
				hasInvalidCommits = true
				fmt.Fprintf(os.Stderr,
					"%s: invalid commit message: %s. Reason: %v\n",
					programName, message, err,
				)
			}
		}

		if hasInvalidCommits {
			os.Exit(ExitFailure)
		}

	case "my-version":
		cfg := fatalOnErr2(loadConfig(configFilename))
		currentVersion := fatalOnErr2(getCurrentVersionByGitTag(cfg.VersionTagFormat))
		fmt.Println(currentVersion.ToString())

	case "next-version":
		cfg := fatalOnErr2(loadConfig(configFilename))

		currentVersion := fatalOnErr2(getCurrentVersionByGitTag(cfg.VersionTagFormat))

		curTag := versionToGitTag(currentVersion, cfg.VersionTagFormat)

		commitStats := fatalOnErr2(processCommitMessages(curTag, cfg.IgnoreInvalidCommits, cfg.AllowedCommitKinds))

		newVersion := currentVersion
		newVersion.Increment(commitStats.HasBreakingChange, len(commitStats.Features) != 0)

		fmt.Println(newVersion.ToString())

	case "init-config":
		defaultConfig := getDefaultConfig()

		// If the config file already exists, ask the user to confirm overwriting it
		if !forceMode && fileExists(defaultConfigFilename) {
			question := fmt.Sprintf(
				"Config file %s already exists. Do you want to overwrite it?",
				defaultConfigFilename,
			)
			if !askYesNo(os.Stdin, os.Stdout, question, false) {
				fmt.Println("Canceled.")
				return
			}
		}

		// Write the default config file
		fatalOnErr(writeJSONFile(defaultConfigFilename, defaultConfig, 0o644))

	case "version":
		filename := "VERSION"

		// Read version file
		contentBytes, err := os.ReadFile(filename)
		if err != nil {
			fatalf("could not read version file %s: %v", filename, err)
		}

		// Parse version
		version, err := parseVersion(strings.TrimSpace(string(contentBytes)))
		if err != nil {
			fatalf("could not parse version file %s: %v", filename, err)
		}

		fmt.Println(version.ToString())

	case "help":
		printHelpAndExit()

	default:
		fmt.Fprintf(os.Stderr, "%s: unknown command %s\n\n", programName, os.Args[1])
		printHelpAndExit()
	}
}

func printHelpAndExit() {
	fmt.Fprintf(os.Stderr, `Usage: %s [command] [command options...]

Commands:

  preview-changelog    show this release changelog (not writing to Changelog file
  add-hook             add Git hook to validate commit messages
  remove-hook          remove Git hook validating commit messages
  lint                 list commits violating Conventional Commits since last bump
  lint-all             list commits violating Conventional Commits since Git Init
  lint-commit message  lint the provided commit message
  my-version           show the current version of your program and exit
  next-version         show the next version of your program and exit
  init-config          add a config file bump-version.json to the project root
  version              show this program version and exit
  help                 show this help and exit

Common options:

  -config filename  use another config filename instead of default
                    (default config name: bump-version.json)
  -force            suppress the prompts like "Are you sure you want to
                    overwrite this file?"
If no command is specified, the default command to run is bump.
`, programName)

	os.Exit(ExitUsage)
}
