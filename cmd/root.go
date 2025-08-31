package cmd

import (
	"fmt"
	"log"
	"os"

	"sshm/internal/config"
	"sshm/internal/ui"

	"github.com/spf13/cobra"
)

// version will be set at build time via -ldflags
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "sshm",
	Short: "SSH Manager - A modern SSH connection manager",
	Long: `SSH Manager (sshm) is a modern command-line tool for managing SSH connections.
It provides an interactive interface to browse and connect to your SSH hosts
configured in your ~/.ssh/config file.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, run interactive mode
		if len(args) == 0 {
			runInteractiveMode()
			return
		}

		// If a host name is provided, connect directly
		hostName := args[0]
		connectToHost(hostName)
	},
}

func runInteractiveMode() {
	// Parse SSH configurations
	hosts, err := config.ParseSSHConfig()
	if err != nil {
		log.Fatalf("Error reading SSH config file: %v", err)
	}

	if len(hosts) == 0 {
		fmt.Println("No SSH hosts found in your ~/.ssh/config file.")
		fmt.Print("Would you like to add a new host now? [y/N]: ")
		var response string
		_, err := fmt.Scanln(&response)
		if err == nil && (response == "y" || response == "Y") {
			err := ui.RunAddForm("")
			if err != nil {
				fmt.Printf("Error adding host: %v\n", err)
			}
			// After adding, try to reload hosts and continue if any exist
			hosts, err = config.ParseSSHConfig()
			if err != nil || len(hosts) == 0 {
				fmt.Println("No hosts available, exiting.")
				os.Exit(1)
			}
		} else {
			fmt.Println("No hosts available, exiting.")
			os.Exit(1)
		}
	}

	// Run the interactive TUI
	if err := ui.RunInteractiveMode(hosts); err != nil {
		log.Fatalf("Error running interactive mode: %v", err)
	}
}

func connectToHost(hostName string) {
	// Parse SSH configurations to verify host exists
	hosts, err := config.ParseSSHConfig()
	if err != nil {
		log.Fatalf("Error reading SSH config file: %v", err)
	}

	// Check if host exists
	var hostFound bool
	for _, host := range hosts {
		if host.Name == hostName {
			hostFound = true
			break
		}
	}

	if !hostFound {
		fmt.Printf("Error: Host '%s' not found in SSH configuration.\n", hostName)
		fmt.Println("Use 'sshm' to see available hosts.")
		os.Exit(1)
	}

	// Connect to the host
	fmt.Printf("Connecting to %s...\n", hostName)
	// Note: In a real implementation, you'd use exec.Command here
	// For now, just print the command that would be executed
	fmt.Printf("ssh %s\n", hostName)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
