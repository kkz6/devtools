package cursorreport

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table represents a formatted table
type Table struct {
	title   string
	headers []string
	rows    [][]string
	widths  []int
}

// NewTable creates a new table
func NewTable(title string) *Table {
	return &Table{
		title:   title,
		headers: []string{},
		rows:    [][]string{},
		widths:  []int{},
	}
}

// AddHeader adds headers to the table
func (t *Table) AddHeader(headers ...string) {
	t.headers = headers
	t.widths = make([]int, len(headers))
	
	// Initialize widths with header lengths
	for i, h := range headers {
		t.widths[i] = len(h)
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells ...string) {
	if len(cells) != len(t.headers) {
		return // Skip invalid rows
	}
	
	t.rows = append(t.rows, cells)
	
	// Update column widths
	for i, cell := range cells {
		if len(cell) > t.widths[i] {
			t.widths[i] = len(cell)
		}
	}
}

// Render renders the table as a string
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}
	
	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)
		
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F97316"))
		
	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4"))
		
	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB"))
	
	// Calculate total width
	totalWidth := 0
	for _, w := range t.widths {
		totalWidth += w + 3 // padding
	}
	totalWidth += len(t.widths) - 1 // separators
	
	var result strings.Builder
	
	// Title
	if t.title != "" {
		result.WriteString(titleStyle.Render(t.title))
		result.WriteString("\n")
	}
	
	// Top border
	result.WriteString(borderStyle.Render("┌"))
	for i, w := range t.widths {
		result.WriteString(borderStyle.Render(strings.Repeat("─", w+2)))
		if i < len(t.widths)-1 {
			result.WriteString(borderStyle.Render("┬"))
		}
	}
	result.WriteString(borderStyle.Render("┐"))
	result.WriteString("\n")
	
	// Headers
	result.WriteString(borderStyle.Render("│"))
	for i, h := range t.headers {
		padded := fmt.Sprintf(" %-*s ", t.widths[i], h)
		result.WriteString(headerStyle.Render(padded))
		if i < len(t.headers)-1 {
			result.WriteString(borderStyle.Render("│"))
		}
	}
	result.WriteString(borderStyle.Render("│"))
	result.WriteString("\n")
	
	// Header separator
	result.WriteString(borderStyle.Render("├"))
	for i, w := range t.widths {
		result.WriteString(borderStyle.Render(strings.Repeat("─", w+2)))
		if i < len(t.widths)-1 {
			result.WriteString(borderStyle.Render("┼"))
		}
	}
	result.WriteString(borderStyle.Render("┤"))
	result.WriteString("\n")
	
	// Rows
	for _, row := range t.rows {
		result.WriteString(borderStyle.Render("│"))
		for i, cell := range row {
			padded := fmt.Sprintf(" %-*s ", t.widths[i], cell)
			result.WriteString(cellStyle.Render(padded))
			if i < len(row)-1 {
				result.WriteString(borderStyle.Render("│"))
			}
		}
		result.WriteString(borderStyle.Render("│"))
		result.WriteString("\n")
	}
	
	// Bottom border
	result.WriteString(borderStyle.Render("└"))
	for i, w := range t.widths {
		result.WriteString(borderStyle.Render(strings.Repeat("─", w+2)))
		if i < len(t.widths)-1 {
			result.WriteString(borderStyle.Render("┴"))
		}
	}
	result.WriteString(borderStyle.Render("┘"))
	
	return result.String()
} 