package ui

import (
	"fmt"
	"strings"

	"sshm/internal/config"
	"sshm/internal/history"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewModel creates a new TUI model with the given SSH hosts
func NewModel(hosts []config.SSHHost, configFile string) Model {
	// Initialize the history manager
	historyManager, err := history.NewHistoryManager()
	if err != nil {
		// Log the error but continue without the history functionality
		fmt.Printf("Warning: Could not initialize history manager: %v\n", err)
		historyManager = nil
	}

	// Create initial styles (will be updated on first WindowSizeMsg)
	styles := NewStyles(80) // Default width

	// Create the model with default sorting by name
	m := Model{
		hosts:          hosts,
		historyManager: historyManager,
		sortMode:       SortByName,
		configFile:     configFile,
		styles:         styles,
		width:          80,
		height:         24,
		ready:          false,
		viewMode:       ViewList,
	}

	// Sort hosts according to the default sort mode
	sortedHosts := m.sortHosts(hosts)

	// Create the search input
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
			// Add the # prefix to each tag and join them with spaces
			var formattedTags []string
			for _, tag := range host.Tags {
				formattedTags = append(formattedTags, "#"+tag)
			}
			tagsStr = strings.Join(formattedTags, " ")
		}

		// Format last login information
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

	// Determine table height: 1 (header) + number of hosts (max 10)
	hostCount := len(rows)
	tableHeight := 1 // header
	if hostCount < 10 {
		tableHeight += hostCount
	} else {
		tableHeight += 10
	}

	// Create the table
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
		BorderForeground(lipgloss.Color(SecondaryColor)).
		BorderBottom(true).
		Bold(false)
	s.Selected = m.styles.Selected

	t.SetStyles(s)

	// Update the model with the table and other properties
	m.table = t
	m.searchInput = ti
	m.filteredHosts = sortedHosts

	// Initialize table styles based on initial focus state
	m.updateTableStyles()

	return m
}

// RunInteractiveMode starts the interactive TUI interface
func RunInteractiveMode(hosts []config.SSHHost, configFile string) error {
	m := NewModel(hosts, configFile)

	// Start the application in alt screen mode for clean output
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
