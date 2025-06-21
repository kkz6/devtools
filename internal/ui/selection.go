package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Simple item for selection lists
type simpleItem struct {
	title string
	index int
}

func (i simpleItem) Title() string       { return i.title }
func (i simpleItem) Description() string { return "" }
func (i simpleItem) FilterValue() string { return i.title }

// Simple delegate for compact lists
type simpleDelegate struct{}

func (d simpleDelegate) Height() int                             { return 1 }
func (d simpleDelegate) Spacing() int                            { return 0 }
func (d simpleDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d simpleDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(simpleItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := lipgloss.NewStyle().PaddingLeft(2).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(PrimaryColor).
				Bold(true).
				Render("▸ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type selectionModel struct {
	list     list.Model
	choice   int
	quitting bool
	title    string
}

func (m selectionModel) Init() tea.Cmd {
	return nil
}

func (m selectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "esc":
			m.choice = -1
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(simpleItem)
			if ok {
				m.choice = i.index
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectionModel) View() string {
	if m.quitting {
		return ""
	}
	
	// Add title
	titleStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		MarginBottom(1)
	
	view := titleStyle.Render(m.title) + "\n"
	view += m.list.View()
	
	// Add help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		MarginTop(1)
	view += "\n" + helpStyle.Render("↑/↓ to navigate • Enter to select • Esc to cancel")
	
	return lipgloss.NewStyle().Margin(1, 0).Render(view)
}

// SelectFromListInteractive allows selecting from a list using arrow keys
func SelectFromListInteractive(title string, options []string) (int, error) {
	// Create list items
	items := []list.Item{}
	for i, option := range options {
		items = append(items, simpleItem{
			title: option,
			index: i,
		})
	}

	// Calculate appropriate height
	listHeight := len(items) + 2 // Add space for borders
	if listHeight > 15 {
		listHeight = 15 // Cap at 15 lines
	}

	// Create the list
	l := list.New(items, simpleDelegate{}, 60, listHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	
	// Simple key bindings
	l.KeyMap = list.KeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
		),
	}

	m := selectionModel{
		list:  l,
		title: title,
		choice: -1,
	}

	// Run the program
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return -1, fmt.Errorf("error running selection: %w", err)
	}

	// Get the choice
	if m, ok := finalModel.(selectionModel); ok {
		if m.choice == -1 {
			return -1, fmt.Errorf("cancelled")
		}
		return m.choice, nil
	}

	return -1, fmt.Errorf("unexpected model type")
} 