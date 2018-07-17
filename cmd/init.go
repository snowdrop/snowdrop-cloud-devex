package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"

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
		log.Debugf("Namespace: %s", namespace)

		// Parse MANIFEST
		application := parseManifest()

		// Get K8s' config file
		kubeCfg := getK8Config(*cmd)

		// Switch to namespace if specified or retrieve the current one if not
		currentNs, err := oc.ExecCommand(oc.Command{Args: []string{"project", "-q", namespace}})
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Using '%s' namespace", currentNs)
		application.Namespace = currentNs

		// Create Kube Rest's Config Client
		restConfig := createKubeRestconfig(kubeCfg)
		clientset := createClientSet(kubeCfg, restConfig)

		// Create ImageStream
		log.Info("Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
		buildpack.CreateImageStreamTemplate(restConfig, application)

		// Create PVC
		log.Info("Create PVC to storage m2 repo")
		buildpack.CreatePVC(clientset, application, "1Gi")

		log.Info("Create DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
		dc := buildpack.CreateDeploymentConfig(restConfig, application)

		log.Info("Create Service using Template")
		buildpack.CreateServiceTemplate(clientset, dc, application)

		log.Info("Create Route using Template")
		buildpack.CreateRouteTemplate(restConfig, application)
	},
}

func init() {

	// Add a defined annotation in order to appear in the help menu
	initCmd.Annotations = map[string]string{"command": "init"}

	rootCmd.AddCommand(initCmd)
}
