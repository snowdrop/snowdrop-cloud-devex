package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"

	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run your SpringBoot's application.",
	Long:  `Run your SpringBoot's application.`,
	Example: ` sb run`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Run command called")

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
		supervisordBin := "/var/lib/supervisord/bin/supervisord"
		supervisordCtl := "ctl"
		cmdName := "run-java"

		log.Info("[Step 5] - Launch the Spring Boot application ...")
		log.Debug("Command :",[]string{"rsh",podName,supervisordBin,supervisordCtl,"stop",cmdName})
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,supervisordBin,supervisordCtl,"stop",cmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,supervisordBin,supervisordCtl,"start",cmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"logs",podName,"-f"}})
	},
}

func init() {
	runCmd.Annotations = map[string]string{"command": "run"}
	rootCmd.AddCommand(runCmd)
}

