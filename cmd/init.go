package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	"github.com/cmoulliard/k8s-supervisor/pkg/common/config"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"
	"k8s.io/client-go/tools/clientcmd"
)

var namespace string

var initCmd = &cobra.Command{
	Use:   "init [flags]",
	Short: "Create a development's pod for the component.",
	Long:  `Create a development's pod for the component.`,
	Example: ` odo init -n bootapp`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Init command called")
		log.Debug("Namespace : ", namespace)

		application := types.Application{
			Namespace: namespace,
			// TODO -> Remove hard coded value
			Name: "spring-boot-http",
		}

		// Get K8s' config file
		var kubeCfg = config.NewKube()
		if cmd.Flag("kubeconfig").Value.String() == "" {
			kubeCfg.Config = config.HomeKubePath()
		} else {
			kubeCfg.Config = cmd.Flag("kubeconfig").Value.String()
		}
		log.Debug("Kubeconfig : ",kubeCfg)

		// Execute oc command to switch to the namespace defined
		log.Info("[Step 1] - Get k8s default's namespace")
		oc.ExecCommand(oc.Command{Args: []string{"project",application.Namespace}})

		// Create Kube Rest's Config Client
		log.Info("[Step 2] - Create kube Rest config client using config's file of the developer's machine")
		kubeRestClient, err := clientcmd.BuildConfigFromFlags(kubeCfg.MasterURL, kubeCfg.Config)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %s", err.Error())
		}

		// Create ImageStream
		log.Info("[Step 3] - Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
		buildpack.CreateImageStreamTemplate(kubeRestClient,application)

		/*		clientset, errclientset := kubernetes.NewForConfig(kubeRestClient)
		if errclientset != nil {
			log.Fatalf("Error building kubernetes clientset: %s", errclientset.Error())
		}*/
	},
}

func init() {
	initCmd.Flags().StringVarP(&namespace,"namespace", "n","",  "Namespace/project")
	initCmd.MarkFlagRequired("namespace")

	// Add a defined annotation in order to appear in the help menu
	initCmd.Annotations = map[string]string{"command": "init"}

	rootCmd.AddCommand(initCmd)
}

