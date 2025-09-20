package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type JSONVersionFileSchema struct {
	Version string `json:"version"`
}

type Version struct {
	Major, Minor, Patch int
}

func (v *Version) ToString() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) Increment(hasBreakingChange, hasFeatures, hasFixes bool) {
	switch {
	case hasBreakingChange:
		v.Major++
		v.Minor = 0
		v.Patch = 0
	case hasFeatures:
		v.Minor++
		v.Patch = 0
	case hasFixes:
		v.Patch++
	}
}

func versionToGitTag(version Version, tagFormat string) string {
	return strings.Replace(tagFormat, versionTagFormatSubst, version.ToString(), 1)
}

func parseVersion(version string) (Version, error) {
	// Get dot-separated version parts
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return Version{}, errors.New("version must have three dot-separated parts")
	}

	var major, minor, patch int
	var err error

	// Parse major version number
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version number: %w", err)
	}

	// Parse minor minor number
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version number: %w", err)
	}

	// Parse patch patch number
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version number: %w", err)
	}

	ver := Version{Major: major, Minor: minor, Patch: patch}

	return ver, nil
}

func getCurrentVersionTag() (string, error) {
	currentVersionTagGitOutput, errOut, err := getCommandOutput("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		return "", fmt.Errorf("error getting current version from Git: %s", errOut)
	}

	currentVersionTag := strings.TrimSpace(currentVersionTagGitOutput)

	return currentVersionTag, nil
}

func getCurrentVersionByGitTag(versionTagFormat string) (Version, error) {
	// Get current version tag from Git
	currentVersionTag, err := getCurrentVersionTag()
	if err != nil {
		return Version{}, err
	}

	// Get current version from current tag
	currentVersionString, err := extractVersionStringFromTag(currentVersionTag, versionTagFormat)
	if err != nil {
		return Version{}, fmt.Errorf("error extracting current version from Git tag: %w", err)
	}

	// Parse current version
	currentVersion, err := parseVersion(currentVersionString)
	if err != nil {
		return Version{}, fmt.Errorf("error parsing current version from Git tag: %w", err)
	}

	return currentVersion, nil
}

func extractVersionStringFromTag(currentVersionTag, versionTagFormat string) (string, error) {
	// Get index of the version placeholder
	subIndex := strings.Index(versionTagFormat, versionTagFormatSubst)
	if subIndex < 0 {
		panic("ExtractVersionStringFromTag assumes valid versionTagFormat")
	}

	// Get substrings before and after the placeholder
	prefix := versionTagFormat[:subIndex]
	suffix := versionTagFormat[subIndex+len(versionTagFormatSubst):]

	if !strings.HasPrefix(currentVersionTag, prefix) ||
		!strings.HasSuffix(currentVersionTag, suffix) {
		return "", fmt.Errorf("version tag %q must have format %s", currentVersionTag, versionTagFormat)
	}

	// Get rid of both the prefix and suffix substrings to
	// extract the actual version number
	return currentVersionTag[len(prefix) : len(currentVersionTag)-len(suffix)], nil
}
