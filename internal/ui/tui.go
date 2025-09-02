package ui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"sshm/internal/config"
	"sshm/internal/history"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var searchStyleFocused = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("36")).
	Padding(0, 1)

var searchStyleUnfocused = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 1)

var headerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("36")).
	Bold(true).
	Align(lipgloss.Center)

const asciiTitle = `
 _____ _____ _   _ _____
|   __|   __|  |  |     |
|__   |__   |     | | | |
|_____|_____|__|__|_|_|_|
`

type SortMode int

const (
	SortByName SortMode = iota
	SortByLastUsed
)

func (s SortMode) String() string {
	switch s {
	case SortByName:
		return "Name (A-Z)"
	case SortByLastUsed:
		return "Last Login"
	default:
		return "Name (A-Z)"
	}
}

type Model struct {
	table          table.Model
	searchInput    textinput.Model
	hosts          []config.SSHHost
	filteredHosts  []config.SSHHost
	searchMode     bool
	deleteMode     bool
	deleteHost     string
	exitAction     string
	exitHostName   string
	historyManager *history.HistoryManager
	sortMode       SortMode
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
		case "tab":
			if !m.deleteMode {
				// Toggle focus between search input and table
				if m.searchMode {
					// Switch from search to table
					m.searchMode = false
					m.searchInput.Blur()
					m.table.Focus()
				} else {
					// Switch from table to search
					m.searchMode = true
					m.table.Blur()
					m.searchInput.Focus()
					return m, textinput.Blink
				}
				return m, nil
			}
		case "enter":
			if m.searchMode {
				// Validate search and return to table mode to allow commands
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

					// Record the connection in history
					if m.historyManager != nil {
						err := m.historyManager.RecordConnection(hostName)
						if err != nil {
							// Log error but don't prevent connection
							fmt.Printf("Warning: Could not record connection history: %v\n", err)
						}
					}

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
		case "s":
			if !m.searchMode && !m.deleteMode {
				// Cycle through sort modes (only 2 modes now)
				m.sortMode = (m.sortMode + 1) % 2
				// Re-apply current filter with new sort mode
				if m.searchInput.Value() != "" {
					m.filteredHosts = m.filterHosts(m.searchInput.Value())
				} else {
					m.filteredHosts = m.sortHosts(m.hosts)
				}
				m.updateTableRows()
				return m, nil
			}
		case "r":
			if !m.searchMode && !m.deleteMode {
				// Switch to sort by recent (last used)
				m.sortMode = SortByLastUsed
				// Re-apply current filter with new sort mode
				if m.searchInput.Value() != "" {
					m.filteredHosts = m.filterHosts(m.searchInput.Value())
				} else {
					m.filteredHosts = m.sortHosts(m.hosts)
				}
				m.updateTableRows()
				return m, nil
			}
		case "n":
			if !m.searchMode && !m.deleteMode {
				// Switch to sort by name
				m.sortMode = SortByName
				// Re-apply current filter with new sort mode
				if m.searchInput.Value() != "" {
					m.filteredHosts = m.filterHosts(m.searchInput.Value())
				} else {
					m.filteredHosts = m.sortHosts(m.hosts)
				}
				m.updateTableRows()
				return m, nil
			}
		}
	}

	// Update components based on mode
	if m.searchMode {
		oldValue := m.searchInput.Value()
		m.searchInput, cmd = m.searchInput.Update(msg)
		// Only update filtered hosts if search value changed
		if m.searchInput.Value() != oldValue {
			if m.searchInput.Value() != "" {
				m.filteredHosts = m.filterHosts(m.searchInput.Value())
			} else {
				m.filteredHosts = m.sortHosts(m.hosts)
			}
			m.updateTableRows()
		}
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

	// Add ASCII title
	view.WriteString(headerStyle.Render(asciiTitle) + "\n")

	// Add search bar (always visible) with appropriate style based on focus
	searchPrompt := "Search (/ to focus, Tab to switch): "
	if m.searchMode {
		view.WriteString(searchStyleFocused.Render(searchPrompt+m.searchInput.View()) + "\n")
	} else {
		view.WriteString(searchStyleUnfocused.Render(searchPrompt+m.searchInput.View()) + "\n")
	}

	// Add sort mode indicator
	sortInfo := fmt.Sprintf("Sort: %s", m.sortMode.String())
	view.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(sortInfo) + "\n\n")

	// Add table with appropriate style based on focus
	if m.searchMode {
		// Table is not focused, use gray border
		tableStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
		view.WriteString(tableStyle.Render(m.table.View()))
	} else {
		// Table is focused, use green border
		tableStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("36"))
		view.WriteString(tableStyle.Render(m.table.View()))
	}

	// Add help text
	if !m.searchMode {
		view.WriteString("\nUse ↑/↓ to navigate • Enter to connect • (a)dd • (e)dit • (d)elete • / to search • Tab to switch")
		view.WriteString("\nSort: (s)witch • (r)ecent • (n)ame • q/ESC to quit")
	} else {
		view.WriteString("\nType to filter hosts • Enter to validate search • Tab to switch to table • ESC to quit")
	}

	return view.String()
}

