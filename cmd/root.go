package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "odo",
	Short: "Odo (Openshift Do)",

	Long: `Odo (OpenShift Do) is a CLI tool for running OpenShift applications in a fast and automated matter.
Odo reduces the complexity of deployment by adding iterative development without the worry of deploying your source code.
Find more information at https://github.com/redhat-developer/odo`,

	Example: `    # Creating and deploying a Node.js project
    git clone https://github.com/openshift/nodejs-ex && cd nodejs-ex
    odo init -p namespace
    odo push -t maven
    odo compile -t maven
    odo run`,
}

func init() {
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "","Path to a kubeconfig ($HOME/.kube/config). Only required if out-of-cluster.")
	rootCmd.PersistentFlags().StringP("masterurl", "m", "","The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		checkError(err,"Root execution")
	}
}

// checkError prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
func checkError(err error, context string, a ...interface{}) {
	if err != nil {
		log.Debugf("Error:\n%v", err)
		if context == "" {
			fmt.Println(errors.Cause(err))
		} else {
			fmt.Printf(fmt.Sprintf("%s\n", context), a...)
		}
		os.Exit(1)
	}
}
