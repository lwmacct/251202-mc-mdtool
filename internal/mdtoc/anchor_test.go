package mdtoc

import (
	"testing"
)

func TestAnchorGenerator_Generate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "chinese text",
			input:    "中文标题",
			expected: "中文标题",
		},
		{
			name:     "mixed text",
			input:    "Hello, World! 你好",
			expected: "hello-world-你好",
		},
		{
			name:     "with numbers",
			input:    "Section 1.1",
			expected: "section-11",
		},
		{
			name:     "with code",
			input:    "Using `fmt.Println`",
			expected: "using-fmtprintln",
		},
		{
			name:     "with bold",
			input:    "This is **bold** text",
			expected: "this-is-bold-text",
		},
		{
			name:     "with link",
			input:    "Check [this link](http://example.com)",
			expected: "check-this-link",
		},
		{
			name:     "special characters",
			input:    "C++ Programming",
			expected: "c-programming",
		},
		{
			name:     "multiple spaces",
			input:    "Hello    World",
			expected: "hello-world",
		},
		{
			name:     "leading trailing spaces",
			input:    "  Hello World  ",
			expected: "hello-world",
		},
		{
			name:     "with underscores",
			input:    "m_250428_dmi_chassis_height",
			expected: "m_250428_dmi_chassis_height",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewAnchorGenerator()
			got := g.Generate(tt.input)
			if got != tt.expected {
				t.Errorf("Generate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAnchorGenerator_DuplicateHandling(t *testing.T) {
	g := NewAnchorGenerator()

	// 第一个标题
	if got := g.Generate("Title"); got != "title" {
		t.Errorf("First Title = %q, want %q", got, "title")
	}

	// 第二个相同标题
	if got := g.Generate("Title"); got != "title-1" {
		t.Errorf("Second Title = %q, want %q", got, "title-1")
	}

	// 第三个相同标题
	if got := g.Generate("Title"); got != "title-2" {
		t.Errorf("Third Title = %q, want %q", got, "title-2")
	}

	// 不同标题
	if got := g.Generate("Other"); got != "other" {
		t.Errorf("Other = %q, want %q", got, "other")
	}
}

func TestAnchorGenerator_Reset(t *testing.T) {
	g := NewAnchorGenerator()

	g.Generate("Title")
	g.Generate("Title")

	// 重置后应该重新开始
	g.Reset()

	if got := g.Generate("Title"); got != "title" {
		t.Errorf("After reset, Title = %q, want %q", got, "title")
	}
}
