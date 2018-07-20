package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"encoding/json"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"
	"github.com/openshift/api/apps/v1"
	"io/ioutil"
	"os"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
)

const stateFileName = ".sb.state"

type state struct {
	Finished bool `json:"finished"`
	Is       bool `json:"is"`
	Pvc      bool `json:"pvc"`
	Svc      bool `json:"svc"`
	Route    bool `json:"route"`
}

var initCmd = &cobra.Command{
	Use:     "init [flags]",
	Short:   "Create a development's pod for the component",
	Long:    `Create a development's pod for the component.`,
	Example: ` sb init -n bootapp`,
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Init command called")
		log.Debugf("Namespace: %s", namespace)

		// check if we have an existing record of the init state
		var state = &state{}
		if _, err := os.Stat(stateFileName); err == nil {
			source, err := ioutil.ReadFile(stateFileName)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(source, state)
			if err != nil {
				log.Fatalf("Couldn't read %s init state file, restarting from scratch", stateFileName)
			}
		}

		// only init if we haven't done it already
		if !state.Finished {
			setup := Setup()

			if !state.Is {
				// Create ImageStreams
				images := []types.Image{
					*buildpack.CreateTypeImage(setup.Application.Name, "latest", "quay.io/snowdrop/spring-boot-s2i", false),
					*buildpack.CreateTypeImage("copy-supervisord", "latest", "quay.io/snowdrop/supervisord", true),
				}

				buildpack.CreateImageStreamTemplate(setup.RestConfig, setup.Application, images)
				log.Info("Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
				state.Is = true
			} else {
				log.Info("ImageStreams already created")
			}

			if !state.Pvc {
				// Create PVC
				log.Info("Create PVC to store m2 repo")
				buildpack.CreatePVC(setup.Clientset, setup.Application, "1Gi")
				state.Pvc = true
			} else {
				log.Info("PVC already created")
			}

			var dc *v1.DeploymentConfig
			log.Info("Create or retrieve DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
			dc = buildpack.CreateOrRetrieveDeploymentConfig(setup.RestConfig, setup.Application)

			if !state.Svc {
				log.Info("Create Service using Template")
				buildpack.CreateServiceTemplate(setup.Clientset, dc, setup.Application)
				state.Svc = true
			} else {
				log.Info("Service already created")
			}

			if !state.Route {
				log.Info("Create Route using Template")
				buildpack.CreateRouteTemplate(setup.RestConfig, setup.Application)
				state.Route = true
			} else {
				log.Info("Route already created")
			}

			bytes, _ := json.Marshal(state)
			ioutil.WriteFile(stateFileName, bytes, 0644)
		} else {
			log.Info("init already done")
		}
	},
}

func init() {

	// Add a defined annotation in order to appear in the help menu
	initCmd.Annotations = map[string]string{"command": "init"}

	rootCmd.AddCommand(initCmd)
}
