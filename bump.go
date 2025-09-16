package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func bumpVersion(cfg *Config) error {
	currentVersion, err := getCurrentVersionByGitTag(cfg.VersionTagFormat)
	if err != nil {
		return err
	}

	curTag := versionToGitTag(currentVersion, cfg.VersionTagFormat)

	commitStats, err := processCommitMessages(
		curTag,
		cfg.IgnoreInvalidCommits,
		cfg.AllowedCommitKinds,
	)
	if err != nil {
		return err
	}

	if !commitStats.VersionCanBeIncremented() {
		warnNoCommitsThatCanIncrementCurrentVersion()
		return nil
	}

	// Determine new version based on the processed commits
	newVersion := currentVersion
	newVersion.Increment(commitStats.HasBreakingChange, len(commitStats.Features) != 0)
	newVersionStr := newVersion.ToString()

	fmt.Printf("Bumping next release version %s...\n", newVersionStr)

	// Generate Changelog
	if err := generateChangelog(currentVersion, newVersion, commitStats, cfg); err != nil {
		return err
	}

	// Update version files
	for _, name := range cfg.VersionFilenames {
		ext := filepath.Ext(name)
		if ext == ".json" {
			schema := JSONVersionFileSchema{Version: newVersionStr}
			if err := writeJSONFile(name, schema, 0o600); err != nil {
				return fmt.Errorf("could not write version file %s: %w", name, err)
			}
		} else {
			if err := os.WriteFile(name, []byte(newVersionStr), 0o600); err != nil {
				return fmt.Errorf("could not write version file %s: %w", name, err)
			}
		}
	}

	// Stage Changelog and version files
	gitAddArgs := []string{"add", cfg.ChangeLogFilename}
	gitAddArgs = append(gitAddArgs, cfg.VersionFilenames...)
	if _, errOut, err := getCommandOutput("git", gitAddArgs...); err != nil {
		return fmt.Errorf("error staging files to Git: %s", errOut)
	}

	// Commit release
	if _, errOut, err := getCommandOutput("git", "commit", "-m", "chore(release): "+newVersionStr); err != nil {
		return fmt.Errorf("error committing changes to Git: %s", errOut)
	}

	// Tag release
	if _, errOut, err := getCommandOutput("git", "tag", "v"+newVersionStr); err != nil {
		return fmt.Errorf("error adding Git tag for this release: %s", errOut)
	}

	return nil
}
