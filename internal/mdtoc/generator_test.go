package mdtoc

import (
	"strings"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name     string
		headers  []*Header
		opts     Options
		expected string
	}{
		{
			name:     "empty headers",
			headers:  []*Header{},
			opts:     DefaultOptions(),
			expected: "",
		},
		{
			name: "simple unordered list",
			headers: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
				{Level: 2, Text: "Section 2", AnchorLink: "section-2"},
			},
			opts: DefaultOptions(),
			expected: `- [Title](#title)
  - [Section 1](#section-1)
  - [Section 2](#section-2)`,
		},
		{
			name: "nested headers",
			headers: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
				{Level: 3, Text: "Subsection 1.1", AnchorLink: "subsection-11"},
				{Level: 2, Text: "Section 2", AnchorLink: "section-2"},
			},
			opts: DefaultOptions(),
			expected: `- [Title](#title)
  - [Section 1](#section-1)
    - [Subsection 1.1](#subsection-11)
  - [Section 2](#section-2)`,
		},
		{
			name: "ordered list",
			headers: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
				{Level: 2, Text: "Section 2", AnchorLink: "section-2"},
			},
			opts: Options{MinLevel: 1, MaxLevel: 3, Ordered: true},
			expected: `1. [Title](#title)
  1. [Section 1](#section-1)
  2. [Section 2](#section-2)`,
		},
		{
			name: "with line numbers",
			headers: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title", Line: 1, EndLine: 10},
				{Level: 2, Text: "Section 1", AnchorLink: "section-1", Line: 11, EndLine: 20},
			},
			opts: Options{MinLevel: 1, MaxLevel: 3, LineNumber: true},
			expected: "- [Title](#title) `:1-10`\n  - [Section 1](#section-1) `:11-20`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.opts)
			got := g.Generate(tt.headers)
			if got != tt.expected {
				t.Errorf("Generate() =\n%s\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestGenerator_GenerateSection(t *testing.T) {
	tests := []struct {
		name     string
		section  *Section
		opts     Options
		expected string
	}{
		{
			name:     "nil section",
			section:  nil,
			opts:     DefaultOptions(),
			expected: "",
		},
		{
			name: "section without sub-headers",
			section: &Section{
				Title:      &Header{Level: 1, Text: "Chapter 1", AnchorLink: "chapter-1"},
				SubHeaders: []*Header{},
			},
			opts:     DefaultOptions(),
			expected: "",
		},
		{
			name: "section with sub-headers",
			section: &Section{
				Title: &Header{Level: 1, Text: "Chapter 1", AnchorLink: "chapter-1"},
				SubHeaders: []*Header{
					{Level: 2, Text: "Section 1.1", AnchorLink: "section-11"},
					{Level: 2, Text: "Section 1.2", AnchorLink: "section-12"},
					{Level: 3, Text: "Subsection 1.2.1", AnchorLink: "subsection-121"},
				},
			},
			opts: Options{MinLevel: 2, MaxLevel: 3},
			expected: `- [Section 1.1](#section-11)
- [Section 1.2](#section-12)
  - [Subsection 1.2.1](#subsection-121)`,
		},
		{
			name: "section with level filtering",
			section: &Section{
				Title: &Header{Level: 1, Text: "Chapter 1", AnchorLink: "chapter-1"},
				SubHeaders: []*Header{
					{Level: 2, Text: "Section 1.1", AnchorLink: "section-11"},
					{Level: 3, Text: "Subsection 1.1.1", AnchorLink: "subsection-111"},
					{Level: 4, Text: "Deep 1.1.1.1", AnchorLink: "deep-1111"},
				},
			},
			opts: Options{MinLevel: 2, MaxLevel: 2}, // Only H2
			expected: `- [Section 1.1](#section-11)`,
		},
		{
			name: "section with ordered list",
			section: &Section{
				Title: &Header{Level: 1, Text: "Chapter 1", AnchorLink: "chapter-1"},
				SubHeaders: []*Header{
					{Level: 2, Text: "Section 1.1", AnchorLink: "section-11"},
					{Level: 2, Text: "Section 1.2", AnchorLink: "section-12"},
				},
			},
			opts: Options{MinLevel: 2, MaxLevel: 3, Ordered: true},
			expected: `1. [Section 1.1](#section-11)
2. [Section 1.2](#section-12)`,
		},
		{
			name: "section with line numbers",
			section: &Section{
				Title: &Header{Level: 1, Text: "Chapter 1", AnchorLink: "chapter-1", Line: 1, EndLine: 5},
				SubHeaders: []*Header{
					{Level: 2, Text: "Section 1.1", AnchorLink: "section-11", Line: 6, EndLine: 15},
					{Level: 2, Text: "Section 1.2", AnchorLink: "section-12", Line: 16, EndLine: 25},
				},
			},
			opts: Options{MinLevel: 2, MaxLevel: 3, LineNumber: true},
			expected: "- [Section 1.1](#section-11) `:6-15`\n- [Section 1.2](#section-12) `:16-25`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.opts)
			got := g.GenerateSection(tt.section)
			if got != tt.expected {
				t.Errorf("GenerateSection() =\n%s\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestGenerator_GenerateSection_RelativeIndent(t *testing.T) {
	// Test that section TOC uses relative indentation based on minimum level in section
	section := &Section{
		Title: &Header{Level: 1, Text: "Chapter", AnchorLink: "chapter"},
		SubHeaders: []*Header{
			{Level: 3, Text: "H3 First", AnchorLink: "h3-first"},
			{Level: 4, Text: "H4 Under", AnchorLink: "h4-under"},
			{Level: 3, Text: "H3 Second", AnchorLink: "h3-second"},
		},
	}

	g := NewGenerator(Options{MinLevel: 1, MaxLevel: 6})
	got := g.GenerateSection(section)

	// H3 should be at root level (no indent), H4 should be indented by 2
	expected := `- [H3 First](#h3-first)
  - [H4 Under](#h4-under)
- [H3 Second](#h3-second)`

	if got != expected {
		t.Errorf("GenerateSection() relative indent =\n%s\nwant:\n%s", got, expected)
	}
}

func TestGenerator_ChineseHeaders(t *testing.T) {
	headers := []*Header{
		{Level: 1, Text: "第一章", AnchorLink: "第一章"},
		{Level: 2, Text: "1.1 概述", AnchorLink: "11-概述"},
		{Level: 2, Text: "1.2 详细说明", AnchorLink: "12-详细说明"},
	}

	g := NewGenerator(DefaultOptions())
	got := g.Generate(headers)

	// Verify Chinese text is preserved
	if !strings.Contains(got, "第一章") {
		t.Error("Generate() should preserve Chinese text")
	}
	if !strings.Contains(got, "1.1 概述") {
		t.Error("Generate() should preserve Chinese text with numbers")
	}
}
