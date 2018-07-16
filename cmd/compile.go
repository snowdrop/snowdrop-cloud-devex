package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"

	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
)

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile local's project within the development's pod",
	Long:  `Compile local's project within the development's pod.`,
	Example: ` sb compile`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Compile command called")

		// Parse MANIFEST - Step 1
		application := parseManifest()
		// Add Namespace's value
		application.Namespace = namespace

		// Get K8s' config file - Step 2
		kubeCfg := getK8Config(*cmd)

		// Create Kube Rest's Config Client - Step 3
		clientset := createClientSet(kubeCfg)

		// Wait till the dev's pod is available
		log.Info("[Step 4] - Wait till the dev's pod is available")
		pod, err := buildpack.WaitAndGetPod(clientset,application)
		if err != nil {
			log.Error("Pod watch error",err)
		}

		podName := pod.Name

		log.Info("[Step 5] - Compile ...")
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,"/var/lib/supervisord/bin/supervisord","ctl","start","compile-java"}})
		oc.ExecCommand(oc.Command{Args: []string{"logs",podName,"-f"}})
	},
}

func init() {
	compileCmd.Annotations = map[string]string{"command": "compile"}
	rootCmd.AddCommand(compileCmd)
}

