package report

import (
	"strings"
	"testing"

	"termwerm/internal/analyzer"
)

func TestGenerateMarkdown(t *testing.T) {
	metrics := []analyzer.FunctionMetric{
		{
			Rating:     analyzer.Efficient,
			RatingTag:  "[▲ Efficient]",
			Complexity: "O(1) / O(n)",
			Language:   "Go",
			Name:       "simpleFunc",
			File:       "main.go",
			Line:       10,
			EndLine:    15,
			Reason:     "Fast linear or constant execution.",
			Tip:        "No nested loops found. Great efficiency!",
			Snippet:    "func simpleFunc() {\n\tprintln(\"hi\")\n}",
		},
		{
			Rating:     analyzer.Moderate,
			RatingTag:  "[● Moderate]",
			Complexity: "O(n²)",
			Language:   "Go",
			Name:       "nestedFunc",
			File:       "main.go",
			Line:       20,
			EndLine:    30,
			MaxDepth:   2,
			Reason:     "Found 2 nested loops.",
			Tip:        "Consider using a hash map.",
			Snippet:    "func nestedFunc() {\n\tfor ...\n}",
		},
	}

	md := GenerateMarkdown(metrics)
	if !strings.Contains(md, "**Summary:** [▲ Efficient]: 1 | [● Moderate]: 1 | [▼ Not Efficient]: 0") {
		t.Errorf("Unexpected summary in markdown: %s", md)
	}
	if !strings.Contains(md, "## ● Moderate") {
		t.Errorf("Expected Moderate section: %s", md)
	}
	if strings.Contains(md, "## ▼ Not Efficient") {
		t.Errorf("Did not expect Inefficient section: %s", md)
	}
}
