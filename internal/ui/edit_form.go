package ui

import (
	"sshm/internal/config"
	"sshm/internal/validation"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyleEdit = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	fieldStyleEdit = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyleEdit = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	helpStyleEdit = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type editFormModel struct {
	inputs       []textinput.Model
	focused      int
	err          string
	success      bool
	originalName string
}

func RunEditForm(hostName string) error {
	// Get the existing host configuration
	host, err := config.GetSSHHost(hostName)
	if err != nil {
		return err
	}

	inputs := make([]textinput.Model, 8)

	// Name input
	inputs[nameInput] = textinput.New()
	inputs[nameInput].Placeholder = "server-name"
	inputs[nameInput].Focus()
	inputs[nameInput].CharLimit = 50
	inputs[nameInput].Width = 30
	inputs[nameInput].SetValue(host.Name)

	// Hostname input
	inputs[hostnameInput] = textinput.New()
	inputs[hostnameInput].Placeholder = "192.168.1.100 or example.com"
	inputs[hostnameInput].CharLimit = 100
	inputs[hostnameInput].Width = 30
	inputs[hostnameInput].SetValue(host.Hostname)

	// User input
	inputs[userInput] = textinput.New()
	inputs[userInput].Placeholder = "root"
	inputs[userInput].CharLimit = 50
	inputs[userInput].Width = 30
	inputs[userInput].SetValue(host.User)

	// Port input
	inputs[portInput] = textinput.New()
	inputs[portInput].Placeholder = "22"
	inputs[portInput].CharLimit = 5
	inputs[portInput].Width = 30
	inputs[portInput].SetValue(host.Port)

	// Identity input
	inputs[identityInput] = textinput.New()
	inputs[identityInput].Placeholder = "~/.ssh/id_rsa"
	inputs[identityInput].CharLimit = 200
	inputs[identityInput].Width = 50
	inputs[identityInput].SetValue(host.Identity)

	// ProxyJump input
	inputs[proxyJumpInput] = textinput.New()
	inputs[proxyJumpInput].Placeholder = "user@jump-host:port or existing-host-name"
	inputs[proxyJumpInput].CharLimit = 200
	inputs[proxyJumpInput].Width = 50
	inputs[proxyJumpInput].SetValue(host.ProxyJump)

	// SSH Options input
	inputs[optionsInput] = textinput.New()
	inputs[optionsInput].Placeholder = "-o Compression=yes -o ServerAliveInterval=60"
	inputs[optionsInput].CharLimit = 500
	inputs[optionsInput].Width = 70
	inputs[optionsInput].SetValue(config.FormatSSHOptionsForCommand(host.Options))

	// Tags input
	inputs[tagsInput] = textinput.New()
	inputs[tagsInput].Placeholder = "production, web, database"
	inputs[tagsInput].CharLimit = 200
	inputs[tagsInput].Width = 50
	if len(host.Tags) > 0 {
		inputs[tagsInput].SetValue(strings.Join(host.Tags, ", "))
	}

	m := editFormModel{
		inputs:       inputs,
		focused:      nameInput,
		originalName: hostName,
	}

	// Open in separate window like add form
	p := tea.NewProgram(&m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (m *editFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *editFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+enter":
			// Allow submission from any field with Ctrl+Enter
			return m, m.submitEditForm()

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Handle form submission
			if s == "enter" && m.focused == len(m.inputs)-1 {
				return m, m.submitEditForm()
			}

			// Cycle inputs
			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs)-1 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}

			for i := range m.inputs {
				if i == m.focused {
					cmds = append(cmds, m.inputs[i].Focus())
					continue
				}
				m.inputs[i].Blur()
			}

			return m, tea.Batch(cmds...)
		}

	case editResult:
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.success = true
			m.err = ""
			return m, tea.Quit
		}
	}

	// Update inputs
	cmd := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmd[i] = m.inputs[i].Update(msg)
	}
	cmds = append(cmds, cmd...)

	return m, tea.Batch(cmds...)
}

func (m *editFormModel) View() string {
	if m.success {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyleEdit.Render("Edit SSH Host Configuration"))
	b.WriteString("\n\n")

	fields := []string{
		"Host Name *",
		"Hostname/IP *",
		"User",
		"Port",
		"Identity File",
		"ProxyJump",
		"SSH Options",
		"Tags (comma-separated)",
	}

	for i, field := range fields {
		b.WriteString(fieldStyleEdit.Render(field))
		b.WriteString("\n")
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n\n")
	}

	if m.err != "" {
		b.WriteString(errorStyleEdit.Render("Error: " + m.err))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyleEdit.Render("Tab/Shift+Tab: navigate • Enter on last field: submit • Ctrl+Enter: submit • Ctrl+C/Esc: cancel"))
	b.WriteString("\n")
	b.WriteString(helpStyleEdit.Render("* Required fields"))

	return b.String()
}

type editResult struct {
	hostname string
	err      error
}

func (m *editFormModel) submitEditForm() tea.Cmd {
	return func() tea.Msg {
		// Get values
		name := strings.TrimSpace(m.inputs[nameInput].Value())
		hostname := strings.TrimSpace(m.inputs[hostnameInput].Value())
		user := strings.TrimSpace(m.inputs[userInput].Value())
		port := strings.TrimSpace(m.inputs[portInput].Value())
		identity := strings.TrimSpace(m.inputs[identityInput].Value())
		proxyJump := strings.TrimSpace(m.inputs[proxyJumpInput].Value())
		options := strings.TrimSpace(m.inputs[optionsInput].Value())

		// Set defaults
		if port == "" {
			port = "22"
		}
		// Do not auto-fill identity with placeholder if left empty; keep it empty so it's optional

		// Validate all fields
		if err := validation.ValidateHost(name, hostname, port, identity); err != nil {
			return editResult{err: err}
		}

		// Parse tags
		tagsStr := strings.TrimSpace(m.inputs[tagsInput].Value())
		var tags []string
		if tagsStr != "" {
			for _, tag := range strings.Split(tagsStr, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}

		// Create updated host configuration
		host := config.SSHHost{
			Name:      name,
			Hostname:  hostname,
			User:      user,
			Port:      port,
			Identity:  identity,
			ProxyJump: proxyJump,
			Options:   config.ParseSSHOptionsFromCommand(options),
			Tags:      tags,
		}

		// Update the configuration
		err := config.UpdateSSHHost(m.originalName, host)
		return editResult{hostname: name, err: err}
	}
}
