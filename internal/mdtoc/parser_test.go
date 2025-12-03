package mdtoc

import (
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		opts     Options
		expected []*Header
	}{
		{
			name: "basic headers",
			content: `# Title
## Section 1
### Subsection 1.1
## Section 2`,
			opts: DefaultOptions(),
			expected: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
				{Level: 3, Text: "Subsection 1.1", AnchorLink: "subsection-11"},
				{Level: 2, Text: "Section 2", AnchorLink: "section-2"},
			},
		},
		{
			name: "with min level filter",
			content: `# Title
## Section 1
### Subsection`,
			opts: Options{MinLevel: 2, MaxLevel: 3},
			expected: []*Header{
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
				{Level: 3, Text: "Subsection", AnchorLink: "subsection"},
			},
		},
		{
			name: "with max level filter",
			content: `# Title
## Section 1
### Subsection
#### Deep`,
			opts: Options{MinLevel: 1, MaxLevel: 2},
			expected: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
			},
		},
		{
			name: "headers in code block ignored",
			content: "# Real Header\n```\n# Not a header\n```\n## Another Header",
			opts:    DefaultOptions(),
			expected: []*Header{
				{Level: 1, Text: "Real Header", AnchorLink: "real-header"},
				{Level: 2, Text: "Another Header", AnchorLink: "another-header"},
			},
		},
		{
			name: "duplicate headers",
			content: `# Title
## Section
## Section
## Section`,
			opts: DefaultOptions(),
			expected: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section", AnchorLink: "section"},
				{Level: 2, Text: "Section", AnchorLink: "section-1"},
				{Level: 2, Text: "Section", AnchorLink: "section-2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.opts)
			got, err := p.Parse([]byte(tt.content))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("Parse() returned %d headers, want %d", len(got), len(tt.expected))
			}

			for i, h := range got {
				exp := tt.expected[i]
				if h.Level != exp.Level {
					t.Errorf("Header[%d].Level = %d, want %d", i, h.Level, exp.Level)
				}
				if h.Text != exp.Text {
					t.Errorf("Header[%d].Text = %q, want %q", i, h.Text, exp.Text)
				}
				if h.AnchorLink != exp.AnchorLink {
					t.Errorf("Header[%d].AnchorLink = %q, want %q", i, h.AnchorLink, exp.AnchorLink)
				}
			}
		})
	}
}

func TestParser_EmptyContent(t *testing.T) {
	p := NewParser(DefaultOptions())
	got, err := p.Parse([]byte(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Parse() returned %d headers for empty content, want 0", len(got))
	}
}

func TestParser_NoHeaders(t *testing.T) {
	p := NewParser(DefaultOptions())
	content := `This is a paragraph.

Another paragraph with some **bold** text.

And a list:
- Item 1
- Item 2
`
	got, err := p.Parse([]byte(content))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Parse() returned %d headers for content without headers, want 0", len(got))
	}
}

func TestParser_LineNumbers(t *testing.T) {
	content := `# Title

## Section 1
Content...

## Section 2
More content...
`
	p := NewParser(DefaultOptions())
	got, err := p.Parse([]byte(content))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []struct {
		text    string
		line    int
		endLine int
	}{
		{"Title", 1, 7},      // H1 包含所有 H2 子内容
		{"Section 1", 3, 5},  // H2 到下一个 H2 前
		{"Section 2", 6, 7},  // H2 到文件末尾
	}

	if len(got) != len(expected) {
		t.Fatalf("Parse() returned %d headers, want %d", len(got), len(expected))
	}

	for i, h := range got {
		exp := expected[i]
		if h.Text != exp.text {
			t.Errorf("Header[%d].Text = %q, want %q", i, h.Text, exp.text)
		}
		if h.Line != exp.line {
			t.Errorf("Header[%d].Line = %d, want %d", i, h.Line, exp.line)
		}
		if h.EndLine != exp.endLine {
			t.Errorf("Header[%d].EndLine = %d, want %d", i, h.EndLine, exp.endLine)
		}
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		content  string
		expected int
	}{
		{"", 0},
		{"line1", 1},
		{"line1\n", 1},
		{"line1\nline2", 2},
		{"line1\nline2\n", 2},
		{"line1\nline2\nline3", 3},
	}

	for _, tt := range tests {
		got := countLines([]byte(tt.content))
		if got != tt.expected {
			t.Errorf("countLines(%q) = %d, want %d", tt.content, got, tt.expected)
		}
	}
}

