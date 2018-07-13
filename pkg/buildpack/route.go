package buildpack

import (
	log "github.com/sirupsen/logrus"
	"github.com/ghodss/yaml"

	restclient "k8s.io/client-go/rest"

	routeclientsetv1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	routev1 "github.com/openshift/api/route/v1"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
)

func CreateRouteTemplate(config *restclient.Config, application types.Application) {
	routeclientsetv1, errrouteclientsetv1 := routeclientsetv1.NewForConfig(config)
	if errrouteclientsetv1 != nil {
		log.Fatal("error creating routeclientsetv1", errrouteclientsetv1.Error())
	}

	// Parse Route Template
	var b = ParseTemplate("route", application)

	// Create Route struct using the generated Route string
	route := routev1.Route{}
	errYamlParsing := yaml.Unmarshal(b.Bytes(), &route)
	if errYamlParsing != nil {
		panic(errYamlParsing)
	}

	// Create the route ...
	_, errRoute := routeclientsetv1.Routes(application.Namespace).Create(&route)
	if errRoute != nil {
		log.Fatal("error creating route", errRoute.Error())
	}

}
