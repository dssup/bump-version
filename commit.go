package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
)

type Commit struct {
	Hash        string
	Kind        string
	Scope       string
	Description string
	Body        string
	Footer      string
	Breaking    bool
}

type CommitStats struct {
	Features          []Commit
	Fixes             []Commit
	HasBreakingChange bool
}

var conventionalCommitRegex = regexp.MustCompile(
	`(?m)^(?P<hash>[0-9a-f]{7,40})\s+(?P<kind>\w+)(\((?P<scope>[\w/\-\.]+)\))?(?P<breaking>!)?: (?P<description>.+)(\n(?P<body>.+))?(\n(?P<footer>.+))?$`,
)

var conventionalCommitMessageOnlyRegex = regexp.MustCompile(
	`(?m)^(?P<kind>\w+)(\((?P<scope>[\w/\-\.]+)\))?(?P<breaking>!)?: (?P<description>.+)(\n(?P<body>.+))?(\n(?P<footer>.+))?$`,
)

func processCommitMessages(sinceTag string, ignoreInvalidCommits bool, allowedKinds []string) (CommitStats, error) {
	commitStats := CommitStats{
		Features: make([]Commit, 0, 4096),
		Fixes:    make([]Commit, 0, 4096),
	}

	gitCommits, err := getGitCommitLines(sinceTag)
	if err != nil {
		return CommitStats{}, err
	}

	hasInvalidCommits := false

	// Process all commits in a single pass
	for _, message := range gitCommits {
		// Parse and validate commit
		commit, err := parseConventionalCommit(message, true, allowedKinds)
		if err != nil {
			hasInvalidCommits = true
			if ignoreInvalidCommits {
				fmt.Fprintf(os.Stderr,
					"%s: WARNING: ignored invalid commit message: %s. Reason: %v\n",
					programName, message, err,
				)
			} else {
				fmt.Fprintf(os.Stderr,
					"%s: ERROR: invalid commit message: %s. Reason: %v\n",
					programName, message, err,
				)
			}
			continue
		}

		// Determine if there's a breaking change
		if commit.Breaking {
			commitStats.HasBreakingChange = true
		}

		// Categorize commit by its kind
		switch commit.Kind {
		case "feat":
			commitStats.Features = append(commitStats.Features, commit)
		case "fix":
			commitStats.Fixes = append(commitStats.Fixes, commit)
		}
	}

	if hasInvalidCommits && !ignoreInvalidCommits {
		os.Exit(ExitFailure)
	}

	return commitStats, nil
}

func getGitCommitLines(sinceTag string) ([]string, error) {
	var args []string
	if sinceTag != "" {
		args = []string{"log", sinceTag + "..HEAD", "--oneline"}
	} else {
		args = []string{"log", "--oneline"}
	}

	// Get all commits since the provided tag
	gitCommitsOut, errOut, err := getCommandOutput("git", args...)
	if err != nil {
		return nil, fmt.Errorf("error getting commits from Git: %s", errOut)
	}

	rawGitCommits := strings.Split(strings.TrimSpace(gitCommitsOut), "\n")

	commits := make([]string, 0, len(rawGitCommits))

	// Trim all commits and skip empty lines
	for _, line := range rawGitCommits {
		message := strings.TrimSpace(line)
		if message != "" {
			commits = append(commits, message)
		}
	}

	return commits, nil
}

func parseConventionalCommit(message string, hasHash bool, allowedKinds []string) (Commit, error) {
	message = strings.TrimSpace(message)

	regex := conventionalCommitMessageOnlyRegex
	if hasHash {
		regex = conventionalCommitRegex
	}

	// Parse and validate commit
	match := regex.FindStringSubmatch(message)
	if match == nil {
		return Commit{}, errors.New("commit message does not match Conventional Message pattern")
	}

	names := regex.SubexpNames()

	// helper to get named regexp group value, or empty string if not present
	reGroup := func(match []string, names []string, name string) string {
		if i := slices.Index(names, name); i >= 0 {
			return match[i]
		}
		return ""
	}

	var commit Commit

	// Extract features from the commit string

	commit.Kind = reGroup(match, names, "kind")
	if !slices.Contains(allowedKinds, commit.Kind) {
		return Commit{}, fmt.Errorf("commit message kind %s not allowed", commit.Kind)
	}

	commit.Hash = reGroup(match, names, "hash")
	commit.Scope = reGroup(match, names, "scope")
	commit.Description = reGroup(match, names, "description")
	commit.Body = strings.TrimSpace(reGroup(match, names, "body"))
	commit.Footer = strings.TrimSpace(reGroup(match, names, "footer"))
	commit.Breaking = reGroup(match, names, "breaking") != "" || strings.Contains(commit.Footer, "BREAKING CHANGE")

	return commit, nil
}
