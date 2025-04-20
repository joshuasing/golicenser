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
	"regexp"
	"testing"
	"text/template"
	"time"
)

func init() {
	timeNow = func() time.Time {
		return time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
	}
}

// TODO(joshuasing): mock git and add test coverage for the Git year modes (fun).

func TestParseYearMode(t *testing.T) {
	t.Parallel()

	type parseTest struct {
		name    string
		s       string
		want    YearMode
		wantErr bool
	}
	tests := []parseTest{
		{
			name: "case insensitive",
			s:    "pReSerVe",
			want: YearModePreserve,
		},
		{
			name:    "invalid",
			s:       "invalid",
			wantErr: true,
		},
	}
	for ym, s := range yearModeStrings {
		tests = append(tests, parseTest{
			name: s,
			s:    s,
			want: ym,
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseYearMode(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYearMode(%q) err = %v, want %v",
					tt.s, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseYearMode(%q) = %v, want %v",
					tt.s, got, tt.want)
			}
		})
	}
}

func TestYearModeString(t *testing.T) {
	for ym, s := range yearModeStrings {
		if got := ym.String(); got != s {
			t.Errorf("YearMode(%d) = %s, want %s", ym, got, s)
		}
	}
}

func TestParseCommentStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		s       string
		want    CommentStyle
		wantErr bool
	}{
		{
			name: CommentStyleLine.String(),
			s:    CommentStyleLine.String(),
			want: CommentStyleLine,
		},
		{
			name: CommentStyleBlock.String(),
			s:    CommentStyleBlock.String(),
			want: CommentStyleBlock,
		},
		{
			name: "case insensitive",
			s:    "BlOcK",
			want: CommentStyleBlock,
		},
		{
			name:    "invalid",
			s:       "invalid",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseCommentStyle(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommentStyle(%q) err = %v, want %v",
					tt.s, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseCommentStyle(%q) = %v, want %v",
					tt.s, got, tt.want)
			}
		})
	}
}

func TestParseComment(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		want      string
		wantStyle CommentStyle
		wantErr   bool
	}{
		{
			name:      "line simple",
			in:        "// Hello world\n",
			want:      "Hello world",
			wantStyle: CommentStyleLine,
		},
		{
			name:      "line multi-line",
			in:        "// Line 1\n// Line 2\n// Line 3\n",
			want:      "Line 1\nLine 2\nLine 3",
			wantStyle: CommentStyleLine,
		},
		{
			name:      "line with blank line",
			in:        "// Line 1\n//\n// Line 2 after blank\n",
			want:      "Line 1\n\nLine 2 after blank",
			wantStyle: CommentStyleLine,
		},
		{
			name:      "line with leading space",
			in:        "//  Line 1\n//   Line 2\n",
			want:      " Line 1\n  Line 2", // one leading space then two spaces
			wantStyle: CommentStyleLine,
		},
		{
			name:      "block",
			in:        "/*\nHello world\n*/\n",
			want:      "Hello world",
			wantStyle: CommentStyleBlock,
		},
		{
			name:      "block singleline",
			in:        "/* Hello world */\n",
			want:      "Hello world",
			wantStyle: CommentStyleBlock,
		},
		{
			name:      "block multiline",
			in:        "/*\nLine 1\nLine 2\n*/\n",
			want:      "Line 1\nLine 2",
			wantStyle: CommentStyleBlock,
		},
		{
			name:      "block singleline no padding",
			in:        "/*Hello world*/\n",
			want:      "Hello world",
			wantStyle: CommentStyleBlock,
		},
		{
			name:      "starred block",
			in:        "/*\n * Hello world\n */\n",
			want:      "Hello world",
			wantStyle: CommentStyleStarredBlock,
		},
		{
			name:      "starred block multiline",
			in:        "/*\n * Line 1\n * Line 2\n */\n",
			want:      "Line 1\nLine 2",
			wantStyle: CommentStyleStarredBlock,
		},
		{
			name:      "starred block no spaces",
			in:        "/*\n *test\n *test 2\n */\n",
			want:      "test\ntest 2",
			wantStyle: CommentStyleStarredBlock,
		},
		{
			name: "starred block real",
			in: `/*
 * Copyright 2013 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */
`,
			want: `Copyright 2013 The Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.`,
			wantStyle: CommentStyleStarredBlock,
		},
		{
			name:      "starred block using first line",
			in:        "/* Test\n * Test 2\n */",
			want:      "Test\nTest 2",
			wantStyle: CommentStyleStarredBlock,
		},
		{
			name:    "invalid block",
			in:      "/* test",
			wantErr: true,
		},
		{
			name:    "random string",
			in:      "hello world",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, cs, err := parseComment(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseComment(%q) err = %v, want %v", tt.in, err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("parseComment(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if cs != tt.wantStyle {
				t.Errorf("parseComment(%q) style = %s, want %s", tt.in, cs, tt.wantStyle)
			}
		})
	}
}

