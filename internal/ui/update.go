package ui

import (
	"fmt"
	"os/exec"

	"sshm/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		// Ajoute ici d'autres tea.Cmd si tu veux charger des données, démarrer un spinner, etc.
	)
}

// Update handles model updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle different message types
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update terminal size and recalculate styles
		m.width = msg.Width
		m.height = msg.Height
		m.styles = NewStyles(m.width)
		m.ready = true

		// Update table height and columns based on new window size
		m.updateTableHeight()
		m.updateTableColumns()

		// Update sub-forms if they exist
		if m.addForm != nil {
			m.addForm.width = m.width
			m.addForm.height = m.height
			m.addForm.styles = m.styles
		}
		if m.editForm != nil {
			m.editForm.width = m.width
			m.editForm.height = m.height
			m.editForm.styles = m.styles
		}
		if m.portForwardForm != nil {
			m.portForwardForm.width = m.width
			m.portForwardForm.height = m.height
			m.portForwardForm.styles = m.styles
		}
		return m, nil

	case addFormSubmitMsg:
		if msg.err != nil {
			// Show error in form
			if m.addForm != nil {
				m.addForm.err = msg.err.Error()
			}
			return m, nil
		} else {
			// Success: refresh hosts and return to list view
			var hosts []config.SSHHost
			var err error

			if m.configFile != "" {
				hosts, err = config.ParseSSHConfigFile(m.configFile)
			} else {
				hosts, err = config.ParseSSHConfig()
			}

			if err != nil {
				return m, tea.Quit
			}
			m.hosts = m.sortHosts(hosts)

			// Reapply search filter if there is one active
			if m.searchInput.Value() != "" {
				m.filteredHosts = m.filterHosts(m.searchInput.Value())
			} else {
				m.filteredHosts = m.hosts
			}

			m.updateTableRows()
			m.viewMode = ViewList
			m.addForm = nil
			m.table.Focus()
			return m, nil
		}

	case addFormCancelMsg:
		// Cancel: return to list view
		m.viewMode = ViewList
		m.addForm = nil
		m.table.Focus()
		return m, nil

	case editFormSubmitMsg:
		if msg.err != nil {
			// Show error in form
			if m.editForm != nil {
				m.editForm.err = msg.err.Error()
			}
			return m, nil
		} else {
			// Success: refresh hosts and return to list view
			var hosts []config.SSHHost
			var err error

			if m.configFile != "" {
				hosts, err = config.ParseSSHConfigFile(m.configFile)
			} else {
				hosts, err = config.ParseSSHConfig()
			}

			if err != nil {
				return m, tea.Quit
			}
			m.hosts = m.sortHosts(hosts)

			// Reapply search filter if there is one active
			if m.searchInput.Value() != "" {
				m.filteredHosts = m.filterHosts(m.searchInput.Value())
			} else {
				m.filteredHosts = m.hosts
			}

			m.updateTableRows()
			m.viewMode = ViewList
			m.editForm = nil
			m.table.Focus()
			return m, nil
		}

	case editFormCancelMsg:
		// Cancel: return to list view
		m.viewMode = ViewList
		m.editForm = nil
		m.table.Focus()
		return m, nil

	case portForwardSubmitMsg:
		if msg.err != nil {
			// Show error in form
			if m.portForwardForm != nil {
				m.portForwardForm.err = msg.err.Error()
			}
			return m, nil
		} else {
			// Success: execute SSH command with port forwarding
			if len(msg.sshArgs) > 0 {
				sshCmd := exec.Command("ssh", msg.sshArgs...)

				// Record the connection in history
				if m.historyManager != nil && m.portForwardForm != nil {
					err := m.historyManager.RecordConnection(m.portForwardForm.hostName)
					if err != nil {
						fmt.Printf("Warning: Could not record connection history: %v\n", err)
					}
				}

				return m, tea.ExecProcess(sshCmd, func(err error) tea.Msg {
					return tea.Quit()
				})
			}

			// If no SSH args, just return to list view
			m.viewMode = ViewList
			m.portForwardForm = nil
			m.table.Focus()
			return m, nil
		}

	case portForwardCancelMsg:
		// Cancel: return to list view
		m.viewMode = ViewList
		m.portForwardForm = nil
		m.table.Focus()
		return m, nil

	case tea.KeyMsg:
		// Handle view-specific key presses
		switch m.viewMode {
		case ViewAdd:
			if m.addForm != nil {
				var newForm *addFormModel
				newForm, cmd = m.addForm.Update(msg)
				m.addForm = newForm
				return m, cmd
			}
		case ViewEdit:
			if m.editForm != nil {
				var newForm *editFormModel
				newForm, cmd = m.editForm.Update(msg)
				m.editForm = newForm
				return m, cmd
			}
		case ViewPortForward:
			if m.portForwardForm != nil {
				var newForm *portForwardModel
				newForm, cmd = m.portForwardForm.Update(msg)
				m.portForwardForm = newForm
				return m, cmd
			}
		case ViewList:
			// Handle list view keys
			return m.handleListViewKeys(msg)
		}
	}

	return m, cmd
}

