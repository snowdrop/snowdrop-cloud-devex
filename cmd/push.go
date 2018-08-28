package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"

	"fmt"
	"github.com/snowdrop/k8s-supervisor/pkg/common/oc"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {
	var (
		mode      string
		artefacts = []string{"src", "pom.xml"}
		modes     = []string{"source", "binary"}
	)

	pushCmd := &cobra.Command{
		Use:     "push",
		Short:   "Push local code to the development pod",
		Long:    `Push local code to the development pod.`,
		Example: ` sb push`,
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var valid bool
			for _, value := range modes {
				if mode == value {
					valid = true
				}
			}

			if !valid {
				log.WithField("mode", mode).Fatal("The provided mode is not supported: ")
			}

			log.Infof("Push command called with mode '%s'", mode)

			setup, pod := SetupAndWaitForPod()
			podName := pod.Name
			containerName := setup.Application.Name

			log.Info("Copy files from the local developer project to the pod")

			switch mode {
			case "source":
				for i := range artefacts {
					log.Debug("Artefact : ", artefacts[i])
					args := []string{"cp", oc.Client.Pwd + "/" + artefacts[i], podName + ":/tmp/src/", "-c", containerName}
					log.Infof("Copy cmd : %s", args)
					oc.ExecCommand(oc.Command{Args: args})
				}
			case "binary":
				targetDir := oc.Client.Pwd + "/target/"
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					log.Error("No output found! Please build the application with 'mvn clean package' before pushing")
				} else {
					filesInTarget, err := ioutil.ReadDir(oc.Client.Pwd + "/target/")
					if err != nil {
						panic(err)
					}

					uberJarFile := ""
					for _, f := range filesInTarget {
						if filepath.Ext(f.Name()) == ".jar" {
							uberJarFile = targetDir + f.Name()
						}
					}

					if uberJarFile != "" {
						args := []string{"cp", uberJarFile, podName + ":/deployments/app.jar", "-c", containerName}
						log.Infof("Copy cmd : %s", args)
						oc.ExecCommand(oc.Command{Args: args})
					} else {
						log.Error("No uber-jar file found! Please build the application with 'mvn clean package' before pushing")
					}

				}

			}
		},
	}

	pushCmd.Flags().StringVarP(&mode, "mode", "", "source",
		fmt.Sprintf("Mode used to push the code to the development pod. Supported modes are '%s'", strings.Join(modes, ",")))
	pushCmd.MarkFlagRequired("mode")
	pushCmd.Annotations = map[string]string{"command": "push"}

	rootCmd.AddCommand(pushCmd)
}
