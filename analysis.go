package golicenser

import (
	"fmt"
	"go/ast"
	"regexp"
	"runtime"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/analysis"
)

const (
	analyzerName = "golicenser"

	DefaultMatchHeaderRegexp = "(?i)copyright"
)

var (
	DefaultMaxConcurrent = runtime.GOMAXPROCS(0) * 2
	DefaultExcludes      = []string{
		"**/testdata/**", // Exclude testdata directories
	}
)

type Config struct {
	Header HeaderOpts

	Exclude           []string
	MaxConcurrent     int
	MatchHeaderRegexp string
}

func NewAnalyzer(cfg Config) (*analysis.Analyzer, error) {
	a, err := newAnalyzer(cfg)
	if err != nil {
		return nil, err
	}

	return &analysis.Analyzer{
		Name:             analyzerName,
		Doc:              "manages license headers",
		URL:              "https://github.com/joshuasing/golicenser",
		Run:              a.run,
		RunDespiteErrors: true,
	}, nil
}

type ExcludeMatcherFunc func(filename string) bool

type analyzer struct {
	cfg           Config
	excludes      []ExcludeMatcherFunc
	headerMatcher *regexp.Regexp

	header *Header
}

func newAnalyzer(cfg Config) (*analyzer, error) {
	if cfg.MaxConcurrent < 1 {
		cfg.MaxConcurrent = DefaultMaxConcurrent
	}
	if cfg.MatchHeaderRegexp == "" {
		cfg.MatchHeaderRegexp = DefaultMatchHeaderRegexp
	}
	if cfg.Exclude == nil {
		cfg.Exclude = DefaultExcludes
	}

	a := &analyzer{cfg: cfg}

	var err error
	a.headerMatcher, err = regexp.Compile(a.cfg.MatchHeaderRegexp)
	if err != nil {
		return nil, fmt.Errorf("compile match header regexp: %w", err)
	}

	// Compile exclude regexes.
	for _, exclude := range cfg.Exclude {
		if exclude == "" {
			continue
		}

		if strings.HasPrefix(exclude, "r!") {
			expr := strings.TrimPrefix(exclude, "r!")
			re, err := regexp.Compile(expr)
			if err != nil {
				return nil, fmt.Errorf("invalid exclude regexp pattern (%s): %w",
					expr, err)
			}
			a.excludes = append(a.excludes, func(filename string) bool {
				return re.MatchString(filename)
			})
			continue
		}

		if !doublestar.ValidatePattern(exclude) {
			return nil, fmt.Errorf("invalid exclude pattern: %s", exclude)
		}
		a.excludes = append(a.excludes, func(filename string) bool {
			matched, _ := doublestar.Match(exclude, filename)
			return matched
		})
	}

	// Create license header.
	a.header, err = NewHeader(cfg.Header)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	var errg errgroup.Group
	errg.SetLimit(a.cfg.MaxConcurrent)

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			// Skip generated files.
			continue
		}

		errg.Go(func() error {
			return a.checkFile(pass, file)
		})
	}
	return nil, errg.Wait()
}

func (a *analyzer) checkFile(pass *analysis.Pass, file *ast.File) error {
	// Check whether the file is excluded.
	filename := pass.Fset.File(file.Pos()).Name()
	for _, exclude := range a.excludes {
		if exclude(filename) {
			return nil
		}
	}

	var header string
	headerPos, headerEnd := file.FileStart, file.FileStart
	if len(file.Comments) > 0 {
		if c := file.Comments[0]; c.Pos() < file.Package {
			headerPos, headerEnd = c.Pos(), c.End()
			for _, comment := range c.List {
				header += comment.Text + "\n"
			}
		}
	}

	if header == "" || !a.headerMatcher.MatchString(header) {
		// License header is missing, generate a new one.
		header = a.header.Create(filename) + "\n"
		pass.Report(analysis.Diagnostic{
			Pos:      file.FileStart,
			Category: analyzerName,
			Message:  "missing license header",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "add license header",
				TextEdits: []analysis.TextEdit{{
					Pos:     file.FileStart,
					NewText: []byte(header),
				}},
			}},
		})
		return nil
	}

	if newHeader, modified := a.header.Update(filename, header); modified {
		pass.Report(analysis.Diagnostic{
			Pos:     headerPos,
			End:     headerEnd,
			Message: "invalid license header",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "update license header",
				TextEdits: []analysis.TextEdit{{
					Pos:     headerPos,
					End:     headerEnd,
					NewText: []byte(newHeader + "\n"),
				}},
			}},
		})
	}

	return nil
}
