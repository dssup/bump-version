package main

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func generateChangelog(currentVersion, newVersion Version, commitStats CommitStats, cfg *Config) error {
	newChangeLog := generateNewChangelogHead(currentVersion, newVersion, commitStats)

	// Merge with the current Changelog if it exists
	if _, err := os.Stat(cfg.ChangeLogFilename); err == nil {
		newChangeLog.WriteByte('\n')

		// Read old changelog file
		changeLogFileContentBytes, err := os.ReadFile(cfg.ChangeLogFilename)
		if err != nil {
			return err
		}

		changeLogFileContent := string(changeLogFileContentBytes)
		changeLogLines := strings.Split(changeLogFileContent, "\n")

		// Remove old header if exists
		headerRe := regexp.MustCompile(`^\s*#\s*Changelog\s*$`)
		if len(changeLogLines) > 0 && headerRe.MatchString(changeLogLines[0]) {
			changeLogLines = changeLogLines[1:]
		}

		// Remove all leading empty lines
		for len(changeLogLines) > 0 && strings.TrimSpace(changeLogLines[0]) == "" {
			changeLogLines = changeLogLines[1:]
		}

		// Append all the rest lines of the old Changelog to the updated changelog
		for _, line := range changeLogLines {
			newChangeLog.WriteString(line)
			newChangeLog.WriteByte('\n')
		}
	}

	// Write updated Changelog
	if err := os.WriteFile(cfg.ChangeLogFilename, newChangeLog.Bytes(), 0o600); err != nil {
		return err
	}

	return nil
}

func generateNewChangelogHead(currentVersion, newVersion Version, commitStats CommitStats) bytes.Buffer {
	curDateStr := time.Now().Format(time.DateOnly)

	var newChangeLog bytes.Buffer

	// Write Changelog header
	fmt.Fprintf(&newChangeLog,
		"# Changelog\n\n## [%s](///compare/v%s...v%s) (%s)\n",
		newVersion.ToString(),
		currentVersion.ToString(),
		newVersion.ToString(),
		curDateStr,
	)

	// Write sorted new feature list if any
	if len(commitStats.Features) != 0 {
		// Write header
		newChangeLog.WriteString("\n\n### Features\n\n\n")

		// Stringify and sort feature list
		sortedFeatures := make([]string, len(commitStats.Features))
		for i, feature := range commitStats.Features {
			sortedFeatures[i] = commitToChangelogRecord(feature)
		}
		sortChangeLogRecords(sortedFeatures)

		// Write feature list
		for _, feature := range sortedFeatures {
			newChangeLog.WriteString(feature)
			newChangeLog.WriteByte('\n')
		}
	}

	// Write sorted fix list if any
	if len(commitStats.Fixes) != 0 {
		// Write header
		newChangeLog.WriteString("\n\n### Bug Fixes\n\n")

		// Stringify and sort fix list
		sortedFixes := make([]string, len(commitStats.Fixes))
		for i, fix := range commitStats.Fixes {
			sortedFixes[i] = commitToChangelogRecord(fix)
		}
		sortChangeLogRecords(sortedFixes)

		// Write fix list
		for _, fix := range sortedFixes {
			newChangeLog.WriteString(fix)
			newChangeLog.WriteByte('\n')
		}
	}

	return newChangeLog
}

func commitToChangelogRecord(c Commit) string {
	breakingChange := ""
	if c.Breaking {
		breakingChange = " [BREAKING CHANGE]"
	}
	if c.Scope != "" {
		return fmt.Sprintf("* **%s:**%s %s %s", c.Scope, breakingChange, c.Description, c.Hash)
	}
	return fmt.Sprintf("*%s %s %s", breakingChange, c.Description, c.Hash)
}

func sortChangeLogRecords(records []string) {
	slices.SortStableFunc(records, func(a, b string) int {
		// skip '*' characters and compare by case-folded runes
		ai, bi := 0, 0
		for ai < len(a) && bi < len(b) {
			// advance to next non-asterisk rune in a
			if a[ai] == '*' {
				ai++
				continue
			}
			// advance to next non-asterisk rune in b
			if b[bi] == '*' {
				bi++
				continue
			}

			ra, sizeA := utf8.DecodeRuneInString(a[ai:])
			rb, sizeB := utf8.DecodeRuneInString(b[bi:])

			fa := unicode.ToLower(ra)
			fb := unicode.ToLower(rb)
			if fa != fb {
				return cmp.Compare(fa, fb)
			}

			ai += sizeA
			bi += sizeB
		}

		// skip remaining '*' in either string
		for ai < len(a) && a[ai] == '*' {
			ai++
		}
		for bi < len(b) && b[bi] == '*' {
			bi++
		}

		// shorter (remaining) string is less
		return cmp.Compare(len(a)-ai, len(b)-bi)
	})
}
