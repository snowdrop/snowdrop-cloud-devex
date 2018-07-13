package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	namespace string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sb",
	Short: "sb (ODO's Spring Boot prototype)",

	Long: `sb (ODO's Spring Boot prototype) is a prototype's project experimenting supervisord, MANIFEST's concepts`,

	Example: `    # Creating and deploying a spring Boot application
    git clone github.com/cmoulliard/k8s-supervisor && cd k8s-supervisor/spring-boot
    sb init -p namespace
    sb push
    sb compile
    sb run`,
}

func init() {
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "","Path to a kubeconfig ($HOME/.kube/config). Only required if out-of-cluster.")
	rootCmd.PersistentFlags().StringP("masterurl", "m", "","The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	// Global flag(s)
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace","n","",  "Namespace/project")
	//rootCmd.MarkPersistentFlagRequired("namespace")
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
