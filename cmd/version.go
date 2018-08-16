package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show sb  client version",
	Long:    `Show sb  client version.`,
	Example: ` sb version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(rootCmd.Use + " v" + VERSION)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}