package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inputModel struct {
	textInput textinput.Model
	err       error
	title     string
	validator func(string) error
}

func initialInputModel(title, placeholder string, password bool, validator func(string) error) inputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	
	if password {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	}

	ti.Prompt = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Render("▸ ")

	return inputModel{
		textInput: ti,
		err:       nil,
		title:     title,
		validator: validator,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			value := m.textInput.Value()
			if m.validator != nil {
				if err := m.validator(value); err != nil {
					m.err = err
					return m, nil
				}
			}
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.textInput.SetValue("")
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		MarginBottom(1)

	s := titleStyle.Render(m.title) + "\n"
	s += m.textInput.View() + "\n"
	
	if m.err != nil {
		errStyle := lipgloss.NewStyle().
			Foreground(ErrorColor).
			MarginTop(1)
		s += errStyle.Render("✗ " + m.err.Error())
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		MarginTop(1)
	s += helpStyle.Render("\n(esc to cancel)")

	return lipgloss.NewStyle().
		Margin(1, 0).
		Render(s)
}

// GetInput prompts for user input with optional validation
func GetInput(title, placeholder string, password bool, validator func(string) error) (string, error) {
	p := tea.NewProgram(initialInputModel(title, placeholder, password, validator))
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	if model, ok := m.(inputModel); ok {
		value := model.textInput.Value()
		if value == "" {
			return "", fmt.Errorf("cancelled")
		}
		return value, nil
	}

	return "", fmt.Errorf("unexpected model type")
}

// GetConfirmation asks for user confirmation
func GetConfirmation(message string) bool {
	confirmStyle := lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)
	
	fmt.Println(confirmStyle.Render(message))
	
	response, err := GetInput(
		"",
		"Type 'yes' to confirm or press ESC to cancel",
		false,
		func(s string) error {
			s = strings.ToLower(strings.TrimSpace(s))
			if s != "yes" && s != "y" && s != "no" && s != "n" && s != "" {
				return fmt.Errorf("please type 'yes' or 'no'")
			}
			return nil
		},
	)
	
	if err != nil {
		return false
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "yes" || response == "y"
}

// SelectFromList allows selecting from a list of options
func SelectFromList(title string, options []string) (int, error) {
	fmt.Println(SubtitleStyle.Render(title))
	
	for i, option := range options {
		fmt.Printf("%s %s\n", 
			Primary.Sprint(fmt.Sprintf("%d.", i+1)),
			option,
		)
	}
	
	input, err := GetInput(
		"Select an option",
		fmt.Sprintf("Enter a number (1-%d)", len(options)),
		false,
		func(s string) error {
			var choice int
			if _, err := fmt.Sscanf(s, "%d", &choice); err != nil {
				return fmt.Errorf("please enter a valid number")
			}
			if choice < 1 || choice > len(options) {
				return fmt.Errorf("please enter a number between 1 and %d", len(options))
			}
			return nil
		},
	)
	
	if err != nil {
		return -1, err
	}
	
	var choice int
	fmt.Sscanf(input, "%d", &choice)
	return choice - 1, nil
} 