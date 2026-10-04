package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"termwerm/internal/analyzer"
	"termwerm/internal/report"
)

type focusArea int

const (
	focusTable focusArea = iota
	focusViewport
)

type Model struct {
	allMetrics   []analyzer.FunctionMetric
	filtered     []analyzer.FunctionMetric
	filterTier   int // 0: All, 1: Efficient, 2: Moderate, 3: Inefficient
	table        table.Model
	viewport     viewport.Model
	focus        focusArea
	width        int
	height       int
	statusMsg    string
	exportFile   string
	effCount     int
	modCount     int
	ineffCount   int
}

var (
	colorGreen  = lipgloss.Color("#10B981")
	colorAmber  = lipgloss.Color("#F59E0B")
	colorRed    = lipgloss.Color("#EF4444")
	colorDim    = lipgloss.Color("#6B7280")
	colorBorder = lipgloss.Color("#374151")
	colorActive = lipgloss.Color("#6366F1")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#4F46E5")).
			Padding(0, 1)

	badgeEff = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorGreen).
			Padding(0, 1)

	badgeMod = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorAmber).
			Padding(0, 1)

	badgeIneff = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorRed).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D1D5DB"))
	tipStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#93C5FD")).Italic(true)
	lineNumSty = lipgloss.NewStyle().Foreground(colorDim)
)

func NewModel(metrics []analyzer.FunctionMetric) Model {
	m := Model{
		allMetrics: metrics,
		filterTier: 0,
		exportFile: "termwerm-audit.md",
		focus:      focusTable,
	}

	for _, metric := range metrics {
		switch metric.Rating {
		case analyzer.Efficient:
			m.effCount++
		case analyzer.Moderate:
			m.modCount++
		case analyzer.Inefficient:
			m.ineffCount++
		}
	}

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "Rating", Width: 16},
			{Title: "Lang", Width: 10},
			{Title: "Function", Width: 18},
			{Title: "Location", Width: 26},
		}),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#3B82F6")).
		Bold(true)
	t.SetStyles(s)

	m.table = t
	m.viewport = viewport.New(0, 0)
	m.applyFilter()
	m.updateDetails()

	return m
}

func (m *Model) applyFilter() {
	var filtered []analyzer.FunctionMetric
	for _, item := range m.allMetrics {
		if m.filterTier == 0 || int(item.Rating) == m.filterTier {
			filtered = append(filtered, item)
		}
	}
	m.filtered = filtered

	var rows []table.Row
	for _, item := range filtered {
		loc := fmt.Sprintf("%s:%d", filepath.Base(item.File), item.Line)
		if len(loc) > 25 {
			loc = "..." + loc[len(loc)-22:]
		}
		rows = append(rows, table.Row{
			item.RatingTag,
			item.Language,
			item.Name,
			loc,
		})
	}
	m.table.SetRows(rows)
	if len(rows) > 0 {
		m.table.SetCursor(0)
	}
}

