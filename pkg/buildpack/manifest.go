package buildpack

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"github.com/ghodss/yaml"
	"encoding/json"
)

func ParseManifest(path string) types.Application {
	log.Debug("Parse Application's Config : ", path)

	source, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// Create an Application with default values
	appConfig := types.NewApplication()

	err = yaml.Unmarshal(source, &appConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Application's config")
	log.Debug("--------------------")
	appFormatted, _ := json.Marshal(appConfig)
	log.Debug(string(appFormatted))

	return appConfig
}
