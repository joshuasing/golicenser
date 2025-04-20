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
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"golang.org/x/tools/go/analysis/analysistest"
)

func init() {
	timeNow = func() time.Time {
		return time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
	}
}

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			Header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Test",
				YearMode: YearModeThisYear,
			},
		}
		a, err := NewAnalyzer(cfg)
		if err != nil {
			t.Fatalf("NewAnalyzer() err = %v", err)
		}

		// Test with sources containing build directives.
		t.Run("builddirective", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/builddirective/")
			_ = analysistest.Run(t, packageDir, a)
		})

		// Different header contains a file with a different license header.
		// This license header should stay as-is and should not be modified.
		t.Run("differentheader", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/differentheader/")
			_ = analysistest.Run(t, packageDir, a)
		})

		// Empty contains an empty file without any existing license header and
		// creates a new header.
		t.Run("empty", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/empty/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})

		// Outdated contains a file with a license header which has a different
		// copyright year. The year should be updated due to YearModeThisYear.
		t.Run("outdated", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/outdated/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})

		// Package comment contains a file with a package-level doc comment and
		// no license header. A license header should be generated without
		// modifying the doc comment.
		t.Run("packagecomment", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/packagecomment/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})
	})

	t.Run("concurrency", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Test",
				YearMode: YearModeThisYear,
			},
			MaxConcurrent: runtime.NumCPU() * 2,
		}
		a, err := NewAnalyzer(cfg)
		if err != nil {
			t.Fatalf("NewAnalyzer() err = %v", err)
		}

		// Different header contains a file with a different license header.
		// This license header should stay as-is and should not be modified.
		t.Run("differentheader", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/differentheader/")
			_ = analysistest.Run(t, packageDir, a)
		})

		// Empty contains an empty file without any existing license header and
		// creates a new header.
		t.Run("empty", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/empty/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})

		// Outdated contains a file with a license header which has a different
		// copyright year. The year should be updated due to YearModeThisYear.
		t.Run("outdated", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/outdated/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})

		// Package comment contains a file with a package-level doc comment and
		// no license header. A license header should be generated without
		// modifying the doc comment.
		t.Run("packagecomment", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/packagecomment/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})
	})

	t.Run("with matcher", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			Header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Test",
				Matcher:  "Copyright \\(c\\) {{.year}} Joshua",
				YearMode: YearModeThisYear,
			},
		}
		a, err := NewAnalyzer(cfg)
		if err != nil {
			t.Fatalf("NewAnalyzer() err = %v", err)
		}

		// differentmatcher contains a header that will be matched by Matcher
		// but is different from the Template.
		t.Run("differentmatcher", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/differentmatcher/")
			_ = analysistest.Run(t, packageDir, a)
		})
	})

	t.Run("with escaped matcher", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			Header: HeaderOpts{
				Template:      "Copyright (c) {{.year}} {{.author}}",
				Author:        "Test",
				Matcher:       "Copyright (c) {{.year}} Joshua",
				MatcherEscape: true,
				YearMode:      YearModeThisYear,
			},
			Exclude: []string{},
		}
		a, err := NewAnalyzer(cfg)
		if err != nil {
			t.Fatalf("NewAnalyzer() err = %v", err)
		}

		// differentmatcher contains a header that will be matched with the
		// escaped Matcher but is different from the Template.
		t.Run("differentmatcher", func(t *testing.T) {
			t.Parallel()
			packageDir := filepath.Join(analysistest.TestData(), "src/differentmatcher/")
			_ = analysistest.Run(t, packageDir, a)
		})
	})

	t.Run("build directive with any matcher", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			Header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Test",
				YearMode: YearModeThisYear,
			},
			CopyrightHeaderMatcher: ".+",
		}
		a, err := NewAnalyzer(cfg)
		if err != nil {
			t.Fatalf("NewAnalyzer() err = %v", err)
		}

		packageDir := filepath.Join(analysistest.TestData(), "src/builddirective/")
		_ = analysistest.Run(t, packageDir, a)
	})

	t.Run("comment style block", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			Header: HeaderOpts{
				Template:     "Copyright (c) {{.year}} {{.author}}",
				Author:       "Test",
				YearMode:     YearModeThisYear,
				CommentStyle: CommentStyleBlock,
			},
		}
		a, err := NewAnalyzer(cfg)
		if err != nil {
			t.Fatalf("NewAnalyzer() err = %v", err)
		}

		packageDir := filepath.Join(analysistest.TestData(), "src/block/")
		_ = analysistest.Run(t, packageDir, a)
	})

	t.Run("comment style starred block", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			Header: HeaderOpts{
				Template:     "Copyright (c) {{.year}} {{.author}}",
				Author:       "Test",
				YearMode:     YearModeThisYear,
				CommentStyle: CommentStyleStarredBlock,
			},
		}
		a, err := NewAnalyzer(cfg)
		if err != nil {
			t.Fatalf("NewAnalyzer() err = %v", err)
		}

		packageDir := filepath.Join(analysistest.TestData(), "src/starred-block/")
		_ = analysistest.Run(t, packageDir, a)
	})
}

