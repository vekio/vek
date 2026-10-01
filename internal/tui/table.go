package tui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

// Table renders any set of columns and rows as a non-interactive Bubbles table.
type Table struct {
	model table.Model
}

// NewTable sizes each column to its widest value and requires rectangular rows.
func NewTable(headers []string, rows [][]string) (Table, error) {
	if len(headers) == 0 {
		return Table{}, fmt.Errorf("table requires at least one column")
	}
	columns := make([]table.Column, len(headers))
	for i, header := range headers {
		header = tableCell(header)
		columns[i] = table.Column{Title: header, Width: max(1, lipgloss.Width(header))}
	}

	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		if len(row) != len(columns) {
			return Table{}, fmt.Errorf("table row %d has %d cells; want %d", i, len(row), len(columns))
		}
		cells := make(table.Row, len(row))
		for j, value := range row {
			cells[j] = tableCell(value)
			columns[j].Width = max(columns[j].Width, lipgloss.Width(cells[j]))
		}
		tableRows[i] = cells
	}

	styles := table.DefaultStyles()
	styles.Header = lipgloss.NewStyle().Padding(0, 1)
	styles.Selected = lipgloss.NewStyle()
	width := 0
	for _, column := range columns {
		width += column.Width + 2 // One cell padding space on each side.
	}
	model := table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithStyles(styles),
		table.WithWidth(width),
		table.WithHeight(len(tableRows)+1),
	)
	return Table{model: model}, nil
}

func (t Table) View() string {
	return strings.TrimRight(t.model.View(), "\n")
}

func tableCell(value string) string {
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return strconv.Quote(value)
	}
	return value
}