func TestSplitSections(t *testing.T) {
	tests := []struct {
		name     string
		headers  []*Header
		expected []struct {
			h1Text      string
			subCount    int
			subTexts    []string
		}
	}{
		{
			name: "multiple H1 with sub-headers",
			headers: []*Header{
				{Level: 1, Text: "Chapter 1"},
				{Level: 2, Text: "Section 1.1"},
				{Level: 2, Text: "Section 1.2"},
				{Level: 1, Text: "Chapter 2"},
				{Level: 2, Text: "Section 2.1"},
				{Level: 3, Text: "Subsection 2.1.1"},
			},
			expected: []struct {
				h1Text   string
				subCount int
				subTexts []string
			}{
				{"Chapter 1", 2, []string{"Section 1.1", "Section 1.2"}},
				{"Chapter 2", 2, []string{"Section 2.1", "Subsection 2.1.1"}},
			},
		},
		{
			name: "H1 without sub-headers",
			headers: []*Header{
				{Level: 1, Text: "Chapter 1"},
				{Level: 2, Text: "Section 1.1"},
				{Level: 1, Text: "Chapter 2"},
				{Level: 1, Text: "Chapter 3"},
			},
			expected: []struct {
				h1Text   string
				subCount int
				subTexts []string
			}{
				{"Chapter 1", 1, []string{"Section 1.1"}},
				{"Chapter 2", 0, nil},
				{"Chapter 3", 0, nil},
			},
		},
		{
			name: "headers before first H1 ignored",
			headers: []*Header{
				{Level: 2, Text: "Orphan Section"},
				{Level: 3, Text: "Orphan Subsection"},
				{Level: 1, Text: "Chapter 1"},
				{Level: 2, Text: "Section 1.1"},
			},
			expected: []struct {
				h1Text   string
				subCount int
				subTexts []string
			}{
				{"Chapter 1", 1, []string{"Section 1.1"}},
			},
		},
		{
			name:    "no H1 headers",
			headers: []*Header{
				{Level: 2, Text: "Section 1"},
				{Level: 2, Text: "Section 2"},
			},
			expected: nil,
		},
		{
			name:     "empty headers",
			headers:  []*Header{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := SplitSections(tt.headers)

			if tt.expected == nil {
				if len(sections) != 0 {
					t.Errorf("SplitSections() returned %d sections, want 0", len(sections))
				}
				return
			}

			if len(sections) != len(tt.expected) {
				t.Fatalf("SplitSections() returned %d sections, want %d", len(sections), len(tt.expected))
			}

			for i, s := range sections {
				exp := tt.expected[i]
				if s.Title.Text != exp.h1Text {
					t.Errorf("Section[%d].Title.Text = %q, want %q", i, s.Title.Text, exp.h1Text)
				}
				if len(s.SubHeaders) != exp.subCount {
					t.Errorf("Section[%d] has %d sub-headers, want %d", i, len(s.SubHeaders), exp.subCount)
				}
				for j, sub := range s.SubHeaders {
					if j < len(exp.subTexts) && sub.Text != exp.subTexts[j] {
						t.Errorf("Section[%d].SubHeaders[%d].Text = %q, want %q", i, j, sub.Text, exp.subTexts[j])
					}
				}
			}
		})
	}
}

func TestParser_ParseAllHeaders(t *testing.T) {
	content := `# Title
## Section 1
### Subsection 1.1
#### Deep 1
## Section 2
`
	// ParseAllHeaders should return ALL headers regardless of MinLevel/MaxLevel
	p := NewParser(Options{MinLevel: 2, MaxLevel: 2})
	got, err := p.ParseAllHeaders([]byte(content))
	if err != nil {
		t.Fatalf("ParseAllHeaders() error = %v", err)
	}

	// Should return all 5 headers, not filtered by level
	if len(got) != 5 {
		t.Errorf("ParseAllHeaders() returned %d headers, want 5", len(got))
	}

	expectedLevels := []int{1, 2, 3, 4, 2}
	for i, h := range got {
		if h.Level != expectedLevels[i] {
			t.Errorf("Header[%d].Level = %d, want %d", i, h.Level, expectedLevels[i])
		}
	}
}
