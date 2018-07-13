package buildpack

import (
	log "github.com/sirupsen/logrus"
	"github.com/ghodss/yaml"
	"k8s.io/client-go/kubernetes"

	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
)

func CreateServiceTemplate(clientset *kubernetes.Clientset, dc *appsv1.DeploymentConfig, application types.Application) {
	// Parse Service Template
	var b = ParseTemplate("service",application)

	// Create Service struct using the generated Service string
	svc := corev1.Service{}
	errYamlParsing := yaml.Unmarshal(b.Bytes(), &svc)
	if errYamlParsing != nil {
		panic(errYamlParsing)
	}

	// Create the Service
	_, errService := clientset.CoreV1().Services(application.Namespace).Create(&svc)
	if errService != nil {
		log.Fatal("Unable to create Service for %s", errService.Error())
	}
}

