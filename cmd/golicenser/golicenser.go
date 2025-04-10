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
	template            string
	templateFile        string
	matchTemplate       string
	matchTemplateFile   string
	matchTemplateRegexp bool
	author              string
	authorRegexp        string
	variables           string
	yearModeStr         string
	commentStyleStr     string
	exclude             string
	maxConcurrent       int
	matchHeaderRegexp   string
)

func init() {
	flag.StringVar(&template, "tmpl", "", "License header template")
	flag.StringVar(&templateFile, "tmpl-file", "license_header.txt",
		"License header template file")
	flag.StringVar(&matchTemplate, "match-tmpl", "",
		"Match license header template")
	flag.StringVar(&matchTemplateFile, "match-tmpl-file", "",
		"Match license header template file (used to detect existing license headers which may be updated)")
	flag.BoolVar(&matchTemplateRegexp, "match-tmpl-regexp", false,
		"Whether the provided match template is a regexp expression")
	flag.StringVar(&author, "author", "", "Copyright author")
	flag.StringVar(&authorRegexp, "author-regexp", "",
		"Regexp to match copyright author (default: match author)")
	flag.StringVar(&variables, "var", "", "Template variables (e.g. a=Hello,b=Test)")
	flag.StringVar(&yearModeStr, "year-mode", golicenser.YearMode(0).String(),
		"Year formatting mode (preserve, preserve-this-year-range, preserve-modified-range, this-year, last-modified, git-range, git-modified-years)")
	flag.StringVar(&commentStyleStr, "comment-style", golicenser.CommentStyle(0).String(),
		"Comment style (line, block)")
	flag.StringVar(&exclude, "exclude", strings.Join(golicenser.DefaultExcludes, ","),
		"Paths to exclude (doublestar or r!-prefixed regexp, comma-separated)")
	flag.IntVar(&maxConcurrent, "max-concurrent", golicenser.DefaultMaxConcurrent,
		"Maximum concurrent processes to use when processing files")
	flag.StringVar(&matchHeaderRegexp, "match-header-regexp", golicenser.DefaultMatchHeaderRegexp,
		"Match header regexp (used to detect any copyright headers)")
}

// TODO(joshuasing): There has to be a better way of doing this...

var analyzer = &analysis.Analyzer{
	Name:  "golicenser",
	Doc:   "manages license headers",
	URL:   "https://github.com/joshuasing/golicenser",
	Flags: flagSet,
	Run: func(pass *analysis.Pass) (any, error) {
		var err error
		if template == "" {
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
		if matchTemplate == "" && matchTemplateFile != "" {
			b, err := os.ReadFile(matchTemplateFile)
			if err != nil {
				log.Fatal("read match template file: ", err)
			}
			matchTemplate = string(b)
		} else {
			if tm, ok := golicenser.TemplateBySPDX(matchTemplate); ok {
				matchTemplate = tm
			}
		}

		// Parse variables
		vars := make(map[string]any)
		if variables != "" {
			for _, v := range strings.Split(variables, ",") {
				parts := strings.SplitN(v, "=", 2)
				if len(parts) != 2 {
					log.Fatal("invalid variable: ", v)
				}
				vars[parts[0]] = parts[1]
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
				Template:                   template,
				MatchTemplate:              matchTemplate,
				MatchTemplateEscapeDisable: matchTemplateRegexp,
				Author:                     author,
				AuthorRegexp:               authorRegexp,
				Variables:                  vars,
				YearMode:                   yearMode,
				CommentStyle:               commentStyle,
			},
			Exclude:           strings.Split(exclude, ","),
			MaxConcurrent:     maxConcurrent,
			MatchHeaderRegexp: matchHeaderRegexp,
		})
		if err != nil {
			log.Fatal(err)
		}

		return a.Run(pass)
	},
	RunDespiteErrors: true,
}

func main() {
	singlechecker.Main(analyzer)
}