func TestCommentStyleRender(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		want  string
		style CommentStyle
	}{
		{
			name:  "line simple",
			in:    "Hello world",
			want:  "// Hello world\n",
			style: CommentStyleLine,
		},
		{
			name:  "line multi-line",
			in:    "Line 1\nLine 2\nLine 3",
			want:  "// Line 1\n// Line 2\n// Line 3\n",
			style: CommentStyleLine,
		},
		{
			name:  "line with blank line",
			in:    "Line 1\n\nLine 2 after blank",
			want:  "// Line 1\n//\n// Line 2 after blank\n",
			style: CommentStyleLine,
		},
		{
			name:  "line with leading space",
			in:    " Line 1\n  Line 2", // one leading space then two spaces
			want:  "//  Line 1\n//   Line 2\n",
			style: CommentStyleLine,
		},
		{
			name:  "block",
			in:    "Hello world",
			want:  "/*\nHello world\n*/\n",
			style: CommentStyleBlock,
		},
		{
			name:  "block mutiline",
			in:    "Line 1\nLine 2",
			want:  "/*\nLine 1\nLine 2\n*/\n",
			style: CommentStyleBlock,
		},
		{
			name:  "starred block",
			in:    "Hello world",
			want:  "/*\n * Hello world\n */\n",
			style: CommentStyleStarredBlock,
		},
		{
			name: "starred block multiline",
			in:   "Line 1\nLine 2\nLine 3",
			want: `/*
 * Line 1
 * Line 2
 * Line 3
 */
`,
			style: CommentStyleStarredBlock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.style.Render(tt.in); got != tt.want {
				t.Errorf("CommentStyle(%+v).Render(%q) = %q, want %q",
					tt.style, tt.in, got, tt.want)
			}
		})
	}
}

func TestNewHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  HeaderOpts
		wantErr bool
	}{
		{
			name: "simple",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Joshua Sing",
			},
		},
		{
			name: "with comment style",
			header: HeaderOpts{
				Template:     "Copyright (c) {{.year}} {{.author}}",
				Author:       "Joshua Sing",
				CommentStyle: CommentStyleBlock,
			},
		},
		{
			name: "missing author",
			header: HeaderOpts{
				Template: "Test",
			},
			wantErr: true,
		},
		{
			name: "with author regexp",
			header: HeaderOpts{
				Template:     "{{.author}}",
				Author:       "Test",
				AuthorRegexp: "(Test|Someone)",
			},
		},
		{
			name: "with invalid author regexp",
			header: HeaderOpts{
				Template:     "{{.author}}",
				Author:       "Test",
				AuthorRegexp: "(Test",
			},
			wantErr: true,
		},
		{
			name: "missing template",
			header: HeaderOpts{
				Author: "Test",
			},
			wantErr: true,
		},
		{
			name: "invalid template",
			header: HeaderOpts{
				Template: "Test {{{ . }}",
				Author:   "test",
			},
			wantErr: true,
		},
		{
			name: "use non-existent variable",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.human}}",
				Author:   "Test",
			},
			wantErr: true,
		},
		{
			name: "basename func",
			header: HeaderOpts{
				Template: "This file is {{basename .filename}}",
				Author:   "test",
			},
		},
		{
			name: "custom variables",
			header: HeaderOpts{
				Template: "{{.project}} by {{.person}}",
				Author:   "test",
				Variables: map[string]*Var{
					"project": {Value: "project"},
					"person":  {Value: "person"},
				},
			},
		},
		{
			name: "custom variables with regexp",
			header: HeaderOpts{
				Template: "{{.project}} by {{.person}}",
				Author:   "test",
				Variables: map[string]*Var{
					"project": {Value: "project", Regexp: "(golicenser|project)"},
					"person":  {Value: "human", Regexp: "(human|person)"},
				},
			},
		},
		{
			name: "custom variables with invalid regexp",
			header: HeaderOpts{
				Template: "{{.project}} by {{.person}}",
				Author:   "test",
				Variables: map[string]*Var{
					"project": {Value: "project", Regexp: "(project"},
					"person":  {Value: "person", Regexp: "person)"},
				},
			},
			wantErr: true,
		},
		{
			name: "with matcher",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Test",
				Matcher:  "Copyright \\(c\\) \\d{4} (Test|Someone)",
			},
		},
		{
			name: "invalid matcher template",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Test",
				Matcher:  "Copyright \\(c\\) {{}.year}} (Test)",
			},
			wantErr: true,
		},
		{
			name: "invalid matcher regexp",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Test",
				Matcher:  "Copyright \\(c\\) {{.year}} (Test",
			},
			wantErr: true,
		},
		{
			name: "MIT",
			header: HeaderOpts{
				Template: LicenseMIT,
				Author:   "Joshua Sing",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHeader(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHeader err = %v, want err %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeaderCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   HeaderOpts
		filename string

		want    string
		wantErr bool // from NewHeader
	}{
		{
			name: "simple",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Joshua Sing",
			},
			want: "// Copyright (c) 2025 Joshua Sing\n",
		},
		{
			name: "block comment",
			header: HeaderOpts{
				Template:     "Copyright (c) {{.year}} {{.author}}",
				Author:       "Joshua Sing",
				CommentStyle: CommentStyleBlock,
			},
			want: "/*\nCopyright (c) 2025 Joshua Sing\n*/\n",
		},
		{
			name: "use filename",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}\nThis file is {{.filename}}.",
				Author:   "Joshua Sing",
			},
			filename: "header_test.go",
			want:     "// Copyright (c) 2025 Joshua Sing\n// This file is header_test.go.\n",
		},
		{
			name: "basename func",
			header: HeaderOpts{
				Template: "This file is {{basename .filename}}",
				Author:   "test",
			},
			filename: "/path/to/this/header_test.go",
			want:     "// This file is header_test.go\n",
		},
		{
			name: "custom variables",
			header: HeaderOpts{
				Template: "{{.project}} by {{.person}}",
				Author:   "test",
				Variables: map[string]*Var{
					"project": {Value: "project"},
					"person":  {Value: "person"},
				},
			},
			want: "// project by person\n",
		},
		{
			name: "custom variables with regexp",
			header: HeaderOpts{
				Template: "{{.project}} by {{.person}}",
				Author:   "test",
				Variables: map[string]*Var{
					"project": {Value: "project", Regexp: "(golicenser|project)"},
					"person":  {Value: "human", Regexp: "(human|person)"},
				},
			},
			want: "// project by human\n",
		},
		{
			name: "MIT",
			header: HeaderOpts{
				Template: LicenseMIT,
				Author:   "Joshua Sing",
			},
			want: `// Copyright (c) 2025 Joshua Sing
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
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, err := NewHeader(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHeader err = %v, want err %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			header, err := h.Create(tt.filename)
			if err != nil {
				t.Errorf("h.Create(%q) err = %v, want nil", header, err)
			}
			if got := header; got != tt.want {
				t.Errorf("h.Create(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestHeaderUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   HeaderOpts
		filename string

		existing     string
		want         string
		wantModified bool
		wantErr      bool // from NewHeader
	}{
		{
			name: "no change",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Joshua Sing",
			},
			existing: "// Copyright (c) 2025 Joshua Sing\n",
			want:     "// Copyright (c) 2025 Joshua Sing\n",
		},
		{
			name: "preserve year",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Joshua Sing",
				YearMode: YearModePreserve,
			},
			existing: "// Copyright (c) 2001 Joshua Sing\n",
			want:     "// Copyright (c) 2001 Joshua Sing\n",
		},
		{
			name: "change year",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}",
				Author:   "Joshua Sing",
				YearMode: YearModeThisYear,
			},
			existing:     "// Copyright (c) 2001 Joshua Sing\n",
			want:         "// Copyright (c) 2025 Joshua Sing\n",
			wantModified: true,
		},
		{
			name: "change block comment to line comment",
			header: HeaderOpts{
				Template:     "Copyright (c) {{.year}} {{.author}}",
				Author:       "Joshua Sing",
				CommentStyle: CommentStyleLine,
			},
			existing:     "/*\nCopyright (c) 2025 Joshua Sing\n*/\n",
			want:         "// Copyright (c) 2025 Joshua Sing\n",
			wantModified: true,
		},
		{
			name: "change line comment to block comment",
			header: HeaderOpts{
				Template:     "Copyright (c) {{.year}} {{.author}}",
				Author:       "Joshua Sing",
				CommentStyle: CommentStyleBlock,
			},
			existing:     "// Copyright (c) 2025 Joshua Sing\n",
			want:         "/*\nCopyright (c) 2025 Joshua Sing\n*/\n",
			wantModified: true,
		},
		{
			name: "custom variables",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}\nProject: {{.project}}, Greet: {{.greet}}",
				Author:   "Joshua Sing",
				YearMode: YearModeThisYear,
				Variables: map[string]*Var{
					"project": {Value: "project"},
					"greet":   {Value: "Hello world"},
				},
			},
			existing:     "// Copyright (c) 2024 Joshua Sing\n// Project: project, Greet: Hello world",
			want:         "// Copyright (c) 2025 Joshua Sing\n// Project: project, Greet: Hello world\n",
			wantModified: true,
		},
		{
			name: "custom variables with regexp",
			header: HeaderOpts{
				Template: "Copyright (c) {{.year}} {{.author}}\nProject: {{.project}}, Greet: {{.greet}}",
				Author:   "Joshua Sing",
				Variables: map[string]*Var{
					"project": {Value: "project"},
					"greet":   {Value: "Hello world", Regexp: "Hello (.+)"},
				},
			},
			existing:     "// Copyright (c) 2025 Joshua Sing\n// Project: project, Greet: Hello there!",
			want:         "// Copyright (c) 2025 Joshua Sing\n// Project: project, Greet: Hello world\n",
			wantModified: true,
		},
		{
			name: "change MIT to OpenBSD",
			header: HeaderOpts{
				Template:      LicenseOpenBSD,
				Matcher:       LicenseMIT,
				MatcherEscape: true,
				Author:        "Joshua Sing",
			},
			wantModified: true,
			existing: `// Copyright (c) 2025 Joshua Sing
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
`,
			want: `// Copyright (c) 2025 Joshua Sing
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
`,
		},
		{
			name: "change MIT block to OpenBSD line",
			header: HeaderOpts{
				Template:      LicenseOpenBSD,
				Matcher:       LicenseMIT,
				MatcherEscape: true,
				CommentStyle:  CommentStyleLine,
				Author:        "Joshua Sing",
			},
			wantModified: true,
			existing: `/*
Copyright (c) 2025 Joshua Sing

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/`,
			want: `// Copyright (c) 2025 Joshua Sing
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, err := NewHeader(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHeader err = %v, want err %v", err, tt.wantErr)
			}
			got, modified, err := h.Update(tt.filename, tt.existing)
			if err != nil {
				t.Errorf("h.Update() err = %v, want nil", err)
			}
			if modified != tt.wantModified {
				t.Errorf("h.Update() modified = %v, want %v",
					modified, tt.wantModified)
			}
			if got != tt.want {
				t.Errorf("h.Update() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHeaderMatcher(t *testing.T) {
	t.Parallel()

	type matchTest struct {
		name      string
		input     string
		wantMatch bool
	}

	tests := []struct {
		name         string
		matcher      string
		escape       bool
		variables    map[string]*Var
		authorRegexp *regexp.Regexp
		wantErr      bool
		matchTests   []matchTest
	}{
		{
			name:         "basic escaped",
			matcher:      "Copyright (c) {{.year}} {{.author}}\nFile: {{.filename}}",
			escape:       true,
			authorRegexp: regexp.MustCompile("Test"),
			matchTests: []matchTest{
				{
					name:      "empty",
					input:     "",
					wantMatch: false,
				},
				{
					name:      "random",
					input:     "Hello world",
					wantMatch: false,
				},
				{
					name:      "exact match",
					input:     "Copyright (c) 2025 Test\nFile: header_test.go",
					wantMatch: true,
				},
				{
					name:      "different year",
					input:     "Copyright (c) 2000 Test\nFile: header_test.go",
					wantMatch: true,
				},
			},
		},
		{
			name:    "custom variables",
			matcher: "{{.project}} by {{.name}} - Copyright (c) {{.year}} {{.author}}",
			escape:  true,
			variables: map[string]*Var{
				"project": {Value: "golicenser", Regexp: "golicenser"},
				"name":    {Value: "joshuasing", Regexp: "joshuasing"},
			},
			authorRegexp: regexp.MustCompile("Test"),
			matchTests: []matchTest{
				{
					name:      "exact match",
					input:     "golicenser by joshuasing - Copyright (c) 2025 Test",
					wantMatch: true,
				},
				{
					name:      "different variable values",
					input:     "golicenser by someone - Copyright (c) 2025 Test",
					wantMatch: false,
				},
			},
		},
		{
			name:    "custom variables with regexp",
			matcher: "{{.project}} by {{.name}} - Copyright (c) {{.year}} {{.author}}",
			escape:  true,
			variables: map[string]*Var{
				"project": {Value: "golicenser", Regexp: "go-?licenser"},
				"name":    {Value: "joshuasing", Regexp: "(joshuasing|someone)"},
			},
			authorRegexp: regexp.MustCompile("Test"),
			matchTests: []matchTest{
				{
					name:      "exact match",
					input:     "golicenser by joshuasing - Copyright (c) 2025 Test",
					wantMatch: true,
				},
				{
					name:      "one different matched",
					input:     "go-licenser by joshuasing - Copyright (c) 2025 Test",
					wantMatch: true,
				},
				{
					name:      "both different matched",
					input:     "go-licenser by someone - Copyright (c) 2025 Test",
					wantMatch: true,
				},
				{
					name:      "unmatched variable value",
					input:     "project by someone - Copyright (c) 2025 Test",
					wantMatch: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpl, err := template.New("").Parse(tt.matcher)
			if err != nil {
				t.Fatalf("compile template: %v", err)
			}

			matcher, err := headerMatcher(tmpl, tt.escape, tt.authorRegexp, tt.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("headerMatcher err = %v, want err %v", err, tt.wantErr)
			}
			t.Logf("Matcher: %v", matcher.String())

			for _, mt := range tt.matchTests {
				t.Run(mt.name, func(t *testing.T) {
					if got := matcher.MatchString(mt.input); got != mt.wantMatch {
						t.Errorf("MatchString(%q) = %v, want %v",
							mt.input, got, mt.wantMatch)
					}
				})
			}
		})
	}
}
