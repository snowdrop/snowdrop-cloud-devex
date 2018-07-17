package buildpack

import (
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"

	imagev1 "github.com/openshift/api/image/v1"
	imageclientsetv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	restclient "k8s.io/client-go/rest"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
)

var (
	appImagename         = "spring-boot-http"
	version              = "1.0"
	supervisordimagename = "copy-supervisord"
)

func CreateImageStreamTemplate(config *restclient.Config, appConfig types.Application) {
	imageClient, err := imageclientsetv1.NewForConfig(config)
	if err != nil {
	}

	images := []types.Image{
		{
			Name: appImagename,
			Repo: "quay.io/snowdrop/spring-boot-s2i",
		},
		{
			Name:           supervisordimagename,
			Repo:           "quay.io/snowdrop/supervisord",
			AnnotationCmds: true,
		},
	}

	appCfg := appConfig
	for _, img := range images {

		appCfg.Image = img

		// Parse ImageStream Template
		var b = ParseTemplate("imagestream", appCfg)

		// Create ImageStream struct using the generated ImageStream string
		img := imagev1.ImageStream{}
		errYamlParsing := yaml.Unmarshal(b.Bytes(), &img)
		if errYamlParsing != nil {
			panic(errYamlParsing)
		}

		_, errImages := imageClient.ImageStreams(appConfig.Namespace).Create(&img)
		if errImages != nil {
			log.Fatalf("Unable to create ImageStream: %s", errImages.Error())
		}
	}

}
