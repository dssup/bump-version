package main

import (
	"errors"
	"fmt"
	"os"
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
	RefersTo    []string // commit hashes this commit refers to
}

type CommitStats struct {
	Features          []Commit
	Fixes             []Commit
	HasBreakingChange bool
	AllCommits        []Commit
}

func (s CommitStats) HasNewFeatures() bool {
	return len(s.Features) != 0
}

func (s CommitStats) HasNewFixes() bool {
	return len(s.Fixes) != 0
}

func (s CommitStats) VersionCanBeIncremented() bool {
	return s.HasNewFeatures() || s.HasNewFixes()
}

func processCommitMessages(sinceTag string, ignoreInvalidCommits bool, allowedKinds []string) (CommitStats, error) {
	commitStats := CommitStats{
		Features:   make([]Commit, 0, 4096),
		Fixes:      make([]Commit, 0, 4096),
		AllCommits: make([]Commit, 0, 4096),
	}

	gitCommits, err := getGitCommitLines(sinceTag)
	if err != nil {
		return CommitStats{}, err
	}

	hasInvalidCommits := false

	// Process all commits in a single pass
	for _, message := range gitCommits {
		// Parse and validate commit
		commit, err := ParseConventionalCommit(message, true, allowedKinds)
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

		commitStats.AllCommits = append(commitStats.AllCommits, commit)
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

func ParseConventionalCommit(message string, hasHash bool, allowedKinds []string) (Commit, error) {
	if strings.TrimSpace(message) == "" {
		return Commit{}, errors.New("empty commit message")
	}

	// Split into lines preserving order
	lines := strings.Split(message, "\n")
	for lineIndex, line := range lines {
		lineNo := lineIndex + 1
		if HasLeadingWhitespace(line) || HasTrailingWhitespace(line) {
			return Commit{}, fmt.Errorf("commit message line number %d (`%s`) should not have have leading or trailing whitespace", lineNo, line)
		}
	}

	commit := Commit{}

	header := lines[0]
	if HasLeadingWhitespace(header) {
		return Commit{}, errors.New("commit message header should not have leading whitespace")
	}
	if HasTrailingWhitespace(header) {
		return Commit{}, errors.New("commit message header should not have trailing whitespace")
	}
	if header == "" {
		return Commit{}, errors.New("empty commit message header")
	}

	// Parse optional leading hash
	if hasHash {
		parts := strings.Fields(header)

		hash, err := parseCommitHash(parts[0])
		if err != nil {
			return Commit{}, err
		}
		commit.Hash = hash

		if len(header) <= len(commit.Hash) || header[len(commit.Hash)] != ' ' {
			return Commit{}, errors.New("hash must be followed by a single space")
		}

		// reconstruct the header without hash and the single space after it
		header = header[len(commit.Hash)+1:]
	}

	// Check for leading spaces again
	if HasLeadingWhitespace(header) {
		return Commit{}, errors.New("commit message header should not have leading whitespace")
	}

	// Find the first ":" that separates header from description
	colonIndex := strings.Index(header, ":")
	if colonIndex == -1 {
		return Commit{}, errors.New("commit message header must have colon to separate commit kind and description")
	}

	// Parse and validate kind (stage 1)
	headerBeforeDescription := header[:colonIndex]
	if HasLeadingWhitespace(headerBeforeDescription) {
		return Commit{}, errors.New("commit message kind should not have leading whitespace")
	}
	if HasTrailingWhitespace(headerBeforeDescription) {
		return Commit{}, errors.New("commit message kind should not have trailing whitespace before colon")
	}
	breaking := false
	if strings.HasSuffix(headerBeforeDescription, "!") {
		breaking = true
		// Get rid of !
		headerBeforeDescription = strings.TrimSuffix(headerBeforeDescription, "!")
		// Check for trailing spaces
		if HasTrailingWhitespace(headerBeforeDescription) {
			return Commit{}, errors.New("commit message kind should not have whitespace before `!'")
		}
	}

	// Extract kind and optional scope (stage 2)
	var kind string
	scope := ""
	if start := strings.Index(headerBeforeDescription, "("); start != -1 {
		// attempt to find matching ')'
		end := strings.LastIndex(headerBeforeDescription, ")")
		if end == -1 || end <= start {
			return Commit{}, errors.New("commit message kind must have matching parentheses")
		}

		kind = headerBeforeDescription[:start]
		if HasTrailingWhitespace(kind) {
			return Commit{}, errors.New("commit message kind should not have spaces before `('")
		}

		scope = headerBeforeDescription[start+1 : end]
		if HasLeadingWhitespace(scope) || HasTrailingWhitespace(scope) {
			return Commit{}, errors.New("commit message scope should not be enclosed with spaces")
		}

		if scope == "" {
			return Commit{}, errors.New("commit message scope cannot be empty within parentheses")
		}
	} else {
		kind = headerBeforeDescription
	}
	if kind == "" {
		return Commit{}, errors.New("commit message header should have non-empty kind")
	}
	commit.Kind = kind
	commit.Scope = scope
	commit.Breaking = breaking

	// Parse and validate description
	description := header[colonIndex+1:]
	if !strings.HasPrefix(description, " ") {
		return Commit{}, errors.New("commit message header must have space after colon")
	}
	description = description[1:] // skip one space
	if HasLeadingWhitespace(description) {
		return Commit{}, errors.New("commit message description should have only one whitespace after colon")
	}
	if HasTrailingWhitespace(headerBeforeDescription) {
		return Commit{}, errors.New("commit message description should not have trailing whitespace")
	}
	if description == "" {
		return Commit{}, errors.New("commit message header should have non-empty description")
	}
	if strings.HasSuffix(description, ".") {
		return Commit{}, errors.New("commit message description should not end with period")
	}
	commit.Description = description

	// Validate kind against allowedKinds
	if !slices.Contains(allowedKinds, commit.Kind) {
		return Commit{}, fmt.Errorf("commit message kind %s not allowed", commit.Kind)
	}

	// The remainder of the message (lines[1:]) is body and footer separated by a blank line.
	bodyLines := []string{}
	footerLines := []string{}
	if len(lines) > 1 {
		// Skip possible empty line immediately after header
		i := 1

		// consume leading blank lines between header and body
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}

		// Collect lines until a blank line that separates body and footer
		for ; i < len(lines); i++ {
			// find first blank line that will separate body from footer
			if strings.TrimSpace(lines[i]) == "" {
				// everything after the blank line (skipping blank lines) is footer
				j := i + 1
				for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
					j++
				}
				if j < len(lines) {
					// footer exists
					bodyLines = append(bodyLines, lines[1:i]...)
					footerLines = append(footerLines, lines[j:]...)
				} else {
					// trailing blank lines only; body is lines[1:i]
					bodyLines = append(bodyLines, lines[1:i]...)
				}
				break
			}
		}

		if len(bodyLines) == 0 && len(footerLines) == 0 {
			// no blank separator found; everything after header is body
			bodyLines = append(bodyLines, lines[1:]...)
		}
	}

	commit.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	commit.Footer = strings.TrimSpace(strings.Join(footerLines, "\n"))

	// If footer mentions BREAKING CHANGE, mark breaking
	if !commit.Breaking && strings.Contains(commit.Footer, "BREAKING CHANGE") {
		commit.Breaking = true
	}

	// Parse references from body
	for _, line := range bodyLines {
		if err := tryParseReference(line, &commit); err != nil {
			return Commit{}, err
		}
	}

	// Parse references from footer
	for _, line := range footerLines {
		if err := tryParseReference(line, &commit); err != nil {
			return Commit{}, err
		}
	}

	// Make sure revert commits have at least one reference to other commits
	if commit.Kind == "revert" && len(commit.RefersTo) == 0 {
		return Commit{}, errors.New("revert commits must have hashes of the commits they revert; add `Refs: ` attribute to commit footer")
	}

	// Check duplicate references
	if len(commit.RefersTo) != 0 {
		alreadySeenRefs := make(map[string]bool, len(commit.RefersTo))
		for _, ref := range commit.RefersTo {
			if alreadySeenRefs[ref] {
				return Commit{}, fmt.Errorf("commit message has duplicated ref hash %s", ref)
			}
			alreadySeenRefs[ref] = true
		}
	}

	return commit, nil
}

