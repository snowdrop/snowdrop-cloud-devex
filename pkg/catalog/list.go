package catalog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

const (
	BROKER_NAME = "openshift-automation-service-broker"
)

func List(config *restclient.Config) {
	serviceCatalogClient := GetClient(config)
	classes, _ := getClusterServiceClasses(serviceCatalogClient)
	log.Info("List of services")
	log.Info("================")

	for  _, class  := range classes {
		log.Infof("ID : %s, Class description : %s", class.Name, class.Spec.Description)
		metaData, _ := PrettyPrint(class.Spec.ExternalMetadata.Raw)
		log.Infof("Additional info : %s",metaData)

		log.Info("---------------------------------------------------")
	}
}

// GetClusterServiceClasses queries the service service catalog to get available clusterServiceClasses
func getClusterServiceClasses(scc *servicecatalogclienset.ServicecatalogV1beta1Client) ([]scv1beta1.ClusterServiceClass, error) {
	classList, err := scc.ClusterServiceClasses().List(metav1.ListOptions{FieldSelector: "spec.clusterServiceBrokerName=" + BROKER_NAME})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list cluster service classes")
	}
	return classList.Items, nil
}