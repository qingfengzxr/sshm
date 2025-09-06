package ui

import (
	"sort"
	"strings"

	"github.com/Gu1llaum-3/sshm/internal/config"
)

// sortHosts sorts hosts according to the current sort mode
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

// filterHosts filters hosts according to the search query (name or tags)
func (m Model) filterHosts(query string) []config.SSHHost {
	var filtered []config.SSHHost

	if query == "" {
		filtered = m.hosts
	} else {
		query = strings.ToLower(query)

		for _, host := range m.hosts {
			// Check the hostname
			if strings.Contains(strings.ToLower(host.Name), query) {
				filtered = append(filtered, host)
				continue
			}

			// Check the hostname
			if strings.Contains(strings.ToLower(host.Hostname), query) {
				filtered = append(filtered, host)
				continue
			}

			// Check the tags
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
