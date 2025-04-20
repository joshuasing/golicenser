// Copyright (c) 2025 Joshua Sing <joshua@joshuasing.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package golicenser

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// timeNow is time.Now. It is a function pointer in order to modify the
// time output during tests for more reliable test runs.
var timeNow = time.Now

// regexpYears matches copyright years present in license headers. It will
// match a single year, year range or listed years (comma-separated), e.g.
// "2025", "2022-2025" and "2022, 2023, 2025".
var regexpYears = regexp.MustCompile(`(?P<year>(\d{4})|(\d{4})-(\d{4])|(\d{4})(?:, (\d{4}))+)`)

// YearMode is a way of representing a copyright year(s) for a file.
type YearMode int

const (
	// YearModePreserve uses the current year in new license headers, and
	// preserves the existing year when updating license headers.
	YearModePreserve YearMode = iota

	// YearModePreserveThisYearRange uses the current year in new license headers,
	// and creates a range of the existing year to the current year when
	// updating license headers.
	YearModePreserveThisYearRange

	// YearModePreserveModifiedRange uses the last modified year in new license
	// headers, and creates a range of the existing year to the last modified
	// year when updating license headers.
	YearModePreserveModifiedRange

	// YearModeThisYear uses the current year in new license headers, and current
	// year when updating license headers.
	YearModeThisYear

	// YearModeLastModified uses the last modified year in new license headers,
	// and the last modified year when updating license headers.
	//
	// The last modified year is detected using either Git, or if modified
	// locally, the local file modification year.
	YearModeLastModified

	// YearModeGitRange uses a range of the Git history creation year to the
	// last modified year in new license headers and when updating license
	// headers.
	// Example: 2022-2025
	YearModeGitRange

	// YearModeGitModifiedList uses modification years from Git history to list
	// each year the file was modified in new license headers and when updating
	// license headers.
	// Example: 2022, 2024, 2025
	YearModeGitModifiedList
)

var yearModeStrings = map[YearMode]string{
	YearModePreserve:              "preserve",
	YearModePreserveThisYearRange: "preserve-this-year-range",
	YearModePreserveModifiedRange: "preserve-modified-range",
	YearModeThisYear:              "this-year",
	YearModeLastModified:          "last-modified",
	YearModeGitRange:              "git-range",
	YearModeGitModifiedList:       "git-modified-list",
}

// ParseYearMode parses a string representation of a year mode.
func ParseYearMode(s string) (YearMode, error) {
	switch strings.ToLower(s) {
	case yearModeStrings[YearModePreserve]:
		return YearModePreserve, nil
	case yearModeStrings[YearModePreserveThisYearRange]:
		return YearModePreserveThisYearRange, nil
	case yearModeStrings[YearModePreserveModifiedRange]:
		return YearModePreserveModifiedRange, nil
	case yearModeStrings[YearModeThisYear]:
		return YearModeThisYear, nil
	case yearModeStrings[YearModeLastModified]:
		return YearModeLastModified, nil
	case yearModeStrings[YearModeGitRange]:
		return YearModeGitRange, nil
	case yearModeStrings[YearModeGitModifiedList]:
		return YearModeGitModifiedList, nil
	default:
		return 0, fmt.Errorf("invalid year mode: %q", s)
	}
}

// String returns a string representation of the year mode.
func (ym YearMode) String() string {
	return yearModeStrings[ym]
}

// CommentStyle is a type of Go source code comment.
type CommentStyle int

const (
	// CommentStyleLine uses C-style line comments (// test).
	CommentStyleLine CommentStyle = iota

	// CommentStyleBlock uses C++-style block comments (/* test */).
	CommentStyleBlock

	// CommentStyleStarredBlock uses aligned, starred block comments, e.g.
	//  /*
	//   * Content
	//   */
	CommentStyleStarredBlock
)

// ParseCommentStyle parses a string representation of a comment style.
func ParseCommentStyle(s string) (CommentStyle, error) {
	switch strings.ToLower(s) {
	case CommentStyleLine.String():
		return CommentStyleLine, nil
	case CommentStyleBlock.String():
		return CommentStyleBlock, nil
	case CommentStyleStarredBlock.String():
		return CommentStyleStarredBlock, nil
	default:
		return 0, fmt.Errorf("invalid comment style: %q", s)
	}
}

// String returns a string representation of the comment style.
func (cs CommentStyle) String() string {
	switch cs {
	case CommentStyleLine:
		return "line"
	case CommentStyleBlock:
		return "block"
	case CommentStyleStarredBlock:
		return "starred-block"
	default:
		return ""
	}
}

// Render renders the string into a comment.
func (cs CommentStyle) Render(s string) string {
	switch cs {
	case CommentStyleLine:
		var b bytes.Buffer
		for _, l := range strings.Split(s, "\n") {
			b.WriteString("//")
			if l != "" {
				b.WriteRune(' ')
				b.WriteString(l)
			}
			b.WriteRune('\n')
		}
		return b.String()
	case CommentStyleBlock:
		return "/*\n" + s + "\n*/\n"
	case CommentStyleStarredBlock:
		var b bytes.Buffer
		b.WriteString("/*\n")
		for _, l := range strings.Split(s, "\n") {
			b.WriteString(" *")
			if l != "" {
				b.WriteRune(' ')
				b.WriteString(l)
			}
			b.WriteRune('\n')
		}
		b.WriteString(" */\n")
		return b.String()
	default:
		// Cannot render as a comment.
		return s
	}
}

