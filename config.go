package main

import (
	"fmt"
	"strings"
)

const (
	defaultConfigFilename = "bump-version.json"
	versionTagFormatSubst = "{version}" // substitute for an actual version
)

type Config struct {
	Version              string   `json:"version"`
	VersionFilenames     []string `json:"versionFilenames"`
	ChangeLogFilename    string   `json:"changeLogFilename"`
	IgnoreInvalidCommits bool     `json:"ignoreInvalidCommits"`
	VersionTagFormat     string   `json:"versionTagFormat"`
	AllowedCommitKinds   []string `json:"allowedCommitKinds"`
	BumpVersionCommit    string   `json:"bumpVersionCommit"`
	ShouldPushToOrigin   bool     `json:"shouldPushToOrigin"`
}

// getDefaultConfig returns the default config as read-only.
func getDefaultConfig() Config {
	return Config{
		Version:              "1.0.0",
		VersionFilenames:     []string{"VERSION"},
		ChangeLogFilename:    "CHANGELOG.md",
		IgnoreInvalidCommits: false,
		VersionTagFormat:     "v{version}",
		AllowedCommitKinds: []string{
			"BREAKING CHANGE",
			"build",
			"chore",
			"ci",
			"docs",
			"feat",
			"fix",
			"perf",
			"refactor",
			"revert",
			"style",
			"test",
		},
		BumpVersionCommit:  "chore(release): {version}",
		ShouldPushToOrigin: false,
	}
}

func loadConfig(name string) (Config, error) {
	defaultConfig := getDefaultConfig()

	// If the config file name is empty, read the default config file,
	// but even if that file does not exist, use the default config
	// object.
	if name == "" {
		name = defaultConfigFilename
		if !fileExists(name) {
			return defaultConfig, nil
		}
	}

	var cfg Config

	// Read config from file
	if err := readJSONFile(name, &cfg); err != nil {
		return Config{}, err
	}

	// Validate version tag format
	if strings.Count(cfg.VersionTagFormat, versionTagFormatSubst) != 1 {
		err := fmt.Errorf(
			"version tag format %q expected to have one occurrence of %s",
			cfg.VersionTagFormat,
			versionTagFormatSubst,
		)
		return Config{}, err
	}

	// Replace zero values with default values
	if cfg.Version == "" {
		cfg.Version = defaultConfig.Version
	}
	if cfg.VersionFilenames == nil {
		cfg.VersionFilenames = defaultConfig.VersionFilenames
	}
	if cfg.ChangeLogFilename == "" {
		cfg.ChangeLogFilename = defaultConfig.ChangeLogFilename
	}
	if !cfg.IgnoreInvalidCommits {
		cfg.IgnoreInvalidCommits = defaultConfig.IgnoreInvalidCommits
	}
	if cfg.VersionTagFormat == "" {
		cfg.VersionTagFormat = defaultConfig.VersionTagFormat
	}
	if cfg.AllowedCommitKinds == nil {
		cfg.AllowedCommitKinds = defaultConfig.AllowedCommitKinds
	}
	if cfg.BumpVersionCommit == "" {
		cfg.BumpVersionCommit = defaultConfig.BumpVersionCommit
	}

	return cfg, nil
}
