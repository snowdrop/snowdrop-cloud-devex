package catalog

import (
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/pkg/errors"
)

const (
	BROKER_NAME = "openshift-automation-service-broker"
)

func List(config *restclient.Config) {
	serviceCatalogClient := getClient(config)
	classes, _ := getClusterServiceClasses(serviceCatalogClient)
	fmt.Println("List of services")

	for  _, class  := range classes {
		fmt.Println("Class description : ", class.Spec.Description)
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

func getClient(config *restclient.Config) *servicecatalogclienset.ServicecatalogV1beta1Client {
	serviceCatalogV1Client, err := servicecatalogclienset.NewForConfig(config)
	if err != nil {
		log.Fatal("error creating service catalog Clientset", err.Error())
	}
	return serviceCatalogV1Client
}
