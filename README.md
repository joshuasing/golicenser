# golicenser

A fast, easy-to-use, license header linter for Go.

- **Fast.** Processes Go files in parallel to quickly lint massive codebases.
- **Easy-to-use.** Provide a license header template and copyright author, and run.
- **Customisable.**
  Configure [year formats (Git supported)](#year-modes), [license header templates](#templates), matchers
  and [comment styles](#comment-styles).
- **Standard.** Written as a [`go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis) Analyzer, making it
  possible to use with tools like golangci-lint.

---

## Comparison

|                             | golicenser                      | [go-header](https://github.com/denis-tingaikin/go-header) | [google/addlicense](https://github.com/google/addlicense) | [plantir/go-license](https://github.com/palantir/go-license) |
|:----------------------------|---------------------------------|-----------------------------------------------------------|-----------------------------------------------------------|--------------------------------------------------------------|
| Actively maintained         | ✅                               | ❌                                                         | ❌                                                         | ❓                                                            |
| Processes files in parallel | ✅                               | ❌                                                         | ✅                                                         | ❌                                                            |
| Update existing headers     | ✅                               | ✅                                                         | ✅                                                         | ❓                                                            |
| Uses `go/analysis`          | ✅                               | ❌                                                         | ❌                                                         | ❌                                                            |
| Year formats                | [Multiple options](#year-modes) | Limited¹                                                  | Current                                                   | Current                                                      |
| Git support                 | ✅                               | Limited¹                                                  | ❌                                                         | ❌                                                            |
| Exclude patterns            | ✅ (regexp & doublestar)         | ❌                                                         | ✅                                                         | ❓                                                            |
| Uses Go `text/template`     | ✅                               | ❌                                                         | ✅                                                         | ❌                                                            |
| Customise header matcher    | ✅                               | ❌                                                         | ❌                                                         | ❌                                                            |
| Supports non-Go files       | Planned                         | ❌                                                         | ✅                                                         | ❌                                                            |
| Supported by golangci-lint  | Planned                         | ✅                                                         | ❌                                                         | ❌                                                            |

*¹ Supports modified year, Git range, range, and current year*<br/>

## Benchmarks

Coming soon.

## Installation

### Binaries

Pre-built binaries are available from [GitHub Releases](https://github.com/joshuasing/golicenser/releases).

You can also use `go install` to build and install a binary from source:

```shell
go install github.com/joshuasing/golicenser/cmd/golicenser@latest
```

## Usage

```shell
golicenser -tmpl=MIT -author="Joshua Sing <joshua@joshuasing.dev>" ./...
# /golicenser/templates.go:1:1: missing license header
# /golicenser/header.go:1:1: missing license header
# /golicenser/git.go:1:1: invalid license header
# /golicenser/analysis.go:1:1: missing license header
# /golicenser/cmd/golicenser/golicenser.go:1:1: invalid license header
# /golicenser/analysis_test.go:1:1: invalid license header
# /golicenser/header_test.go:1:1: invalid license header
```

<details>
  <summary>Command help</summary>

```
golicenser: manages license headers

Usage: golicenser [-flag] [package]


Flags:
  -V    print version and exit
  -all
        no effect (deprecated)
  -author string
        Copyright author
  -author-regexp string
        Regexp to match copyright author (default: match author)
  -c int
        display offending line with this many lines of context (default -1)
  -comment-style string
        Comment style (line, block) (default "line")
  -cpuprofile string
        write CPU profile to this file
  -debug string
        debug flags, any subset of "fpstv"
  -diff
        with -fix, don't update the files, but print a unified diff
  -exclude string
        Paths to exclude (doublestar or r!-prefixed regexp, comma-separated) (default "**/testdata/**")
  -fix
        apply all suggested fixes
  -flags
        print analyzer flags in JSON
  -json
        emit JSON output
  -match-header-regexp string
        Match header regexp (used to detect any copyright headers) (default "(?i)copyright")
  -match-tmpl string
        Match license header template
  -match-tmpl-file string
        Match license header template file (used to detect existing license headers which may be updated)
  -match-tmpl-regexp
        Whether the provided match template is a regexp expression
  -max-concurrent int
        Maximum concurrent processes to use when processing files (default 32)
  -memprofile string
        write memory profile to this file
  -source
        no effect (deprecated)
  -tags string
        no effect (deprecated)
  -test
        indicates whether test files should be analyzed, too (default true)
  -tmpl string
        License header template
  -tmpl-file string
        License header template file (default "license_header.txt")
  -trace string
        write trace log to this file
  -v    no effect (deprecated)
  -var string
        Template variables (e.g. a=Hello,b=Test)
  -year-mode string
        Year formatting mode (preserve, preserve-this-year-range, preserve-modified-range, this-year, last-modified, git-range, git-modified-years) (default "preserve")
```

</details>

### Templates

golicenser uses the Go [`text/template`](https://pkg.go.dev/text/template) package to render license templates.

It is recommended to be familiar with the `text/template` syntax, however only the basics are needed. To use a variable,
the format is `{{.variableName}}`. For example, a basic template might look like this:

```text
Copyright (c) {{.year}} {{.author}}
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
```

#### Variables

Custom variables can be configured in order to deduplicate repeated strings.

#### Built-in variables

These variables are provided by golicenser.

- `year` - Copyright year(s) for file. The format is dependent on the configured [year mode](#year-modes).
- `author` - The copyright author.
- `filename` - The current filename. The root for the file is the directory where `golicenser` is run. You can use the
  `basename` function (e.g. `{{basename .filename}}`) to render only the file name if wanted.

#### Built-in functions

A few basic functions are provided by the `text/template` package. In addition to these, golicenser adds:

- `basename` ([filepath.Base](https://pkg.go.dev/path/filepath#Base)) - Returns the last element of a path.

### Exclude

There may be some cases where you want to exclude certain paths from being linted. You can provide a list of regexp
and/or doublestar patterns to be excluded.

Regexp patterns must be prefixed with `r!`, otherwise the pattern will be parsed
using [doublestar](https://github.com/bmatcuk/doublestar).

### Year modes

golicenser provides several "year modes", which are different ways of detecting and displaying the copyright year(s) for
a file. The current supported options are:

| Mode                       | New header                                                            | Updated header                                                        |
|----------------------------|-----------------------------------------------------------------------|-----------------------------------------------------------------------|
| `preserve`                 | Current year                                                          | Existing year                                                         |
| `preserve-this-year-range` | Current year                                                          | Existing year to current year (if different, e.g. `2024-2025`)        |
| `preserve-modified-range`  | Last modified year                                                    | Existing year to last modified year (if different, e.g. `2022-2025`)  |
| `this-year`                | Current year                                                          | Current year                                                          |
| `last-modified`            | Last modified year                                                    | Last modified year                                                    |
| `git-range`                | Git history creation year to last modified year (e.g. `2022-2025`)    | Git history creation year to last modified year (e.g. `2022-2025`)    |
| `git-modified-list`        | List of all modified years from Git history (e.g. `2022, 2024, 2025`) | List of all modified years from Git history (e.g. `2022, 2024, 2025`) |

*Last modified year* is detected using either Git or the local filesystem.

### Comment styles

golicenser supports configuring the comment type used for the license headers. The options are:

- `line` - C-style line comments (`// test`).
- `block` - C++-style block comments (`/* test */`)

### Matchers

golicenser allows changing the regexp matchers used to detect existing license headers, and to tell when they should be
replaced.

#### Matcher

The matcher is a regexp expression used to match the license headers generated by golicenser (or otherwise used by the
project). Matched headers are updated or replaced by golicenser.

By default, the regexp-escaped value of license header template will be used.

In some cases it may be easier to provide the plain license header here, as such it is possible to set `MatcherEscape` (
or `-matcher-escape`) to always regexp-escape the matcher value.

##### Variables

Variables are replaced with regexp expressions, in order to match outdated headers (e.g. when modified and the copyright
year changes).

- `author` - Provided author regexp (default: author value)
- `filename` - `.+`
- `year` - `(\d{4})|(\d{4})-(\d{4])|(\d{4})(?:, (\d{4}))+` - Matches a single year, year range or listed years.

#### Copyright header matcher

The copyright header matcher is used to detect **any** copyright header. By default, `(?i)copyright` is used, however 
this can be customised in order to prevent false-positives.

If a copyright header exists and does not match the [header matcher](#matcher), then a license header will not be
created. If a file starts with a comment that does not match the copyright header matcher, then a new license header
will be generated.

## Contributing

All contributions are welcome! If you have found something you think could be improved, please feel free to participate
by creating an issue or pull request!

### Building

Steps to build a golicenser binary from source.

#### Prerequisites

- [Go v1.24 or newer](https://go.dev/dl/)

#### Build

- Make: `make` (`make deps lint-deps` if missing dependencies)
- Standalone: `go build -o ./bin/golicenser ./cmd/golicenser/`

### Contact

This project is maintained by Joshua Sing. You can see a list of ways to contact me
here: https://joshuasing.dev/#contact

#### Security vulnerabilities

I take the security of my projects very seriously. As such, I strongly encourage responsible disclosure of security
vulnerabilities.

If you have discovered a security vulnerability in golicenser, please report it in accordance with the
project [Security Policy](SECURITY.md). **Never use GitHub issues to report a security vulnerability.**

### License

golicenser is distributed under the terms of the MIT License.<br/>
For more information, please refer to the [LICENSE](LICENSE) file.