func (m *Model) updateDetails() {
	if len(m.filtered) == 0 {
		m.viewport.SetContent("No functions match the selected filter.")
		return
	}

	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filtered) {
		cursor = 0
	}
	metric := m.filtered[cursor]

	var badge string
	switch metric.Rating {
	case analyzer.Efficient:
		badge = badgeEff.Render(metric.RatingTag)
	case analyzer.Moderate:
		badge = badgeMod.Render(metric.RatingTag)
	default:
		badge = badgeIneff.Render(metric.RatingTag)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s   Complexity: %s\n\n", badge, metric.Complexity))
	sb.WriteString(fmt.Sprintf("%s %s (%s:%d-%d)\n", labelStyle.Render("Function:"), metric.Name, metric.File, metric.Line, metric.EndLine))
	sb.WriteString(fmt.Sprintf("%s   %s\n", labelStyle.Render("Reason:"), metric.Reason))
	sb.WriteString(fmt.Sprintf("%s      %s\n\n", labelStyle.Render("Tip:"), tipStyle.Render(metric.Tip)))
	sb.WriteString(labelStyle.Render("Code Snippet:") + "\n")
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	snipLines := strings.Split(metric.Snippet, "\n")
	for i, l := range snipLines {
		sb.WriteString(fmt.Sprintf("%s %s\n", lineNumSty.Render(fmt.Sprintf("%4d |", metric.Line+i)), l))
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoTop()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		leftWidth := m.width * 45 / 100
		rightWidth := m.width - leftWidth - 4
		headerHeight := 3
		footerHeight := 2
		mainHeight := m.height - headerHeight - footerHeight - 2
		if mainHeight < 5 {
			mainHeight = 5
		}

		m.table.SetWidth(leftWidth)
		m.table.SetHeight(mainHeight)

		colW := (leftWidth - 10) / 3
		if colW < 8 {
			colW = 8
		}
		m.table.SetColumns([]table.Column{
			{Title: "Rating", Width: 18},
			{Title: "Lang", Width: 8},
			{Title: "Function", Width: colW},
			{Title: "Location", Width: leftWidth - 28 - colW},
		})

		m.viewport.Width = rightWidth
		m.viewport.Height = mainHeight
		m.updateDetails()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			if m.focus == focusTable {
				m.focus = focusViewport
				m.table.Blur()
			} else {
				m.focus = focusTable
				m.table.Focus()
			}

		case "0", "1", "2", "3":
			tier := int(msg.String()[0] - '0')
			m.filterTier = tier
			m.applyFilter()
			m.updateDetails()
			m.statusMsg = fmt.Sprintf("Filter set to: %s", filterName(tier))

		case "m":
			err := report.ExportMarkdown(m.exportFile, m.allMetrics)
			if err != nil {
				m.statusMsg = fmt.Sprintf("Error saving report: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Saved to %s", m.exportFile)
			}

		default:
			if m.focus == focusViewport {
				m.viewport, cmd = m.viewport.Update(msg)
			} else {
				oldCursor := m.table.Cursor()
				m.table, cmd = m.table.Update(msg)
				if m.table.Cursor() != oldCursor {
					m.updateDetails()
				}
			}
		}
	}

	return m, cmd
}

func filterName(tier int) string {
	switch tier {
	case 1:
		return "[▲ Efficient]"
	case 2:
		return "[● Moderate]"
	case 3:
		return "[▼ Not Efficient]"
	default:
		return "All"
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading termwerm..."
	}

	total := len(m.allMetrics)
	filterStr := filterName(m.filterTier)
	header := fmt.Sprintf(" %s  Total: %d | %s: %d | %s: %d | %s: %d   Filter: [%s]",
		titleStyle.Render("termwerm"),
		total,
		lipgloss.NewStyle().Foreground(colorGreen).Render("▲ Efficient"), m.effCount,
		lipgloss.NewStyle().Foreground(colorAmber).Render("● Moderate"), m.modCount,
		lipgloss.NewStyle().Foreground(colorRed).Render("▼ Not Efficient"), m.ineffCount,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#38BDF8")).Render(filterStr),
	)

	tableBorder := colorBorder
	viewportBorder := colorBorder
	if m.focus == focusTable {
		tableBorder = colorActive
	} else {
		viewportBorder = colorActive
	}

	leftBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(tableBorder).
		Width(m.table.Width()).
		Height(m.table.Height()).
		Render(m.table.View())

	rightBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(viewportBorder).
		Width(m.viewport.Width).
		Height(m.viewport.Height).
		Padding(0, 1).
		Render(m.viewport.View())

	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox)

	status := m.statusMsg
	if status == "" {
		status = "j/k or ↑/↓: Select  |  1/2/3: Filter (0: All)  |  tab: Focus Snippet  |  m: Export MD  |  q: Quit"
	}
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Background(lipgloss.Color("#1F2937")).
		Padding(0, 1).
		Width(m.width).
		Render(status)

	return lipgloss.JoinVertical(lipgloss.Left, header, mainRow, footer)
}
