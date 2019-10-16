# Changelog
## Unrealeased
### Added

- Create tarballs when building
- Strip the `.git` suffix when displaying names of repositories
- If a repository cannot be found, add a `.git` suffix to its name and try to
  reopen it

## v0.2.0 - 2019-10-13
### Added

- Create a build Makefile
- Add the `domain` and `git-url` config options
- Show `go-import` meta tags

### Changed

- Rename the `root` config option to `repo-root`
- Refactor `handler.go`

## v0.1.0 - 2019-10-10

- Initial release