// sortHosts sorts hosts based on the current sort mode
func (m Model) sortHosts(hosts []config.SSHHost) []config.SSHHost {
	if m.historyManager == nil {
		return sortHostsByName(hosts)
	}

	switch m.sortMode {
	case SortByLastUsed:
		return m.historyManager.SortHostsByLastUsed(hosts)
	case SortByName:
		fallthrough
	default:
		return sortHostsByName(hosts)
	}
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
// based on the longest tags string, with a minimum of 8 and maximum of 40 characters
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
	if maxLength > 40 {
		maxLength = 40
	}

	return maxLength
}

// formatTimeAgo formats a time into a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	switch {
	case duration < time.Minute:
		seconds := int(duration.Seconds())
		if seconds <= 1 {
			return "1 second ago"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// calculateLastLoginColumnWidth calculates the optimal width for the Last Login column
// based on the longest time format, with a minimum of 12 and maximum of 20 characters
func calculateLastLoginColumnWidth(hosts []config.SSHHost, historyManager *history.HistoryManager) int {
	maxLength := 12 // Minimum width to accommodate the "Last Login" header

	if historyManager != nil {
		for _, host := range hosts {
			if lastConnect, exists := historyManager.GetLastConnectionTime(host.Name); exists {
				timeStr := formatTimeAgo(lastConnect)
				if len(timeStr) > maxLength {
					maxLength = len(timeStr)
				}
			}
		}
	}

	// Add some padding (2 characters) for better visual spacing
	maxLength += 2

	// Cap the maximum width to avoid extremely wide columns
	if maxLength > 20 {
		maxLength = 20
	}

	return maxLength
}

// calculateInfoColumnWidth calculates the optimal width for the combined Info column
// based on the longest combined tags and history string, with a minimum of 12 and maximum of 60 characters
// enterEditMode initializes edit mode for a specific host

// NewModel creates a new TUI model with the given SSH hosts
func NewModel(hosts []config.SSHHost) Model {
	// Initialize history manager
	historyManager, err := history.NewHistoryManager()
	if err != nil {
		// Log error but continue without history functionality
		fmt.Printf("Warning: Could not initialize history manager: %v\n", err)
		historyManager = nil
	}

	// Create the model with default sorting by name
	m := Model{
		hosts:          hosts,
		historyManager: historyManager,
		sortMode:       SortByName,
	}

	// Sort hosts based on default sort mode
	sortedHosts := m.sortHosts(hosts)

	// Create search input
	ti := textinput.New()
	ti.Placeholder = "Search hosts or tags..."
	ti.CharLimit = 50
	ti.Width = 50

	// Calculate optimal width for the Name column
	nameWidth := calculateNameColumnWidth(sortedHosts)

	// Calculate optimal width for the Tags column
	tagsWidth := calculateTagsColumnWidth(sortedHosts)

	// Calculate optimal width for the Last Login column
	lastLoginWidth := calculateLastLoginColumnWidth(sortedHosts, historyManager)

	// Create table columns
	columns := []table.Column{
		{Title: "Name", Width: nameWidth},
		{Title: "Hostname", Width: 25},
		{Title: "User", Width: 12},
		{Title: "Port", Width: 6},
		{Title: "Tags", Width: tagsWidth},
		{Title: "Last Login", Width: lastLoginWidth},
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

		// Format last login info
		var lastLoginStr string
		if historyManager != nil {
			if lastConnect, exists := historyManager.GetLastConnectionTime(host.Name); exists {
				lastLoginStr = formatTimeAgo(lastConnect)
			}
		}

		rows = append(rows, table.Row{
			host.Name,
			host.Hostname,
			host.User,
			host.Port,
			tagsStr,
			lastLoginStr,
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

	// Update the model with the table and other properties
	m.table = t
	m.searchInput = ti
	m.filteredHosts = sortedHosts

	return m
}

// RunInteractiveMode starts the interactive TUI
func RunInteractiveMode(hosts []config.SSHHost) error {
	for {
		m := NewModel(hosts)

		// Start the application in alt screen mode for clean exit
		p := tea.NewProgram(m, tea.WithAltScreen())
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
	var filtered []config.SSHHost

	if query == "" {
		filtered = m.hosts
	} else {
		query = strings.ToLower(query)

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
	}

	return m.sortHosts(filtered)
}

// updateTableRows updates the table with filtered hosts
func (m *Model) updateTableRows() {
	var rows []table.Row
	hostsToShow := m.filteredHosts
	if hostsToShow == nil {
		hostsToShow = m.hosts
	}

	for _, host := range hostsToShow {
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

		// Format last login info
		var lastLoginStr string
		if m.historyManager != nil {
			if lastConnect, exists := m.historyManager.GetLastConnectionTime(host.Name); exists {
				lastLoginStr = formatTimeAgo(lastConnect)
			}
		}

		rows = append(rows, table.Row{
			host.Name,
			host.Hostname,
			host.User,
			host.Port,
			tagsStr,
			lastLoginStr,
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