func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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
			m.updateTableStyles()
			m.table.Blur()
			m.searchInput.Focus()
			return m, textinput.Blink
		}
	case "tab":
		if !m.deleteMode {
			// Switch focus between search input and table
			if m.searchMode {
				// Switch from search to table
				m.searchMode = false
				m.updateTableStyles()
				m.searchInput.Blur()
				m.table.Focus()
			} else {
				// Switch from table to search
				m.searchMode = true
				m.updateTableStyles()
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
			m.updateTableStyles()
			m.searchInput.Blur()
			m.table.Focus()
			return m, nil
		} else if m.deleteMode {
			// Confirm deletion
			var err error
			if m.configFile != "" {
				err = config.DeleteSSHHostFromFile(m.deleteHost, m.configFile)
			} else {
				err = config.DeleteSSHHost(m.deleteHost)
			}
			if err != nil {
				// Could display an error message here
				m.deleteMode = false
				m.deleteHost = ""
				m.table.Focus()
				return m, nil
			}
			// Refresh the hosts list
			var hosts []config.SSHHost
			var parseErr error

			if m.configFile != "" {
				hosts, parseErr = config.ParseSSHConfigFile(m.configFile)
			} else {
				hosts, parseErr = config.ParseSSHConfig()
			}

			if parseErr != nil {
				// Could display an error message here
				m.deleteMode = false
				m.deleteHost = ""
				m.table.Focus()
				return m, nil
			}
			m.hosts = m.sortHosts(hosts)

			// Reapply search filter if there is one active
			if m.searchInput.Value() != "" {
				m.filteredHosts = m.filterHosts(m.searchInput.Value())
			} else {
				m.filteredHosts = m.hosts
			}

			m.updateTableRows()
			m.deleteMode = false
			m.deleteHost = ""
			m.table.Focus()
			return m, nil
		} else {
			// Connect to the selected host
			selected := m.table.SelectedRow()
			if len(selected) > 0 {
				hostName := selected[0] // The hostname is in the first column

				// Record the connection in history
				if m.historyManager != nil {
					err := m.historyManager.RecordConnection(hostName)
					if err != nil {
						// Log the error but don't prevent the connection
						fmt.Printf("Warning: Could not record connection history: %v\n", err)
					}
				}

				// Build the SSH command with the appropriate config file
				var sshCmd *exec.Cmd
				if m.configFile != "" {
					sshCmd = exec.Command("ssh", "-F", m.configFile, hostName)
				} else {
					sshCmd = exec.Command("ssh", hostName)
				}

				return m, tea.ExecProcess(sshCmd, func(err error) tea.Msg {
					return tea.Quit()
				})
			}
		}
	case "e":
		if !m.searchMode && !m.deleteMode {
			// Edit the selected host
			selected := m.table.SelectedRow()
			if len(selected) > 0 {
				hostName := selected[0] // The hostname is in the first column
				editForm, err := NewEditForm(hostName, m.styles, m.width, m.height, m.configFile)
				if err != nil {
					// Handle error - could show in UI
					return m, nil
				}
				m.editForm = editForm
				m.viewMode = ViewEdit
				return m, textinput.Blink
			}
		}
	case "a":
		if !m.searchMode && !m.deleteMode {
			// Add a new host
			m.addForm = NewAddForm("", m.styles, m.width, m.height, m.configFile)
			m.viewMode = ViewAdd
			return m, textinput.Blink
		}
	case "d":
		if !m.searchMode && !m.deleteMode {
			// Delete the selected host
			selected := m.table.SelectedRow()
			if len(selected) > 0 {
				hostName := selected[0] // The hostname is in the first column
				m.deleteMode = true
				m.deleteHost = hostName
				m.table.Blur()
				return m, nil
			}
		}
	case "f":
		if !m.searchMode && !m.deleteMode {
			// Port forwarding for the selected host
			selected := m.table.SelectedRow()
			if len(selected) > 0 {
				hostName := selected[0] // The hostname is in the first column
				m.portForwardForm = NewPortForwardForm(hostName, m.styles, m.width, m.height, m.configFile)
				m.viewMode = ViewPortForward
				return m, textinput.Blink
			}
		}
	case "s":
		if !m.searchMode && !m.deleteMode {
			// Cycle through sort modes (only 2 modes now)
			m.sortMode = (m.sortMode + 1) % 2
			// Re-apply the current filter with the new sort mode
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
			// Re-apply the current filter with the new sort mode
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
			// Re-apply the current filter with the new sort mode
			if m.searchInput.Value() != "" {
				m.filteredHosts = m.filterHosts(m.searchInput.Value())
			} else {
				m.filteredHosts = m.sortHosts(m.hosts)
			}
			m.updateTableRows()
			return m, nil
		}
	}

	// Update the appropriate component based on mode
	if m.searchMode {
		oldValue := m.searchInput.Value()
		m.searchInput, cmd = m.searchInput.Update(msg)
		// Update filtered hosts only if the search value has changed
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
