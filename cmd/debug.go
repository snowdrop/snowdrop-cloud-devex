package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"

	"github.com/cmoulliard/k8s-supervisor/pkg/common/config"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
)

var (
	ports string
)
var debugCmd = &cobra.Command{
	Use:     "debug [flags]",
	Short:   "Debug your SpringBoot application",
	Long:    `Debug your SpringBoot application.`,
	Example: ` sb debug -p 5005:5005`,
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Debug command called")

		_, pod := SetupAndWaitForPod()
		podName := pod.Name

		// Append Debug Env Vars and update POD
		//log.Info("[Step 5] - Add new ENV vars for remote Debugging")
		//pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env,debugEnvVars()...)
		//clientset.CoreV1().Pods(application.Namespace).Update(pod)

		log.Info("Restart the Spring Boot application ...")
		oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "stop", config.RunCmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "start", config.RunCmdName}})

		// Forward local to Remote port
		log.Info("Remote Debug the Spring Boot Application ...")
		oc.ExecCommand(oc.Command{Args: []string{"port-forward", podName, ports}})
	},
}

func init() {
	debugCmd.Flags().StringVarP(&ports, "ports", "p", "5005:5005", "Local and remote ports to be used to forward traffic between the dev pod and your machine.")
	//debugCmd.MarkFlagRequired("ports")

	debugCmd.Annotations = map[string]string{"command": "debug"}
	rootCmd.AddCommand(debugCmd)
}

func debugEnvVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "JAVA_DEBUG",
			Value: "true",
		},
		{
			Name:  "JAVA_DEBUG_PORT",
			Value: "5005",
		},
	}
}
