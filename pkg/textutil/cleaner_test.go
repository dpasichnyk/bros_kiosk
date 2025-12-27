package textutil

import (
	"testing"
)

func TestCleanSummary(t *testing.T) {
	tests := []struct {
		name    string
		summary string
		title   string
		want    string
	}{
		{
			name:    "No Overlap",
			summary: "This is a summary.",
			title:   "Different Title",
			want:    "This is a summary.",
		},
		{
			name:    "Exact Overlap",
			summary: "Breaking News: Something happened. More details here.",
			title:   "Breaking News: Something happened",
			want:    "More details here.",
		},
		{
			name:    "Fuzzy Overlap",
			summary: "Breaking News - Something happened. More details.",
			title:   "Breaking News: Something happened", // Colon vs Dash difference
			want:    "More details.",
		},
		{
			name:    "HTML Strip",
			summary: "<p><b>Bold</b> summary</p>",
			title:   "Title",
			want:    "Bold summary",
		},
		{
			name:    "Title Equals Summary",
			summary: "Just the title",
			title:   "Just the title",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanSummary(tt.summary, tt.title); got != tt.want {
				t.Errorf("CleanSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}
