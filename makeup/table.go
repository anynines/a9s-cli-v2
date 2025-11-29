package makeup

import (
	"fmt"
	"strings"

	"github.com/anynines/a9s-cli-v2/cost/klutchaws"
	"github.com/charmbracelet/lipgloss"
)

// RenderCostTable renders a BOM-like table using lipgloss styling.
func RenderCostTable(report klutchaws.Report) string {
	headers := []string{"Resource", "Category", "Qty", "Unit", "Unit Price", "Hourly", "Monthly", "Notes"}
	rows := make([][]string, 0, len(report.Items)+2)
	for _, item := range report.Items {
		rows = append(rows, []string{
			item.Name,
			item.Category,
			fmt.Sprintf("%.2f", item.Quantity),
			item.Unit,
			fmt.Sprintf("$%.4f", item.UnitPrice),
			fmt.Sprintf("$%.2f", item.Hourly),
			fmt.Sprintf("$%.2f", item.Monthly),
			item.Notes,
		})
	}
	rows = append(rows, []string{"", "", "", "", "", "", "", ""})
	rows = append(rows, []string{
		"TOTAL",
		"",
		"",
		"",
		"",
		fmt.Sprintf("$%.2f", report.TotalHourly),
		fmt.Sprintf("$%.2f", report.TotalMonthly),
		fmt.Sprintf("Region: %s", report.Region),
	})

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row {
			if len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	cell := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e4833e"))
	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D9DCCF"))
	divider := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, false, true).
		BorderForeground(lipgloss.Color("#505d7a"))

	var lines []string
	lines = append(lines, renderRow(headers, widths, cell, headerStyle))
	lines = append(lines, divider.Render(strings.Repeat("─", totalWidth(widths)+len(widths)*2)))
	for idx, row := range rows {
		style := rowStyle
		if idx == len(rows)-1 {
			style = headerStyle
		}
		lines = append(lines, renderRow(row, widths, cell, style))
	}

	table := strings.Join(lines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#505d7a")).
		Margin(1, 0, 0, 0).
		Render(table)
	return box
}

func renderRow(cols []string, widths []int, base, style lipgloss.Style) string {
	rendered := make([]string, len(cols))
	for i, col := range cols {
		rendered[i] = style.Render(base.Width(widths[i]).Render(col))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func totalWidth(widths []int) int {
	sum := 0
	for _, w := range widths {
		sum += w
	}
	return sum
}
