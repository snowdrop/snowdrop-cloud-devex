package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"
)

var initCmd = &cobra.Command{
	Use:   "init [flags]",
	Short: "Create a development's pod for the component",
	Long:  `Create a development's pod for the component.`,
	Example: ` sb init -n bootapp`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Init command called")
		log.Debug("Namespace : ", namespace)

		// Parse MANIFEST - Step 1
		application := parseManifest()

		// Add Namespace's value
		application.Namespace = namespace

		// Get K8s' config file - Step 2
		kubeCfg := getK8Config(*cmd)

		// Switch to namespace if specified or retrieve the current one if not
		currentNs, err := oc.ExecCommand(oc.Command{Args: []string{"project", "-q", namespace}})
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Using '%s' namespace", currentNs)
		application.Namespace = currentNs

		// Create Kube Rest's Config Client - Step 3
		restConfig := createKubeRestconfig(kubeCfg)
		clientset := createClientSet(kubeCfg)

		// Create ImageStream
		log.Info("[Step 4] - Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
		buildpack.CreateImageStreamTemplate(restConfig,application)

		// Create PVC
		log.Info("[Step 5] - Create PVC to storage m2 repo")
		buildpack.CreatePVC(clientset,application,"1Gi")

		log.Info("[Step 6] - Create DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
		dc := buildpack.CreateDeploymentConfig(restConfig,application)

		log.Info("[Step 7] - Create Service using Template")
		buildpack.CreateServiceTemplate(clientset, dc, application)

		log.Info("[Step 8] - Create Route using Template")
		buildpack.CreateRouteTemplate(restConfig,application)
	},
}

func init() {

	// Add a defined annotation in order to appear in the help menu
	initCmd.Annotations = map[string]string{"command": "init"}

	rootCmd.AddCommand(initCmd)
}