// parseComment parses a comment and returns the comment content and detected
// comment style. An error will be returned if the comment cannot be parsed.
func parseComment(s string) (string, CommentStyle, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return "", 0, fmt.Errorf("invalid comment: %q", s)
	}

	switch {
	case s[0] == '/' && s[1] == '/':
		var b bytes.Buffer
		for i, l := range strings.Split(s, "\n") {
			if i != 0 {
				b.WriteRune('\n')
			}
			l = l[2:]
			if isDirective(l) {
				// Ignore build directive comments.
				return "", CommentStyleLine, nil
			}
			if len(l) > 1 && l[0] == ' ' {
				l = l[1:]
			}
			b.WriteString(l)
		}
		return b.String(), CommentStyleLine, nil
	case strings.HasPrefix(s, "/*") && strings.HasSuffix(s, "*/"):
		s = strings.TrimSpace(s[2 : len(s)-2])
		if len(s) < 2 {
			return s, CommentStyleBlock, nil
		}

		var b bytes.Buffer
		var starred bool
		lines := strings.Split(s, "\n")
		if l := lines[0]; len(l) > 0 {
			if l[0] == '*' {
				l = l[1:]
				if len(l) > 0 && l[0] == ' ' {
					l = l[1:]
				}
				starred = true
			}
			b.WriteString(l)
		}

		for _, l := range lines[1:] {
			b.WriteRune('\n')
			if strings.HasPrefix(l, " *") {
				starred = true
				if len(l) > 2 && l[2] == ' ' {
					b.WriteString(l[3:])
					continue
				}
				b.WriteString(l[2:])
				continue
			}

			// Not a starred block comment, fallback to block and just return
			// the raw comment content.
			return s, CommentStyleBlock, nil
		}

		if !starred {
			return b.String(), CommentStyleBlock, nil
		}
		return b.String(), CommentStyleStarredBlock, nil
	default:
		return "", 0, fmt.Errorf("cannot detect comment type: %q", s)
	}
}

// Header is a helper for generating and updating license headers.
type Header struct {
	tmpl    *template.Template
	matcher *regexp.Regexp

	author       string
	variables    map[string]*Var
	yearMode     YearMode
	commentStyle CommentStyle
}

var tmplFuncMap = template.FuncMap{
	"basename": filepath.Base,
}

// Var is a template variable.
type Var struct {
	// Value is the variable value.
	Value string

	// Regexp is a regexp used to match the variable value.
	// If empty, the regexp-escaped value of Value will be used.
	Regexp string
}

// HeaderOpts are the options for creating a license header.
type HeaderOpts struct {
	Template      string
	Matcher       string
	MatcherEscape bool
	Author        string
	AuthorRegexp  string
	Variables     map[string]*Var
	YearMode      YearMode
	CommentStyle  CommentStyle
}

// NewHeader creates a new header with the given options.
func NewHeader(opts HeaderOpts) (*Header, error) {
	if opts.Author == "" {
		return nil, fmt.Errorf("invalid author: %q", opts.Author)
	}
	if opts.Template == "" {
		return nil, fmt.Errorf("invalid template: %q", opts.Template)
	}

	// Parse template.
	t, err := template.New("").Funcs(tmplFuncMap).
		Option("missingkey=error").Parse(opts.Template)
	if err != nil {
		return nil, fmt.Errorf("new template: %w", err)
	}

	// Test executing the template.
	m := map[string]any{
		"author":   opts.Author,
		"filename": "test",
		"year":     "2025",
	}
	addVariables(m, opts.Variables)
	if err = t.Execute(io.Discard, m); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	// Test compiling variable regexps.
	for name, v := range opts.Variables {
		switch v.Regexp {
		case "":
			v.Regexp = regexp.QuoteMeta(v.Value)
		default:
			if _, err = regexp.Compile(v.Regexp); err != nil {
				return nil, fmt.Errorf("compile %q regexp: %w", name, err)
			}
		}
	}

	// Create author regexp
	authorRegexpStr := opts.AuthorRegexp
	if authorRegexpStr == "" {
		authorRegexpStr = regexp.QuoteMeta(opts.Author)
	}
	var authorRegexp *regexp.Regexp
	if authorRegexp, err = regexp.Compile(authorRegexpStr); err != nil {
		return nil, fmt.Errorf("compile author regexp: %w", err)
	}

	var matcher *regexp.Regexp
	if opts.Matcher != "" {
		mt, err := template.New("").Funcs(tmplFuncMap).
			Option("missingkey=error").Parse(opts.Matcher)
		if err != nil {
			return nil, fmt.Errorf("new matcher template: %w", err)
		}
		matcher, err = headerMatcher(mt, opts.MatcherEscape, authorRegexp, opts.Variables)
		if err != nil {
			return nil, fmt.Errorf("create header matcher: %w", err)
		}
	} else {
		// If a matcher wasn't provided, create a matcher using the header
		// template (regexp-escaped).
		matcher, err = headerMatcher(t, true, authorRegexp, opts.Variables)
		if err != nil {
			return nil, fmt.Errorf("create header matcher: %w", err)
		}
	}

	return &Header{
		tmpl:         t,
		matcher:      matcher,
		author:       opts.Author,
		variables:    opts.Variables,
		yearMode:     opts.YearMode,
		commentStyle: opts.CommentStyle,
	}, nil
}

