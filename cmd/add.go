package cmd

import (
	"fmt"
	"sshm/internal/ui"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [hostname]",
	Short: "Add a new SSH host configuration",
	Long:  `Add a new SSH host configuration with an interactive form.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var hostname string
		if len(args) > 0 {
			hostname = args[0]
		}

		err := ui.RunAddForm(hostname)
		if err != nil {
			fmt.Printf("Error adding host: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
