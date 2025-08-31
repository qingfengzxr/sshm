package ui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"sshm/internal/config"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var searchStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("36")).
	Padding(0, 1)

type Model struct {
	table         table.Model
	searchInput   textinput.Model
	hosts         []config.SSHHost
	filteredHosts []config.SSHHost
	searchMode    bool
	deleteMode    bool
	deleteHost    string
	exitAction    string
	exitHostName  string
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			if m.searchMode {
				// Exit search mode
				m.searchMode = false
				m.searchInput.Blur()
				m.table.Focus()
				return m, nil
			}
			if m.deleteMode {
				// Exit delete mode
				m.deleteMode = false
				m.deleteHost = ""
				m.table.Focus()
				return m, nil
			}
			return m, tea.Quit
		case "q":
			if !m.searchMode && !m.deleteMode {
				return m, tea.Quit
			}
		case "/", "ctrl+f":
			if !m.searchMode && !m.deleteMode {
				// Enter search mode
				m.searchMode = true
				m.table.Blur()
				m.searchInput.Focus()
				return m, textinput.Blink
			}
		case "enter":
			if m.searchMode {
				// Exit search mode and focus table
				m.searchMode = false
				m.searchInput.Blur()
				m.table.Focus()
				return m, nil
			} else if m.deleteMode {
				// Confirm deletion
				err := config.DeleteSSHHost(m.deleteHost)
				if err != nil {
					// Could show error message here
					m.deleteMode = false
					m.deleteHost = ""
					m.table.Focus()
					return m, nil
				}
				// Refresh the host list
				hosts, err := config.ParseSSHConfig()
				if err != nil {
					// Could show error message here
					m.deleteMode = false
					m.deleteHost = ""
					m.table.Focus()
					return m, nil
				}
				m.hosts = sortHostsByName(hosts)
				m.filteredHosts = m.hosts
				m.updateTableRows()
				m.deleteMode = false
				m.deleteHost = ""
				m.table.Focus()
				return m, nil
			} else {
				// Connect to selected host
				selected := m.table.SelectedRow()
				if len(selected) > 0 {
					hostName := selected[0] // Host name is in the first column
					return m, tea.ExecProcess(exec.Command("ssh", hostName), func(err error) tea.Msg {
						return tea.Quit()
					})
				}
			}
		case "e":
			if !m.searchMode && !m.deleteMode {
				// Edit selected host using dedicated edit form
				selected := m.table.SelectedRow()
				if len(selected) > 0 {
					hostName := selected[0] // Host name is in the first column
					// Store the edit action and exit
					m.exitAction = "edit"
					m.exitHostName = hostName
					return m, tea.Quit
				}
			}
		case "a":
			if !m.searchMode && !m.deleteMode {
				// Add new host using dedicated add form
				m.exitAction = "add"
				return m, tea.Quit
			}
		case "d":
			if !m.searchMode && !m.deleteMode {
				// Delete selected host
				selected := m.table.SelectedRow()
				if len(selected) > 0 {
					hostName := selected[0] // Host name is in the first column
					m.deleteMode = true
					m.deleteHost = hostName
					m.table.Blur()
					return m, nil
				}
			}
		}
	}

	// Update components based on mode
	if m.searchMode {
		m.searchInput, cmd = m.searchInput.Update(msg)
		// Filter hosts when search input changes
		if m.searchInput.Value() != "" {
			m.filteredHosts = m.filterHosts(m.searchInput.Value())
		} else {
			m.filteredHosts = m.hosts
		}
		m.updateTableRows()
	} else {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.deleteMode {
		return m.renderDeleteConfirmation()
	}

	var view strings.Builder

	// Add search bar
	searchPrompt := "Search (/ to search, ESC to exit search): "
	if m.searchMode {
		view.WriteString(searchStyle.Render(searchPrompt+m.searchInput.View()) + "\n\n")
	}

	// Add table
	view.WriteString(baseStyle.Render(m.table.View()))

	// Add help text
	if !m.searchMode {
		view.WriteString("\nUse ↑/↓ to navigate • Enter to connect • (a)dd • (e)dit • (d)elete • / to search • (q)uit")
	} else {
		view.WriteString("\nType to filter hosts by name or tag • Enter to select • ESC to exit search")
	}

	return view.String()
}

