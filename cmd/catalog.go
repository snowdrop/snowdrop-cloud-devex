package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"github.com/snowdrop/k8s-supervisor/pkg/catalog"
	log "github.com/sirupsen/logrus"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog [options]",
	Short: "List, select or bind a service from a catalog.",
	Long:  `List, select or bind a service from a catalog.`,
	Example: fmt.Sprintf("%s\n%s\n%s",
		catalogListCmd.Example,
		catalogSelectCmd.Example,
		catalogBindCmd.Example),
}

var catalogListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all available services from the catalog",
	Long:    "List all available services from the Service Catalog's broker.",
	Example: ` sb catalog list`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Catalog command called")
		setup := Setup()

		catalog.List(setup.RestConfig)
	},
}

var catalogSelectCmd = &cobra.Command{
	Use:     "select",
	Short:   "Select a service and install it",
	Long:    "Select a service and install it in a namespace.",
	Example: ` sb catalog select`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(rootCmd.Use + " v" + VERSION + " (" + GITCOMMIT + ")")
	},
}
var catalogBindCmd = &cobra.Command{
	Use:     "bind",
	Short:   "Bind a service to a secret's namespace",
	Long:    "Bind a service to a secret's namespace.",
	Example: ` sb catalog bind`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(rootCmd.Use + " v" + VERSION + " (" + GITCOMMIT + ")")
	},
}

func init() {
	catalogCmd.AddCommand(catalogListCmd)
	catalogCmd.AddCommand(catalogSelectCmd)
	catalogCmd.AddCommand(catalogBindCmd)

	catalogCmd.Annotations = map[string]string{"command": "catalog"}
	rootCmd.AddCommand(catalogCmd)
}
