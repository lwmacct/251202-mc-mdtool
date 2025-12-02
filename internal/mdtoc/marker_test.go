package mdtoc

import (
	"strings"
	"testing"
)

func TestMarkerHandler_FindMarkers(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantStart int
		wantEnd   int
		wantFound bool
	}{
		{
			name:      "no markers",
			content:   "# Title\nSome content",
			wantStart: -1,
			wantEnd:   -1,
			wantFound: false,
		},
		{
			name:      "one marker",
			content:   "# Title\n<!--TOC-->\nSome content",
			wantStart: 1,
			wantEnd:   -1,
			wantFound: true,
		},
		{
			name:      "two markers",
			content:   "# Title\n<!--TOC-->\nTOC content\n<!--TOC-->\nRest",
			wantStart: 1,
			wantEnd:   3,
			wantFound: true,
		},
		{
			name:      "marker with whitespace",
			content:   "# Title\n  <!--TOC-->  \nSome content",
			wantStart: 1,
			wantEnd:   -1,
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindMarkers([]byte(tt.content))

			if got.StartLine != tt.wantStart {
				t.Errorf("StartLine = %d, want %d", got.StartLine, tt.wantStart)
			}
			if got.EndLine != tt.wantEnd {
				t.Errorf("EndLine = %d, want %d", got.EndLine, tt.wantEnd)
			}
			if got.Found != tt.wantFound {
				t.Errorf("Found = %v, want %v", got.Found, tt.wantFound)
			}
		})
	}
}

