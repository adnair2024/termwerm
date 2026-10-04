package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"termwerm/internal/analyzer"
	"termwerm/internal/report"
	"termwerm/internal/ui"
)

func main() {
	exportFlag := flag.Bool("export", false, "Write audit report to termwerm-audit.md and exit without opening TUI")
	flag.Parse()

	targetDir := "."
	args := flag.Args()
	if len(args) > 0 {
		targetDir = args[0]
	}

	info, err := os.Stat(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: directory %q not found: %v\n", targetDir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %q is not a directory\n", targetDir)
		os.Exit(1)
	}

	metrics, err := analyzer.ScanDirectory(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if *exportFlag {
		outFile := "termwerm-audit.md"
		if err := report.ExportMarkdown(outFile, metrics); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting report: %v\n", err)
			os.Exit(1)
		}
		eff, mod, ineff := 0, 0, 0
		for _, m := range metrics {
			switch m.Rating {
			case analyzer.Efficient:
				eff++
			case analyzer.Moderate:
				mod++
			case analyzer.Inefficient:
				ineff++
			}
		}
		fmt.Printf("termwerm scan complete for %q\n", targetDir)
		fmt.Printf("Total Functions: %d\n", len(metrics))
		fmt.Printf("  [▲ Efficient]:     %d\n", eff)
		fmt.Printf("  [● Moderate]:      %d\n", mod)
		fmt.Printf("  [▼ Not Efficient]: %d\n", ineff)
		fmt.Printf("Report successfully saved to %s\n", outFile)
		return
	}

	p := tea.NewProgram(ui.NewModel(metrics), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running termwerm: %v\n", err)
		os.Exit(1)
	}
}
