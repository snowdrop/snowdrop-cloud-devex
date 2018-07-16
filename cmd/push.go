package cmd

import (
	"strings"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
)

var (
	mode string
	artefacts = []string{ "src", "pom.xml"}
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local code to the development's pod",
	Long:  `Push local code to the development's pod.`,
	Example: ` sb push`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		modeType := cmd.Flag("mode").Value.String();

		log.Info("Push command called")

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
		containerName := application.Name

		log.Info("[Step 5] - Copy files from the local developer's project to the pod")

		switch modeType {
		case "source":
			for i := range artefacts {
				log.Debug("Artefact : ",artefacts[i])
				log.Infof("Copy cmd : %s",[]string{"cp",oc.Client.Pwd + "/" + artefacts[i],podName+":/tmp/src/","-c",containerName})
				oc.ExecCommand(oc.Command{Args: []string{"cp",oc.Client.Pwd + "/" + artefacts[i],podName+":/tmp/src/","-c",containerName}})
			}
		case "binary":
			uberjarName := strings.Join([]string{application.Name,application.Version},"-") +  ".jar"
			log.WithField("uberjarname",uberjarName).Debug("Uber jar name : ")
			log.Infof("Copy cmd : %s",[]string{"cp",oc.Client.Pwd + "/target/" + uberjarName,podName+":/deployments","-c",containerName})
			oc.ExecCommand(oc.Command{Args: []string{"cp",oc.Client.Pwd + "/target/" + uberjarName,podName+":/deployments","-c",containerName}})
		default:
			log.WithField("mode",modeType).Fatal("The provided mode is not supported : ")
		}
	},
}

func init() {
	pushCmd.Flags().StringVarP(&mode,"mode","","","Source code or Binary compiled as uberjar within target directory")
	pushCmd.MarkFlagRequired("mode")
	pushCmd.Annotations = map[string]string{"command": "push"}

	rootCmd.AddCommand(pushCmd)
}

