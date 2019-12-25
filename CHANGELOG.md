# Changelog
## v0.4.0 - 2019-12-25
### Added

- Add more tests to `config_test.go`
- Test the `git.go` file
- Add Makefile targets to run tests
- Optionally log requests made to the router

### Changed

- Add a margin at the page bottom
- Rename `TreeBlob` to `Blob`

## Fixed

- Fix bugs when regular files were at the `repo-root`
- Ignore `generate.go` when building or testing
- Add a DOCTYPE
- Add a HTML lang attribute

## v0.3.0 - 2019-10-17
### Added

- Create tarballs when building
- Strip the `.git` suffix when displaying names of repositories
- If a repository cannot be found, add a `.git` suffix to its name and try to
  reopen it
- Add mobile CSS

### Changed

- Use and enforce vanity import paths

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