func TestMarkerHandler_InsertTOC(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		toc      string
		expected string
	}{
		{
			name:     "no marker - unchanged",
			content:  "# Title\nContent",
			toc:      "- [Title](#title)",
			expected: "# Title\nContent",
		},
		{
			name:    "one marker - insert",
			content: "# Title\n<!--TOC-->\nContent",
			toc:     "- [Title](#title)",
			expected: `# Title
<!--TOC-->

- [Title](#title)

<!--TOC-->
Content`,
		},
		{
			name: "two markers - replace",
			content: `# Title
<!--TOC-->
Old TOC
<!--TOC-->
Content`,
			toc: "- [Title](#title)",
			expected: `# Title
<!--TOC-->

- [Title](#title)

<!--TOC-->
Content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertTOC([]byte(tt.content), tt.toc))

			if got != tt.expected {
				t.Errorf("InsertTOC() =\n%s\n\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_ExtractExistingTOC(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "no markers",
			content:  "# Title\nContent",
			expected: "",
		},
		{
			name:     "one marker",
			content:  "# Title\n<!--TOC-->\nContent",
			expected: "",
		},
		{
			name: "two markers with content",
			content: `# Title
<!--TOC-->

- [Section](#section)

<!--TOC-->
Content`,
			expected: "- [Section](#section)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.ExtractExistingTOC([]byte(tt.content))

			if got != tt.expected {
				t.Errorf("ExtractExistingTOC() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_FindH1Lines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []int
	}{
		{
			name:     "no H1",
			content:  "## Section\n### Subsection",
			expected: nil,
		},
		{
			name:     "single H1",
			content:  "# Title\n## Section",
			expected: []int{0},
		},
		{
			name: "multiple H1",
			content: `# Chapter 1
## Section 1.1
# Chapter 2
## Section 2.1
# Chapter 3`,
			expected: []int{0, 2, 4},
		},
		{
			name: "H1 in code block ignored",
			content: "# Real H1\n```\n# Not H1\n```\n# Another H1",
			expected: []int{0, 4},
		},
		{
			name:     "H2 not matched",
			content:  "## Not H1\n### Also not H1",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindH1Lines([]byte(tt.content))

			if len(got) != len(tt.expected) {
				t.Fatalf("FindH1Lines() returned %d lines, want %d", len(got), len(tt.expected))
			}
			for i, line := range got {
				if line != tt.expected[i] {
					t.Errorf("FindH1Lines()[%d] = %d, want %d", i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestMarkerHandler_InsertSectionTOCs(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		sectionTOCs []SectionTOC
		expected    string
	}{
		{
			name:        "empty section TOCs",
			content:     "# Title\nContent",
			sectionTOCs: []SectionTOC{},
			expected:    "# Title\nContent",
		},
		{
			name:    "single section TOC",
			content: "# Chapter 1\n\nContent...\n\n## Section 1.1\n\nMore content",
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
			},
			expected: `# Chapter 1

<!--TOC-->

- [Section 1.1](#section-11)

<!--TOC-->

Content...

## Section 1.1

More content`,
		},
		{
			name: "multiple section TOCs",
			content: `# Chapter 1

## Section 1.1

# Chapter 2

## Section 2.1`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
				{H1Line: 4, TOC: "- [Section 2.1](#section-21)"},
			},
			expected: `# Chapter 1

<!--TOC-->

- [Section 1.1](#section-11)

<!--TOC-->

## Section 1.1

# Chapter 2

<!--TOC-->

- [Section 2.1](#section-21)

<!--TOC-->

## Section 2.1`,
		},
		{
			name: "section without sub-headers skipped",
			content: `# Chapter 1

## Section 1.1

# Chapter 2

No sub-headers here`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
				// Chapter 2 has empty TOC, should be skipped
				{H1Line: 4, TOC: ""},
			},
			expected: `# Chapter 1

<!--TOC-->

- [Section 1.1](#section-11)

<!--TOC-->

## Section 1.1

# Chapter 2

No sub-headers here`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertSectionTOCs([]byte(tt.content), tt.sectionTOCs))

			if got != tt.expected {
				t.Errorf("InsertSectionTOCs() =\n%s\n\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_FindFirstHeading(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "no heading",
			content:  "Just some text\nNo headers here",
			expected: -1,
		},
		{
			name:     "H1 first",
			content:  "# Title\nContent",
			expected: 0,
		},
		{
			name:     "H2 first",
			content:  "Some text\n## Section\nContent",
			expected: 1,
		},
		{
			name:     "heading in code block ignored",
			content:  "```\n# Not heading\n```\n## Real heading",
			expected: 3,
		},
		{
			name:     "fenced code with tilde",
			content:  "~~~\n# Not heading\n~~~\n## Real heading",
			expected: 3,
		},
		{
			name:     "H3 to H6 detection",
			content:  "Some text\n### H3\nContent",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindFirstHeading([]byte(tt.content))

			if got != tt.expected {
				t.Errorf("FindFirstHeading() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_InsertTOCAfterFirstHeading(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		toc      string
		contains []string
	}{
		{
			name:    "insert after H1",
			content: "# Title\n\nContent here",
			toc:     "- [Section](#section)",
			contains: []string{
				"# Title",
				"<!--TOC-->",
				"- [Section](#section)",
				"Content here",
			},
		},
		{
			name:    "insert at beginning when no heading",
			content: "Just some text\nNo headers",
			toc:     "- [Item](#item)",
			contains: []string{
				"<!--TOC-->",
				"- [Item](#item)",
				"Just some text",
			},
		},
		{
			name:    "insert after H2 when no H1",
			content: "Intro\n## Section\nContent",
			toc:     "- [Link](#link)",
			contains: []string{
				"## Section",
				"<!--TOC-->",
				"- [Link](#link)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertTOCAfterFirstHeading([]byte(tt.content), tt.toc))

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("InsertTOCAfterFirstHeading() should contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

func TestMarkerHandler_UpdateSectionTOCs(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		sectionTOCs []SectionTOC
		contains    []string
		excludes    []string
	}{
		{
			name: "update existing section TOCs",
			content: `# Chapter 1

<!--TOC-->

- [Old Link](#old-link)

<!--TOC-->

## Section 1.1

# Chapter 2

<!--TOC-->

- [Old Link 2](#old-link-2)

<!--TOC-->

## Section 2.1`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
				{H1Line: 6, TOC: "- [Section 2.1](#section-21)"},
			},
			contains: []string{
				"[Section 1.1](#section-11)",
				"[Section 2.1](#section-21)",
			},
			excludes: []string{
				"[Old Link]",
				"[Old Link 2]",
			},
		},
		{
			name:    "no existing markers - insert new",
			content: "# Chapter 1\n\n## Section 1.1",
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
			},
			contains: []string{
				"<!--TOC-->",
				"[Section 1.1](#section-11)",
			},
		},
		{
			name: "single unpaired marker - insert mode",
			content: `# Chapter 1

<!--TOC-->

## Section 1.1`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
			},
			contains: []string{
				"[Section 1.1](#section-11)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.UpdateSectionTOCs([]byte(tt.content), tt.sectionTOCs))

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("UpdateSectionTOCs() should contain %q, got:\n%s", s, got)
				}
			}

			for _, s := range tt.excludes {
				if strings.Contains(got, s) {
					t.Errorf("UpdateSectionTOCs() should NOT contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

func TestMarkerHandler_CustomMarker(t *testing.T) {
	customMarker := "<!-- TABLE OF CONTENTS -->"
	h := NewMarkerHandler(customMarker)

	content := "# Title\n<!-- TABLE OF CONTENTS -->\nContent"
	markers := h.FindMarkers([]byte(content))

	if !markers.Found {
		t.Error("Should find custom marker")
	}
	if markers.StartLine != 1 {
		t.Errorf("StartLine = %d, want 1", markers.StartLine)
	}
}

func TestMarkerHandler_EmptyMarker(t *testing.T) {
	// Empty marker should default to DefaultMarker
	h := NewMarkerHandler("")

	content := "# Title\n<!--TOC-->\nContent"
	markers := h.FindMarkers([]byte(content))

	if !markers.Found {
		t.Error("Should find default marker when empty string provided")
	}
}
