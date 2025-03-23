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
	t.Run("simple", func(t *testing.T) {
		cfg := Config{
			Header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Test",
				YearMode: YearModeThisYear,
			},
			Exclude: []string{},
		}
		a, err := NewAnalyzer(cfg)
		if err != nil {
			t.Fatalf("NewAnalyzer() err = %v", err)
		}

		// Different header contains a file with a different license header.
		// This license header should stay as-is and should not be modified.
		t.Run("differentheader", func(t *testing.T) {
			packageDir := filepath.Join(analysistest.TestData(), "src/differentheader/")
			_ = analysistest.Run(t, packageDir, a)
		})

		// Empty contains an empty file without any existing license header and
		// creates a new header.
		t.Run("empty", func(t *testing.T) {
			packageDir := filepath.Join(analysistest.TestData(), "src/empty/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})

		// Outdated contains a file with a license header which has a different
		// copyright year. The year should be updated due to YearModeThisYear.
		t.Run("outdated", func(t *testing.T) {
			packageDir := filepath.Join(analysistest.TestData(), "src/outdated/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})

		// Package comment contains a file with a package-level doc comment and
		// no license header. A license header should be generated without
		// modifying the doc comment.
		t.Run("packagecomment", func(t *testing.T) {
			packageDir := filepath.Join(analysistest.TestData(), "src/packagecomment/")
			_ = analysistest.RunWithSuggestedFixes(t, packageDir, a)
		})
	})
}