func warnNoCommitsThatCanIncrementCurrentVersion() {
	fmt.Fprintln(os.Stderr, programName+": no commits that can increment current version")
}

func parseCommitHash(hash string) (string, error) {
	if len(hash) < 7 || len(hash) > 40 {
		return "", errors.New("invalid commit hash length")
	}

	for _, digit := range hash {
		if !IsLowerHexDigit(digit) {
			return "", errors.New("commit message expected to have lowercase hexadecimal digits only")
		}
	}

	return hash, nil
}

func tryParseReference(line string, commit *Commit) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		return nil
	}

	if HasLeadingWhitespace(line) || HasTrailingWhitespace(line) {
		return fmt.Errorf("footer line attribute should not be enclosed with spaces: %s", line)
	}

	// Parse attribute name
	attrName := line[:colonIndex]
	if HasTrailingWhitespace(attrName) {
		return fmt.Errorf("footer line attribute should not have spaces before colon: %s", line)
	}

	// Parse attribute value
	attrValue := line[colonIndex+1:]
	if !strings.HasPrefix(attrValue, " ") {
		return fmt.Errorf("footer line attribute value should not have only one space after colon: %s", line)
	}
	attrValue = attrValue[1:] // skip leading space
	if HasLeadingWhitespace(attrValue) {
		return fmt.Errorf("footer line attribute value should not have spaces after colon with spaces: %s", line)
	}

	// Dispatch attribyte based on its name
	refsAttr := "Refs"
	if strings.EqualFold(attrName, refsAttr) {
		if attrName != refsAttr {
			return fmt.Errorf("attribute %s should have consistent case like this: %s", attrName, refsAttr)
		}

		hashListParts := strings.Split(attrValue, ",")
		hashes := make([]string, 0, len(hashListParts))
		for hashIndex, hashStr := range hashListParts {
			if hashIndex != 0 {
				if !strings.HasPrefix(hashStr, " ") {
					return fmt.Errorf("hash list in attribute %s should be delimited by comma and a single space: %s", attrName, hashStr)
				}
				hashStr = hashStr[1:] // skip space
			}
			if HasLeadingWhitespace(hashStr) || HasTrailingWhitespace(hashStr) {
				return fmt.Errorf("hash list in attribute %s should be delimited by comma and a single space: %s", attrName, hashStr)
			}

			hash, err := parseCommitHash(hashStr)
			if err != nil {
				return fmt.Errorf("invalid hash %s: %w", hashStr, err)
			}

			hashes = append(hashes, hash)
		}

		commit.RefersTo = append(commit.RefersTo, hashes...)
	}

	return nil
}