// sortHostsByName sorts a slice of SSH hosts alphabetically by name
func sortHostsByName(hosts []config.SSHHost) []config.SSHHost {
	sorted := make([]config.SSHHost, len(hosts))
	copy(sorted, hosts)

	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	return sorted
}

// calculateNameColumnWidth calculates the optimal width for the Name column
// based on the longest host name, with a minimum of 8 and maximum of 40 characters
func calculateNameColumnWidth(hosts []config.SSHHost) int {
	maxLength := 8 // Minimum width to accommodate the "Name" header

	for _, host := range hosts {
		if len(host.Name) > maxLength {
			maxLength = len(host.Name)
		}
	}

	// Add some padding (2 characters) for better visual spacing
	maxLength += 2

	// Cap the maximum width to avoid extremely wide columns
	if maxLength > 40 {
		maxLength = 40
	}

	return maxLength
}

// calculateTagsColumnWidth calculates the optimal width for the Tags column
// based on the longest tags string, with a minimum of 8 and maximum of 50 characters
func calculateTagsColumnWidth(hosts []config.SSHHost) int {
	maxLength := 8 // Minimum width to accommodate the "Tags" header

	for _, host := range hosts {
		// Format tags exactly the same way they appear in the table
		var tagsStr string
		if len(host.Tags) > 0 {
			// Add # prefix to each tag and join with spaces
			var formattedTags []string
			for _, tag := range host.Tags {
				formattedTags = append(formattedTags, "#"+tag)
			}
			tagsStr = strings.Join(formattedTags, " ")
		}

		if len(tagsStr) > maxLength {
			maxLength = len(tagsStr)
		}
	}

	// Add some padding (2 characters) for better visual spacing
	maxLength += 2

	// Cap the maximum width to avoid extremely wide columns
	if maxLength > 50 {
		maxLength = 50
	}

	return maxLength
}

