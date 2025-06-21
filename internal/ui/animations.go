package ui

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
)

// StartSpinner creates and starts an animated spinner
func StartSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
	s.Prefix = " "
	s.Suffix = fmt.Sprintf(" %s", InfoStyle.Render(message))
	s.Color("magenta", "bold")
	s.Start()
	return s
}

// ShowSuccess displays a success message with animation
func ShowSuccess(message string) {
	pterm.Success.WithShowLineNumber(false).Println(message)
}

// ShowError displays an error message with animation
func ShowError(message string) {
	pterm.Error.WithShowLineNumber(false).Println(message)
}

// ShowWarning displays a warning message with animation
func ShowWarning(message string) {
	pterm.Warning.WithShowLineNumber(false).Println(message)
}

// ShowInfo displays an info message with animation
func ShowInfo(message string) {
	pterm.Info.WithShowLineNumber(false).Println(message)
}

// ShowBanner displays an animated banner
func ShowBanner() {
	banner := pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("Dev", pterm.NewStyle(pterm.FgMagenta)),
		pterm.NewLettersFromStringWithStyle("Tools", pterm.NewStyle(pterm.FgCyan)),
	)
	banner.Render()
	
	// Add a subtitle
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Italic(true).
		MarginLeft(2).
		Render("Your development toolkit manager")
	
	fmt.Println(subtitle)
	fmt.Println()
}

// ProgressBar represents an animated progress bar
type ProgressBar struct {
	bar *pterm.ProgressbarPrinter
}

// NewProgressBar creates a new progress bar
func NewProgressBar(title string, total int) *ProgressBar {
	bar, _ := pterm.DefaultProgressbar.
		WithTotal(total).
		WithTitle(title).
		WithShowCount(false).
		WithShowPercentage(true).
		WithShowElapsedTime(false).
		WithBarCharacter("█").
		WithLastCharacter("▓").
		WithElapsedTimeRoundingFactor(time.Second).
		WithBarStyle(pterm.NewStyle(pterm.FgMagenta)).
		Start()
	
	return &ProgressBar{bar: bar}
}

// Increment increments the progress bar
func (p *ProgressBar) Increment() {
	p.bar.Increment()
}

// UpdateTitle updates the progress bar title
func (p *ProgressBar) UpdateTitle(title string) {
	p.bar.UpdateTitle(title)
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() {
	p.bar.Stop()
}

// AnimatedText displays text with a typewriter effect
func AnimatedText(text string, delay time.Duration) {
	for _, char := range text {
		fmt.Print(string(char))
		time.Sleep(delay)
	}
	fmt.Println()
}

// ShowLoadingAnimation displays a loading animation for a given duration
func ShowLoadingAnimation(message string, work func() error) error {
	spinner := StartSpinner(message)
	err := work()
	spinner.Stop()
	
	if err != nil {
		ShowError(fmt.Sprintf("Failed: %v", err))
	} else {
		ShowSuccess("Complete!")
	}
	
	return err
}

// CreateBox creates a styled box around content
func CreateBox(title, content string) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Width(60)
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)
	
	fullContent := titleStyle.Render(title) + "\n" + content
	return box.Render(fullContent)
}

// ShowAnimatedList displays a list with animation
func ShowAnimatedList(title string, items []string) {
	fmt.Println()
	fmt.Println(SubtitleStyle.Render(title))
	
	for i, item := range items {
		time.Sleep(50 * time.Millisecond)
		bullet := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Render(fmt.Sprintf("  %d.", i+1))
		fmt.Printf("%s %s\n", bullet, item)
	}
	fmt.Println()
} 