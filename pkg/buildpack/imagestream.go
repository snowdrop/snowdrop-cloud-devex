package buildpack

import (
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"

	imagev1 "github.com/openshift/api/image/v1"
	imageclientsetv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	restclient "k8s.io/client-go/rest"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
)

func CreateImageStreamTemplate(config *restclient.Config, appConfig types.Application, images []types.Image) {
	imageClient := getImageClient(config)

	appCfg := appConfig
	for _, img := range images {

		appCfg.Image = img

		// first check that the image stream hasn't already been created
		if oc.Exists("imagestream", img.Name) {
			log.Infof("'%s' ImageStream already exists, skipping", img.Name)
		} else {
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
}

func getImageClient(config *restclient.Config) *imageclientsetv1.ImageV1Client {
	imageClient, err := imageclientsetv1.NewForConfig(config)
	if err != nil {
		log.Fatal("Couldn't get ImageV1Client: %s", err)
	}
	return imageClient
}

func DeleteImageStreams(config *restclient.Config, appConfig types.Application, images []types.Image) {
	for _, img := range images {
		// first check that the image stream hasn't already been created
		if oc.Exists("imagestream", img.Name) {
			client := getImageClient(config)
			err := client.ImageStreams(appConfig.Namespace).Delete(img.Name, deleteOptions)
			if err != nil {
				log.Fatalf("Unable to delete ImageStream: %s", img.Name)
			}
		}
	}
}

func CreateTypeImage(dockerImage bool, name string, tag string, repo string, annotationCmd bool) *types.Image {
	return &types.Image{
		DockerImage:    dockerImage,
		Name:           name,
		Repo:           repo,
		AnnotationCmds: annotationCmd,
		Tag:            tag,
	}
}
