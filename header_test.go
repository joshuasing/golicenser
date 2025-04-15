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
	"testing"
	"time"
)

func init() {
	timeNow = func() time.Time {
		return time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
	}
}

// TODO(joshuasing): mock git and add test coverage for the Git year modes (fun).

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
			in:    "Hello world",
			want:  "/*\nHello world\n*/\n",
			style: CommentStyleBlock,
		},
		{
			in:    "Line 1\nLine 2",
			want:  "/*\nLine 1\nLine 2\n*/\n",
			style: CommentStyleBlock,
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

func TestCommentStyleParse(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		want  string
		style CommentStyle
	}{
		{
			name:  "line simple",
			in:    "// Hello world\n",
			want:  "Hello world",
			style: CommentStyleLine,
		},
		{
			name:  "line multi-line",
			in:    "// Line 1\n// Line 2\n// Line 3\n",
			want:  "Line 1\nLine 2\nLine 3",
			style: CommentStyleLine,
		},
		{
			name:  "line with blank line",
			in:    "// Line 1\n//\n// Line 2 after blank\n",
			want:  "Line 1\n\nLine 2 after blank",
			style: CommentStyleLine,
		},
		{
			name:  "line with leading space",
			in:    "//  Line 1\n//   Line 2\n",
			want:  " Line 1\n  Line 2", // one leading space then two spaces
			style: CommentStyleLine,
		},
		{
			in:    "/*\nHello world\n*/\n",
			want:  "Hello world",
			style: CommentStyleBlock,
		},
		{
			in:    "/*\nLine 1\nLine 2\n*/\n",
			want:  "Line 1\nLine 2",
			style: CommentStyleBlock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.style.Parse(tt.in); got != tt.want {
				t.Errorf("CommentStyle(%+v).Parse(%q) = %q, want %q",
					tt.style, tt.in, got, tt.want)
			}
		})
	}
}

func TestHeaderCreate(t *testing.T) {
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
			h, err := NewHeader(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHeader err = %v, want err %v", err, tt.wantErr)
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
			name: "change MIT to OpenBSD",
			header: HeaderOpts{
				Template:      LicenseOpenBSD,
				MatchTemplate: LicenseMIT,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
