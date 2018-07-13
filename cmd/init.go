package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
)

var (
	Namespace string
)

var initCmd = &cobra.Command{
	Use:   "init [flags]",
	Short: "Create a development's pod for the component.",
	Long:  `Create a development's pod for the component.`,
	Example: ` odo init -n bootapp`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Init called")
		log.Debug("Next action to be performed is ....")

		// Retrieve stdout / io.Writer
		//stdout := os.Stdout

		// Retrieve the client
		// client := getOcClient()

		// Application
		// currentApplication, err := application.GetCurrent(client)
		//checkError(err, "")

		// Project
		// currentProject := project.GetCurrent(client)

		//var argComponent string

		//if len(args) == 1 {
		//	argComponent = args[0]
		//}

		// Retrieve and set the currentComponent
		// currentComponent := getComponent(client, argComponent, currentApplication, currentProject)

		// Retrieve the log
		// err = component.GetLogs(client, currentComponent, currentApplication, logFollow, stdout)
		//checkError(err, "Unable to retrieve logs, does your component exist?")
	},
}

func init() {
	initCmd.Flags().StringP("namespace", "n","",  "Namespace/project")
	initCmd.MarkFlagRequired("namespace")

	// Add a defined annotation in order to appear in the help menu
	initCmd.Annotations = map[string]string{"command": "init"}

	rootCmd.AddCommand(initCmd)
}

