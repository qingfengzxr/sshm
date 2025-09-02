package ui

import (
	"strings"

	"sshm/internal/config"
	"sshm/internal/history"

	"github.com/charmbracelet/bubbles/table"
)

// calculateNameColumnWidth calculates the optimal width for the Name column
// based on the longest hostname, with a minimum of 8 and maximum of 40 characters
func calculateNameColumnWidth(hosts []config.SSHHost) int {
	maxLength := 8 // Minimum width to accommodate the "Name" header

	for _, host := range hosts {
		if len(host.Name) > maxLength {
			maxLength = len(host.Name)
		}
	}

	// Add some padding (2 characters) for better visual spacing
	maxLength += 2

	// Limit the maximum width to avoid extremely large columns
	if maxLength > 40 {
		maxLength = 40
	}

	return maxLength
}

// calculateTagsColumnWidth calculates the optimal width for the Tags column
// based on the longest tag string, with a minimum of 8 and maximum of 40 characters
func calculateTagsColumnWidth(hosts []config.SSHHost) int {
	maxLength := 8 // Minimum width to accommodate the "Tags" header

	for _, host := range hosts {
		// Format tags exactly as they appear in the table
		var tagsStr string
		if len(host.Tags) > 0 {
			// Add the # prefix to each tag and join them with spaces
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

	// Limit the maximum width to avoid extremely large columns
	if maxLength > 40 {
		maxLength = 40
	}

	return maxLength
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

	// Limit the maximum width to avoid extremely large columns
	if maxLength > 20 {
		maxLength = 20
	}

	return maxLength
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
			// Add the # prefix to each tag and join them with spaces
			var formattedTags []string
			for _, tag := range host.Tags {
				formattedTags = append(formattedTags, "#"+tag)
			}
			tagsStr = strings.Join(formattedTags, " ")
		}

		// Format last login information
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
