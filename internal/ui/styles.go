package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

var (
	// Colors
	PrimaryColor   = lipgloss.Color("#7D56F4")
	SecondaryColor = lipgloss.Color("#F97316")
	SuccessColor   = lipgloss.Color("#10B981")
	ErrorColor     = lipgloss.Color("#EF4444")
	WarningColor   = lipgloss.Color("#F59E0B")
	InfoColor      = lipgloss.Color("#3B82F6")
	
	// Styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Background(lipgloss.Color("#1A1A2E")).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)
	
	SubtitleStyle = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true).
		MarginBottom(1)
	
	MenuItemStyle = lipgloss.NewStyle().
		PaddingLeft(2)
	
	SelectedMenuItemStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Background(lipgloss.Color("#2D2D44")).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1)
	
	DescriptionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		PaddingLeft(4)
	
	SuccessStyle = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)
	
	ErrorStyle = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)
	
	WarningStyle = lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)
	
	InfoStyle = lipgloss.NewStyle().
		Foreground(InfoColor)
	
	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)
	
	// Terminal colors for simpler outputs
	Success = color.New(color.FgGreen, color.Bold)
	Error   = color.New(color.FgRed, color.Bold)
	Warning = color.New(color.FgYellow, color.Bold)
	Info    = color.New(color.FgCyan)
	Primary = color.New(color.FgMagenta, color.Bold)
)

// GetGradientTitle creates a gradient title
func GetGradientTitle(text string) string {
	gradient := []lipgloss.Color{
		"#7D56F4",
		"#8B5CF6",
		"#A78BFA",
		"#C4B5FD",
		"#A78BFA",
		"#8B5CF6",
		"#7D56F4",
	}
	
	style := lipgloss.NewStyle()
	result := ""
	
	for i, char := range text {
		colorIndex := i % len(gradient)
		result += style.Foreground(gradient[colorIndex]).Render(string(char))
	}
	
	return result
} 