// NewModel creates a new TUI model with the given SSH hosts
func NewModel(hosts []config.SSHHost) Model {
	// Sort hosts alphabetically by name
	sortedHosts := sortHostsByName(hosts)

	// Create search input
	ti := textinput.New()
	ti.Placeholder = "Search hosts or tags..."
	ti.CharLimit = 50
	ti.Width = 50

	// Calculate optimal width for the Name column
	nameWidth := calculateNameColumnWidth(sortedHosts)

	// Calculate optimal width for the Tags column
	tagsWidth := calculateTagsColumnWidth(sortedHosts)

	// Create table columns
	columns := []table.Column{
		{Title: "Name", Width: nameWidth},
		{Title: "Hostname", Width: 25},
		{Title: "User", Width: 12},
		{Title: "Port", Width: 6},
		{Title: "Tags", Width: tagsWidth},
	}

	// Convert hosts to table rows
	var rows []table.Row
	for _, host := range sortedHosts {
		// Format tags for display
		var tagsStr string
		if len(host.Tags) > 0 {
			// Add # prefix to each tag and join with spaces
			var formattedTags []string
			for _, tag := range host.Tags {
				formattedTags = append(formattedTags, "#"+tag)
			}
			tagsStr = strings.Join(formattedTags, " ")
		}

		rows = append(rows, table.Row{
			host.Name,
			host.Hostname,
			host.User,
			host.Port,
			tagsStr,
		})
	}

	// Déterminer la hauteur du tableau : 1 (header) + nombre de hosts (max 10)
	hostCount := len(rows)
	tableHeight := 1 // header
	if hostCount < 10 {
		tableHeight += hostCount
	} else {
		tableHeight += 10
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return Model{
		table:         t,
		searchInput:   ti,
		hosts:         sortedHosts,
		filteredHosts: sortedHosts,
		searchMode:    false,
	}
}

// RunInteractiveMode starts the interactive TUI
func RunInteractiveMode(hosts []config.SSHHost) error {
	for {
		m := NewModel(hosts)

		// Start the application in terminal (without alt screen)
		p := tea.NewProgram(m)
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}

		// Check if the final model indicates an action
		if model, ok := finalModel.(Model); ok {
			if model.exitAction == "edit" && model.exitHostName != "" {
				// Launch the dedicated edit form (opens in separate window)
				if err := RunEditForm(model.exitHostName); err != nil {
					fmt.Printf("Error editing host: %v\n", err)
					// Continue the loop to return to the main interface
					continue
				}

				// Clear screen before returning to TUI
				fmt.Print("\033[2J\033[H")

				// Refresh the hosts list after editing
				refreshedHosts, err := config.ParseSSHConfig()
				if err != nil {
					return fmt.Errorf("error refreshing hosts after edit: %w", err)
				}
				hosts = refreshedHosts

				// Continue the loop to return to the main interface
				continue
			} else if model.exitAction == "add" {
				// Launch the dedicated add form (opens in separate window)
				if err := RunAddForm(""); err != nil {
					fmt.Printf("Error adding host: %v\n", err)
					// Continue the loop to return to the main interface
					continue
				}

				// Clear screen before returning to TUI
				fmt.Print("\033[2J\033[H")

				// Refresh the hosts list after adding
				refreshedHosts, err := config.ParseSSHConfig()
				if err != nil {
					return fmt.Errorf("error refreshing hosts after add: %w", err)
				}
				hosts = refreshedHosts

				// Continue the loop to return to the main interface
				continue
			}
		}

		// If no special command, exit normally
		break
	}

	return nil
}

// filterHosts filters hosts based on search query (name or tags)
func (m Model) filterHosts(query string) []config.SSHHost {
	if query == "" {
		return sortHostsByName(m.hosts)
	}

	query = strings.ToLower(query)
	var filtered []config.SSHHost

	for _, host := range m.hosts {
		// Check host name
		if strings.Contains(strings.ToLower(host.Name), query) {
			filtered = append(filtered, host)
			continue
		}

		// Check hostname
		if strings.Contains(strings.ToLower(host.Hostname), query) {
			filtered = append(filtered, host)
			continue
		}

		// Check tags
		for _, tag := range host.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				filtered = append(filtered, host)
				break
			}
		}
	}

	return sortHostsByName(filtered)
}

// updateTableRows updates the table with filtered hosts
func (m *Model) updateTableRows() {
	var rows []table.Row
	hostsToShow := m.filteredHosts
	if hostsToShow == nil {
		hostsToShow = m.hosts
	}

	// Sort hosts alphabetically by name
	sortedHosts := sortHostsByName(hostsToShow)

	for _, host := range sortedHosts {
		// Format tags for display
		var tagsStr string
		if len(host.Tags) > 0 {
			// Add # prefix to each tag and join with spaces
			var formattedTags []string
			for _, tag := range host.Tags {
				formattedTags = append(formattedTags, "#"+tag)
			}
			tagsStr = strings.Join(formattedTags, " ")
		}

		rows = append(rows, table.Row{
			host.Name,
			host.Hostname,
			host.User,
			host.Port,
			tagsStr,
		})
	}

	m.table.SetRows(rows)
}

// enterEditMode initializes edit mode for a specific host
// renderDeleteConfirmation renders the delete confirmation dialog
func (m Model) renderDeleteConfirmation() string {
	var view strings.Builder

	view.WriteString(lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")). // Red border
		Padding(1, 2).
		Render(fmt.Sprintf("⚠️  Delete SSH Host\n\nAre you sure you want to delete host '%s'?\n\nThis action cannot be undone.\n\nPress Enter to confirm or Esc to cancel", m.deleteHost)))

	return view.String()
}
