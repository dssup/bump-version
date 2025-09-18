# Changelog

## [0.2.0](///compare/v0.1.4...v0.2.0) (2025-09-18)


### Features

* add revert commits support and add more commit message checks 6f9b374
* **config:** add "bumpVersionCommit" config property 22aeae8
* replace regexp conventional commit parser with a full-featured parser (better errors, extensible) 0d6fb6d


### Bug Fixes

* embed utility's VERSION file to not require it alongside the program 263b5b1
* ensure the utility cd's back to Git repo root path even if started in deeper levels dcc6009

## [0.1.4](///compare/v0.1.3...v0.1.4) (2025-09-16)


### Bug Fixes

* ensure the version cannot be bumped if there's no commits that can increment it cc80e11

## [0.1.3](///compare/v0.1.2...v0.1.3) (2025-09-16)


### Bug Fixes

* ensure version numbers in version, my-version, next-version commands are printed as X.Y.Z 4b65de8

## [0.1.2](///compare/v0.1.1...v0.1.2) (2025-09-16)


### Bug Fixes

* **usage:** ensure option descriptions have proper alignment f1bbf77

## [0.1.1](///compare/v0.1.0...v0.1.1) (2025-09-16)


### Bug Fixes

* ensure the lint and lint-all commands use the regexp that considers Git hash 8000879

## [0.1.0](///compare/v0.0.1...v0.1.0) (2025-09-16)


### Features


* add the initial implementation of the utility ac773e5