func TestNewAnalyzer(t *testing.T) {
	t.Parallel()

	header := HeaderOpts{
		Template: "test",
		Author:   "test",
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		check   func(t *testing.T, a *analyzer)
	}{
		{
			name: "defaults",
			cfg: Config{
				Header: header,
			},
			check: func(t *testing.T, a *analyzer) {
				t.Helper()
				if a.cfg.CopyrightHeaderMatcher != DefaultCopyrightHeaderMatcher {
					t.Errorf("CopyrightHeaderMatcher = %v, want %v",
						a.cfg.CopyrightHeaderMatcher, DefaultCopyrightHeaderMatcher)
				}
			},
		},
		{
			name: "invalid copyright header matcher",
			cfg: Config{
				Header:                 header,
				CopyrightHeaderMatcher: "(test",
			},
			wantErr: true,
		},
		{
			name: "excludes",
			cfg: Config{
				Header: header,
				Exclude: []string{
					"/abc/*",
					"", // empty strings should be ignored
					"/test/**",
				},
			},
			check: func(t *testing.T, a *analyzer) {
				t.Helper()

				if l := len(a.excludes); l != 2 {
					t.Errorf("excludes len = %d, want 2", l)
				}
				tests := map[string]bool{
					"afile.go":       false,
					"/subdir/test":   false,
					"/abc/":          true,
					"/abc/test":      true,
					"/test/somefile": true,
					"/test/":         true,
				}
				for path, want := range tests {
					var excluded bool
					for _, exclude := range a.excludes {
						if !excluded && exclude(path) {
							excluded = true
						}
					}
					if excluded != want {
						t.Errorf("exclude(%q) = %v, want %v", path, excluded, want)
					}
				}
			},
		},
		{
			name: "excludes regex",
			cfg: Config{
				Header: header,
				Exclude: []string{
					"r!(.+)_test\\.go",
					"", // empty strings should be ignored
					"r!(dir[0-9])/(test1|test2)\\.go",
				},
			},
			check: func(t *testing.T, a *analyzer) {
				t.Helper()

				if l := len(a.excludes); l != 2 {
					t.Errorf("excludes len = %d, want 2", l)
				}
				tests := map[string]bool{
					"afile.go":                   false,
					"/subdir/test":               false,
					"/golicenser_test.go":        true,
					"/subdir/golicenser_test.go": true,
					"/dir1/test1.go":             true,
					"/dir2/test2.go":             true,
					"/subdir/dir1/test1.go":      true,
					"/dir1/otherfile.go":         false,
				}
				for path, want := range tests {
					var excluded bool
					for _, exclude := range a.excludes {
						if !excluded && exclude(path) {
							excluded = true
						}
					}
					if excluded != want {
						t.Errorf("exclude(%q) = %v, want %v", path, excluded, want)
					}
				}
			},
		},
		{
			name: "excludes invalid regex",
			cfg: Config{
				Header: header,
				Exclude: []string{
					"/abc/*",
					"r!test)",
				},
			},
			wantErr: true,
		},
		{
			name: "excludes invalid doublestar",
			cfg: Config{
				Header: header,
				Exclude: []string{
					"/abc/*",
					"**/test/*{*",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, err := newAnalyzer(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("newAnalyzer err = %v, want %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, a)
			}
		})
	}
}
