package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/k8s-supervisor/pkg/catalog"
	"github.com/spf13/cobra"
)

func init() {
	catalogListCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all available services from the catalog",
		Long:    "List all available services from the Service Catalog's broker.",
		Example: ` sb catalog list`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Catalog list command called")
			setup := Setup()

			catalog.List(setup.RestConfig)
		},
	}

	catalogInstanceCmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a service instance",
		Long:    "Create a service instance and install it in a namespace.",
		Example: ` sb catalog create <instance name>`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Catalog select command called")
			setup := Setup()

			catalog.Create(setup.RestConfig, setup.Application, args[0])
		},
	}

	var (
		secret   string
		instance string
	)
	catalogBindCmd := &cobra.Command{
		Use:     "bind",
		Short:   "Bind a service to a secret's namespace",
		Long:    "Bind a service to a secret's namespace.",
		Example: ` sb catalog bind --secret foo --toInstance instance`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Catalog Bind command called")
			setup := Setup()

			catalog.Bind(setup.RestConfig, setup.Application, instance, secret)
			catalog.MountSecretAsEnvFrom(setup.RestConfig, setup.Application, secret)
		},
	}
	catalogBindCmd.Flags().StringVarP(&secret, "secret", "s", "secret name", "Secret name to bind")
	catalogBindCmd.Flags().StringVarP(&instance, "toInstance", "i", "instance name", "Instance name to bind the instance to")

	catalogCmd := &cobra.Command{
		Use:   "catalog [options]",
		Short: "List, select or bind a service from a catalog.",
		Long:  `List, select or bind a service from a catalog.`,
		Example: fmt.Sprintf("%s\n%s\n%s",
			catalogListCmd.Example,
			catalogInstanceCmd.Example,
			catalogBindCmd.Example),
	}

	catalogCmd.AddCommand(catalogListCmd)
	catalogCmd.AddCommand(catalogInstanceCmd)
	catalogCmd.AddCommand(catalogBindCmd)

	catalogCmd.Annotations = map[string]string{"command": "catalog"}
	rootCmd.AddCommand(catalogCmd)
}
