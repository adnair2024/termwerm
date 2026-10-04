package report

import (
	"fmt"
	"os"
	"strings"
	"time"

	"termwerm/internal/analyzer"
)

func GenerateMarkdown(metrics []analyzer.FunctionMetric) string {
	var effCount, modCount, ineffCount int
	var ineffList, modList []analyzer.FunctionMetric

	for _, m := range metrics {
		switch m.Rating {
		case analyzer.Efficient:
			effCount++
		case analyzer.Moderate:
			modCount++
			modList = append(modList, m)
		case analyzer.Inefficient:
			ineffCount++
			ineffList = append(ineffList, m)
		}
	}

	var sb strings.Builder
	sb.WriteString("# termwerm Code Efficiency Audit\n\n")
	sb.WriteString(fmt.Sprintf("**Scan Date:** %s  \n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Summary:** [▲ Efficient]: %d | [● Moderate]: %d | [▼ Not Efficient]: %d  \n\n", effCount, modCount, ineffCount))
	sb.WriteString("---\n\n")

	if len(ineffList) > 0 {
		sb.WriteString("## ▼ Not Efficient\n\n")
		for _, m := range ineffList {
			writeMetricBlock(&sb, m)
		}
	}

	if len(modList) > 0 {
		sb.WriteString("## ● Moderate\n\n")
		for _, m := range modList {
			writeMetricBlock(&sb, m)
		}
	}

	if len(ineffList) == 0 && len(modList) == 0 {
		sb.WriteString("All analyzed functions are [▲ Efficient]! No performance bottlenecks detected.\n")
	}

	return sb.String()
}

func writeMetricBlock(sb *strings.Builder, m analyzer.FunctionMetric) {
	sb.WriteString(fmt.Sprintf("### `%s`\n", m.Name))
	sb.WriteString(fmt.Sprintf("- **Location:** `%s:%d`\n", m.File, m.Line))
	sb.WriteString(fmt.Sprintf("- **Language:** %s\n", m.Language))
	sb.WriteString(fmt.Sprintf("- **Complexity:** %s\n", m.Complexity))
	sb.WriteString(fmt.Sprintf("- **Issue:** %s\n", m.Reason))
	sb.WriteString(fmt.Sprintf("- **Tip:** %s\n\n", m.Tip))

	langTag := strings.ToLower(m.Language)
	if langTag == "c++" {
		langTag = "cpp"
	} else if langTag == "c#" {
		langTag = "csharp"
	}
	sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n---\n\n", langTag, m.Snippet))
}

func ExportMarkdown(targetPath string, metrics []analyzer.FunctionMetric) error {
	content := GenerateMarkdown(metrics)
	return os.WriteFile(targetPath, []byte(content), 0644)
}
