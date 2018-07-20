package buildpack

import (
	buildclientsetv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	restclient "k8s.io/client-go/rest"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"log"
)

func CreateBuild(config *restclient.Config, appConfig types.Application) {
	buildClient, err := buildclientsetv1.NewForConfig(config)
	if err != nil {
	}

	build := buildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: appConfig.Name,
			Labels: map[string]string{
				"app": appConfig.Name,
				"io.openshift.odo": "inject-supervisord",
			},
		},
		Spec: buildv1.BuildSpec{
			CommonSpec: buildv1.CommonSpec{
				Source:buildv1.BuildSource{
					Type: buildv1.BuildSourceBinary,
				},
				Strategy: buildv1.BuildStrategy{
					Type: buildv1.DockerBuildStrategyType,
				},
				Output:buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: appConfig.Name + "2" + ":latest",
					},
				},
			},
		},

	}

	_, errbuild := buildClient.Builds(appConfig.Namespace).Create(&build)
	if errbuild != nil {
		log.Fatalf("Unable to create Build: %s", errbuild.Error())
	}
}

