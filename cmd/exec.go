package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"

	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
	"fmt"
)

var (
	supervisordBin = "/var/lib/supervisord/bin/supervisord"
	supervisordCtl = "ctl"
	cmdName = "run-java"
)

var exeCmd = &cobra.Command{
	Use:   "exec [options]",
	Short: "Stop, start or restart your SpringBoot's application.",
	Long:  `Stop, start or restart your SpringBoot's application.`,
	Example: fmt.Sprintf("%s\n%s\n%s",
		execStartCmd.Example,
		execStopCmd.Example,
		execRestartCmd.Example),
}

var execStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start your SpringBoot's application.",
	Long:  `Start your SpringBoot's application.`,
	Example: `  sb exec start`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Exec start command called")

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
		action := "start"

		log.Info("[Step 5] - Start the Spring Boot application ...")
		log.Debug("Command :",[]string{"rsh",podName,supervisordBin,supervisordCtl,action,cmdName})
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,supervisordBin,supervisordCtl,action,cmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"logs",podName,"-f"}})
	},
}
var execStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop your SpringBoot's application.",
	Long:  `Stop your SpringBoot's application.`,
	Example: `  sb exec stop`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Exec stop command called")

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
		action := "stop"

		log.Info("[Step 5] - Start the Spring Boot application ...")
		log.Debug("Command :",[]string{"rsh",podName,supervisordBin,supervisordCtl,action,cmdName})
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,supervisordBin,supervisordCtl,action,cmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"logs",podName,"-f"}})
	},
}
var execRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart your SpringBoot's application.",
	Long:  `Restart your SpringBoot's application.`,
	Example: `  sb exec restart`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Exec restart command called")

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

		log.Info("[Step 5] - Restart the Spring Boot application ...")
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,supervisordBin,supervisordCtl,"stop",cmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,supervisordBin,supervisordCtl,"start",cmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"logs",podName,"-f"}})
	},
}

func init() {
	exeCmd.AddCommand(execStartCmd)
	exeCmd.AddCommand(execStopCmd)
	exeCmd.AddCommand(execRestartCmd)

	runCmd.Annotations = map[string]string{"command": "run"}
	rootCmd.AddCommand(exeCmd)
}

