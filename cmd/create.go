package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/snowdrop/k8s-supervisor/pkg/scaffold"
)

var createCmd = &cobra.Command{
	Use:     "create [flags]",
	Short:   "Create a Spring Boot maven project",
	Long:    `Create a Spring Boot maven project".`,
	Example: ` sb create`,
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Create command called")
		p := scaffold.Project{
			GroupId: "me.snowdrop",
			ArtifactId: "cool",
			Version: "1.0",
			PackageName: "io.openshift",
			SnowdropBomVersion: "1.5.15.Final",
			SpringVersion: "1.5.15.Release",
		}
		scaffold.GenerateProjectFiles(p)
	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	createCmd.Annotations = map[string]string{"command": "create"}

	rootCmd.AddCommand(createCmd)
}
