package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
)

var cleanCmd = &cobra.Command{
	Use:     "clean [flags]",
	Short:   "Remove development pod for the component",
	Long:    `Remove development pod for the component.`,
	Example: ` sb clean`,
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Clean command called")

		setup := Setup()

		// Create ImageStreams
		images := []types.Image{
			*buildpack.CreateTypeImage(true, "dev-s2i", "latest", "quay.io/snowdrop/spring-boot-s2i", false),
			*buildpack.CreateTypeImage(true, "copy-supervisord", "latest", "quay.io/snowdrop/supervisord", true),
		}
		buildpack.DeleteImageStreams(setup.RestConfig, setup.Application, images)

		buildpack.DeletePVC(setup.Clientset, setup.Application)

		buildpack.DeleteDeploymentConfig(setup.RestConfig, setup.Application)

		buildpack.DeleteService(setup.Clientset, setup.Application)

		buildpack.DeleteRoute(setup.RestConfig, setup.Application)

		log.Info("Deleted resources")
	},
}

func init() {

	// Add a defined annotation in order to appear in the help menu
	cleanCmd.Annotations = map[string]string{"command": "clean"}

	rootCmd.AddCommand(cleanCmd)
}
