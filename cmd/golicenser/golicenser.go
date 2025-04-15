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

// Package main provides a CLI for running golicenser.
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/joshuasing/golicenser"
)

var flagSet flag.FlagSet

var (
	template               string
	templateFile           string
	matcher                string
	matcherFile            string
	matcherEscape          bool
	author                 string
	authorRegexp           string
	variables              string
	yearModeStr            string
	commentStyleStr        string
	exclude                string
	maxConcurrent          int
	copyrightHeaderMatcher string
)

func init() {
	flagSet.StringVar(&template, "tmpl", "", "License header template")
	flagSet.StringVar(&templateFile, "tmpl-file", "license_header.txt",
		"License header template file")
	flagSet.StringVar(&matcher, "matcher", "",
		"License header matcher (This is template, when executed it must become valid regexp)")
	flagSet.StringVar(&matcherFile, "matcher-file", "",
		"License header matcher file)")
	flagSet.BoolVar(&matcherEscape, "matcher-escape", false,
		"Whether to regexp-escape the matcher")
	flagSet.StringVar(&author, "author", "", "Copyright author")
	flagSet.StringVar(&authorRegexp, "author-regexp", "",
		"Regexp to match copyright author (default: match author)")
	flagSet.StringVar(&variables, "var", "", "Template variables (e.g. a=Hello,b=Test)")
	flagSet.StringVar(&yearModeStr, "year-mode", golicenser.YearMode(0).String(),
		"Year formatting mode (preserve, preserve-this-year-range, preserve-modified-range, this-year, last-modified, git-range, git-modified-years)")
	flagSet.StringVar(&commentStyleStr, "comment-style", golicenser.CommentStyle(0).String(),
		"Comment style (line, block)")
	flagSet.StringVar(&exclude, "exclude", strings.Join(golicenser.DefaultExcludes, ","),
		"Paths to exclude (doublestar or r!-prefixed regexp, comma-separated)")
	flagSet.IntVar(&maxConcurrent, "max-concurrent", golicenser.DefaultMaxConcurrent,
		"Maximum concurrent processes to use when processing files")
	flagSet.StringVar(&copyrightHeaderMatcher, "copyright-header-matcher", golicenser.DefaultCopyrightHeaderMatcher,
		"Copyright header matcher regexp (used to detect existence of any copyright header)")
}

// TODO(joshuasing): There has to be a better way of doing this...

var analyzer = &analysis.Analyzer{
	Name: "golicenser",
	Doc:  "manages license headers",
	URL:  "https://github.com/joshuasing/golicenser",
	Run: func(pass *analysis.Pass) (any, error) {
		var err error
		if template == "" {
			//nolint:gosec // Reading user-defined file.
			b, err := os.ReadFile(templateFile)
			if err != nil {
				log.Fatal("read template file: ", err)
			}
			template = string(b)
		} else {
			if tm, ok := golicenser.TemplateBySPDX(template); ok {
				template = tm
			}
		}
		if matcher == "" && matcherFile != "" {
			//nolint:gosec // Reading user-defined file.
			b, err := os.ReadFile(matcherFile)
			if err != nil {
				log.Fatal("read matcher file: ", err)
			}
			matcher = string(b)
		} else {
			if tm, ok := golicenser.TemplateBySPDX(matcher); ok {
				matcher = tm
			}
		}

		// Parse variables
		vars := make(map[string]golicenser.Var)
		if variables != "" {
			for _, v := range strings.Split(variables, ",") {
				parts := strings.SplitN(v, "=", 2)
				if len(parts) != 2 {
					log.Fatal("invalid variable:", v)
				}
				vars[parts[0]] = golicenser.Var{Value: parts[1]}
			}
		}

		// Parse year mode
		var yearMode golicenser.YearMode
		if yearMode, err = golicenser.ParseYearMode(yearModeStr); err != nil {
			log.Fatal("parse year mode: ", err)
		}

		// Parse comment style
		var commentStyle golicenser.CommentStyle
		if commentStyle, err = golicenser.ParseCommentStyle(commentStyleStr); err != nil {
			log.Fatal("parse comment style: ", err)
		}

		a, err := golicenser.NewAnalyzer(golicenser.Config{
			Header: golicenser.HeaderOpts{
				Template:      template,
				Matcher:       matcher,
				MatcherEscape: matcherEscape,
				Author:        author,
				AuthorRegexp:  authorRegexp,
				Variables:     vars,
				YearMode:      yearMode,
				CommentStyle:  commentStyle,
			},
			Exclude:                strings.Split(exclude, ","),
			MaxConcurrent:          maxConcurrent,
			CopyrightHeaderMatcher: copyrightHeaderMatcher,
		})
		if err != nil {
			log.Fatal(err)
		}

		return a.Run(pass)
	},
	RunDespiteErrors: true,
}

func main() {
	analyzer.Flags = flagSet
	singlechecker.Main(analyzer)
}
