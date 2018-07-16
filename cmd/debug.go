package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
)

var (
	ports string
)
var debugCmd = &cobra.Command{
	Use:   "debug [flags]",
	Short: "Debug your SpringBoot's application",
	Long:  `Debug your SpringBoot's application.`,
	Example: ` sb debug -p 5005:5005`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Run command called")

		// Parse MANIFEST - Step 1
		application := parseManifest()
		// Add Namespace's value
		application.Namespace = namespace

		// Get K8s' config file - Step 2
		kubeCfg := getK8Config()

		// Create Kube Rest's Config Client
		log.Info("[Step 3] - Create kube Rest config client using config's file of the developer's machine")
		kubeRestClient, err := clientcmd.BuildConfigFromFlags(kubeCfg.MasterURL, kubeCfg.Config)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %s", err.Error())
		}

		clientset, errclientset := kubernetes.NewForConfig(kubeRestClient)
		if errclientset != nil {
			log.Fatalf("Error building kubernetes clientset: %s", errclientset.Error())
		}

		// Wait till the dev's pod is available
		log.Info("[Step 4] - Wait till the dev's pod is available")
		pod, err := buildpack.WaitAndGetPod(clientset,application)
		if err != nil {
			log.Error("Pod watch error",err)
		}

		podName := pod.Name

		// Append Debug Env Vars and update POD
		//log.Info("[Step 5] - Add new ENV vars for remote Debugging")
		//pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env,debugEnvVars()...)
		//clientset.CoreV1().Pods(application.Namespace).Update(pod)

		// Start Java Application
		log.Info("[Step 5] - Restart Java Application")
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,"/var/lib/supervisord/bin/supervisord","ctl","stop","run-java"}})
		oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,"/var/lib/supervisord/bin/supervisord","ctl","start","run-java"}})

		// Forward local to Remote port
		log.Info("[Step 6] - Remote Debug the spring Boot Application ...")
		oc.ExecCommand(oc.Command{Args: []string{"port-forward",podName,ports}})
	},
}

func init() {
	debugCmd.Flags().StringVarP(&ports,"ports","p","5005:5005","Local and remote ports to be used to forward trafic between the dpo and your machine.")
	//debugCmd.MarkFlagRequired("ports")

	debugCmd.Annotations = map[string]string{"command": "debug"}
	rootCmd.AddCommand(debugCmd)
}

func debugEnvVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "JAVA_DEBUG",
			Value: "true",
		},
		{
			Name: "JAVA_DEBUG_PORT",
			Value: "5005",
		},
	}
}

