package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"
)

var initCmd = &cobra.Command{
	Use:     "init [flags]",
	Short:   "Create a development's pod for the component",
	Long:    `Create a development's pod for the component.`,
	Example: ` sb init -n bootapp`,
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Init command called")

		setup := Setup()
		// Create ImageStream
		log.Info("Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
		buildpack.CreateImageStreamTemplate(setup.RestConfig, setup.Application)

		// Create PVC
		log.Info("Create PVC to storage m2 repo")
		buildpack.CreatePVC(setup.Clientset, setup.Application, "1Gi")

		log.Info("Create DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
		dc := buildpack.CreateDeploymentConfig(setup.RestConfig, setup.Application)

		log.Info("Create Service using Template")
		buildpack.CreateServiceTemplate(setup.Clientset, dc, setup.Application)

		log.Info("Create Route using Template")
		buildpack.CreateRouteTemplate(setup.RestConfig, setup.Application)
	},
}

func init() {

	// Add a defined annotation in order to appear in the help menu
	initCmd.Annotations = map[string]string{"command": "init"}

	rootCmd.AddCommand(initCmd)
}
