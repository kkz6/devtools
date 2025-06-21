package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kkz6/devtools/internal/types"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(PrimaryColor)
	
	titleStyle = lipgloss.NewStyle().
		Background(PrimaryColor).
		Foreground(lipgloss.Color("#FAFAFA")).
		Bold(true).
		Padding(0, 1)
)

type item struct {
	title       string
	description string
	id          string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("▸ " + strings.Join(s, " "))
		}
		str = fmt.Sprintf("%d. %s", index+1, i.title)
	}

	fmt.Fprint(w, fn(str))
	if i.description != "" {
		desc := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			PaddingLeft(4).
			Render(i.description)
		fmt.Fprint(w, "\n"+desc)
	}
}

type menuModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		// Set height to use most of the terminal, leaving room for margins
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.id
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menuModel) View() string {
	if m.choice != "" {
		return ""
	}
	if m.quitting {
		return ""
	}
	return docStyle.Render(m.list.View())
}

// ShowAnimatedMenu displays an animated menu and returns the selected module ID
func ShowAnimatedMenu(modules []types.ModuleInfo) (string, error) {
	// Create list items
	items := []list.Item{}
	for _, module := range modules {
		items = append(items, item{
			title:       module.Name,
			description: module.Description,
			id:          module.ID,
		})
	}
	
	// Add exit option
	items = append(items, item{
		title:       "Exit",
		description: "Quit the application",
		id:          "exit",
	})

	// Create the list
	const defaultWidth = 80
	// Set a larger default height to show more items
	const listHeight = 30

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = GetGradientTitle("🚀 DevTools Manager by Karthick")
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.
		Foreground(lipgloss.Color("#9CA3AF"))
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.
		Foreground(lipgloss.Color("#9CA3AF"))

	// Custom key bindings
	l.KeyMap = list.KeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown"),
			key.WithHelp("→/l/pgdn", "next page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup"),
			key.WithHelp("←/h/pgup", "prev page"),
		),
	}

	m := menuModel{list: l}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running menu: %w", err)
	}

	// Get the choice
	if m, ok := finalModel.(menuModel); ok {
		if m.choice == "exit" {
			return "", fmt.Errorf("user exited")
		}
		return m.choice, nil
	}

	return "", fmt.Errorf("unexpected model type")
} 