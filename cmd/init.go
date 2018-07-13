package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	"github.com/cmoulliard/k8s-supervisor/pkg/common/config"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"os"
)

var initCmd = &cobra.Command{
	Use:   "init [flags]",
	Short: "Create a development's pod for the component.",
	Long:  `Create a development's pod for the component.`,
	Example: ` sb init -n bootapp`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("sb Init command called")
		log.Debug("Namespace : ", namespace)

		// Parse MANIFEST
		log.Info("[Step 1] - Parse MANIFEST of the project if it exists")
		current, _ := os.Getwd()
		application := buildpack.ParseManifest(current + "/MANIFEST")

		// Add Namespace's value
		application.Namespace = namespace

		// Get K8s' config file
		log.Info("[Step 2] - Get K8s config file")
		var kubeCfg = config.NewKube()
		if cmd.Flag("kubeconfig").Value.String() == "" {
			kubeCfg.Config = config.HomeKubePath()
		} else {
			kubeCfg.Config = cmd.Flag("kubeconfig").Value.String()
		}
		log.Debug("Kubeconfig : ",kubeCfg)

		// Execute oc command to switch to the namespace defined
		log.Info("[Step 3] - Get k8s default's namespace")
		oc.ExecCommand(oc.Command{Args: []string{"project",application.Namespace}})

		// Create Kube Rest's Config Client
		log.Info("[Step 4] - Create kube Rest config client using config's file of the developer's machine")
		kubeRestClient, err := clientcmd.BuildConfigFromFlags(kubeCfg.MasterURL, kubeCfg.Config)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %s", err.Error())
		}

		// Create ImageStream
		log.Info("[Step 5] - Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
		buildpack.CreateImageStreamTemplate(kubeRestClient,application)

		log.Info("[Step 6] - Create DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
		dc := buildpack.CreateDeploymentConfig(kubeRestClient,application)

		clientset, errclientset := kubernetes.NewForConfig(kubeRestClient)
		if errclientset != nil {
			log.Fatalf("Error building kubernetes clientset: %s", errclientset.Error())
		}

		log.Info("[Step 7] - Create Service using Template")
		buildpack.CreateServiceTemplate(clientset, dc, application)

		log.Info("[Step 8] - Create Route using Template")
		buildpack.CreateRouteTemplate(kubeRestClient,application)
	},
}

func init() {

	// Add a defined annotation in order to appear in the help menu
	initCmd.Annotations = map[string]string{"command": "init"}

	rootCmd.AddCommand(initCmd)
}