// Create creates a new license header for the file.
func (h *Header) Create(filename string) (string, error) {
	header, err := h.render(filename, timeNow().Format("2006"))
	if err != nil {
		return "", fmt.Errorf("render header: %w", err)
	}
	return h.commentStyle.Render(header), nil
}

// Update updates an existing license header if it matches the
func (h *Header) Update(filename, header string) (string, bool, error) {
	header, cs, err := parseComment(header)
	if err != nil {
		return "", false, fmt.Errorf("parse header comment: %w", err)
	}
	match := h.matcher.FindStringSubmatch(header)
	if match == nil {
		return header, false, nil
	}

	var year string
	switch h.yearMode {
	case YearModePreserve:
		if i := h.matcher.SubexpIndex("year"); i != -1 {
			year = match[i]
		}
	case YearModePreserveThisYearRange:
		if i := h.matcher.SubexpIndex("year"); i != -1 {
			year = match[i]
			if parts := strings.SplitN(year, "-", 2); len(parts) > 1 {
				year = parts[0]
			}
			if currentYear := timeNow().Format("2006"); year != currentYear {
				year += "-" + currentYear
			}
		}
	case YearModePreserveModifiedRange:
		if i := h.matcher.SubexpIndex("year"); i != -1 {
			year = match[i]
			if modTime, err := lastModTime(filename); err == nil {
				if parts := strings.SplitN(year, "-", 2); len(parts) > 1 {
					year = parts[0]
				}
				if modifiedYear := modTime.Format("2006"); year != modifiedYear {
					year += "-" + modifiedYear
				}
			}
		}
	case YearModeThisYear:
		// Handled below switch.
	case YearModeLastModified:
		if modTime, err := lastModTime(filename); err == nil {
			year = modTime.Format("2006")
		}
	case YearModeGitRange:
		if created, modified, err := gitModRange(filename); err == nil {
			if created.Year() == modified.Year() {
				year = created.Format("2006")
				break
			}
			year = created.Format("2006") + "-" + modified.Format("2006")
		}
	case YearModeGitModifiedList:
		if modTimes, err := gitModTimes(filename); err == nil && len(modTimes) > 0 {
			year = modTimes[0].Format("2006")
			for i, modTime := range modTimes[1:] {
				if modTimes[i].Year() == modTime.Year() {
					continue
				}
				year = year + ", " + modTime.Format("2006")
			}
		}
	}
	if year == "" {
		year = timeNow().Format("2006")
	}

	newHeader, err := h.render(filename, year)
	if err != nil {
		return "", false, fmt.Errorf("render header: %w", err)
	}
	modified := newHeader != header || h.commentStyle != cs
	return h.commentStyle.Render(newHeader), modified, nil
}

func (h *Header) render(filename, year string) (string, error) {
	// Built-in variables.
	m := map[string]any{
		"author":   h.author,
		"filename": filename,
		"year":     year,
	}
	addVariables(m, h.variables)

	var b bytes.Buffer
	if err := h.tmpl.Execute(&b, m); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return b.String(), nil
}

func headerMatcher(tmpl *template.Template, escapeTmpl bool, authorRegexp *regexp.Regexp, variables map[string]*Var) (*regexp.Regexp, error) {
	m := map[string]string{
		"author":   "__VAR_author__",
		"filename": "__VAR_filename__",
		"year":     "__VAR_year__",
	}
	regexps := map[string]string{
		"author":   authorRegexp.String(),
		"filename": "(?P<filename>.+)",
		"year":     regexpYears.String(),
	}
	for k, v := range variables {
		m[k] = "__VAR_" + k + "__"
		regexps[k] = v.Regexp
	}

	// Execute matcher template.
	var b bytes.Buffer
	if err := tmpl.Execute(&b, m); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	headerExpr := b.String()

	// Optionally escape the rendered template.
	if escapeTmpl {
		headerExpr = regexp.QuoteMeta(b.String())
	}

	// Replace variable placeholders with regexp expressions.
	replacements := make([]string, 0, len(m)*2)
	for k, v := range m {
		replacements = append(replacements, v, regexps[k])
	}
	headerExpr = strings.NewReplacer(replacements...).Replace(headerExpr)

	// Compile header matcher regexp.
	return regexp.Compile(headerExpr)
}

func addVariables(m map[string]any, vars map[string]*Var) {
	for k, v := range vars {
		m[k] = v.Value
	}
}
