# AGENTS.md

## Project Overview

*bump-version* — A Go CLI tool for automating:
- Semantic Versioning
- Changelog generation
- Conventional Commits validation
- Git release commits and tags

## Language & Framework

- **Language:** Go 1.26
- **Build:** `cd scripts && go run . build` (outputs to `build/` directory)

## Commands

| Command | Description |
|---------|-------------|
| `bump-version` | Default command — bumps version, updates CHANGELOG, creates Git commit and tag |
| `bump-version preview-changelog` | Show changelog without writing to file |
| `bump-version add-hook` | Add Git pre-commit hook for commit message validation |
| `bump-version remove-hook` | Remove Git pre-commit hook |
| `bump-version lint` | Validate commits since last release |
| `bump-version lint-all` | Validate all commits |
| `bump-version lint-commit "message"` | Validate a single commit message |
| `bump-version my-version` | Show current version |
| `bump-version next-version` | Show next version without making changes |
| `bump-version init-config` | Create config file |
| `bump-version cancel` | Cancel recent version bump |
| `bump-version version` | Show tool version |

**Options:**
- `-config <filename>` — Use custom config file (default: `bump-version.cfg`)
- `-force` — Suppress confirmation prompts

## Version Bumping Logic

Based on Conventional Commits since last tag:
1. **MAJOR** — `BREAKING CHANGE` in commit body/footer, or `!` after type (e.g., `feat!: ...`)
2. **MINOR** — `feat:` commits present
3. **PATCH** — `fix:` commits present
4. **No change** — None of the above

## Testing

Run tests with: `go test ./...`

## Linting

Uses `.golangci.yml` — run with: `golangci-lint run`
