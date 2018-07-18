package buildpack

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"encoding/json"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"github.com/ghodss/yaml"
	"os"
	"path"
)

func ParseManifest(manifestPath string, appName string) types.Application {
	log.Debugf("Parsing Application Config at %s", manifestPath)

	// Create an Application with default values
	appConfig := types.NewApplication()

	// if we have a manifest file, use it to replace default values
	if _, err := os.Stat(manifestPath); err == nil {
		source, err := ioutil.ReadFile(manifestPath)
		if err != nil {
			panic(err)
		}

		err = yaml.Unmarshal(source, &appConfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Infof("No MANIFEST file detected, using default values")
	}

	// if we specified an application name, use it and override any set value
	if len(appName) > 0 {
		appConfig.Name = appName
	} else {
		if len(appConfig.Name) == 0 {
			// we need to set an application name, use the current directory name as default
			dir, _ := path.Split(manifestPath)
			appConfig.Name = path.Base(dir)
		}
	}

	log.Infof("Application '%s' configured", appConfig.Name)

	if log.GetLevel() == log.DebugLevel {
		log.Debug("Application's config")
		log.Debug("--------------------")
		appFormatted, _ := json.Marshal(appConfig)
		log.Debug(string(appFormatted))
	}

	return appConfig
}
