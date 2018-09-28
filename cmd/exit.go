package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func init() {
	exitCmd := &cobra.Command{
		Use:     "exit",
		Short:   "Exit this tool",
		Long:    `Exist this tool`,
		Example: ` exit`,
		Aliases: []string{"exit", "quit"},
		Run: func(cmd *cobra.Command, args []string) {
			os.Exit(0)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	exitCmd.Annotations = map[string]string{"command": "exit"}

	rootCmd.AddCommand(exitCmd)
}
