package mdtoc

import (
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
