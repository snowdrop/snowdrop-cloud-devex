package catalog

import (
	"fmt"
	"log"
	"github.com/pkg/errors"
	restclient "k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
)
func List(config *restclient.Config) {
	serviceCatalogClient := getClient(config)
	classes, _ := getClusterServiceClasses(serviceCatalogClient)
	fmt.Print("List of services : ", classes)
}

func getClient(config *restclient.Config) *servicecatalogclienset.ServicecatalogV1beta1Client {
	serviceCatalogV1Client, err := servicecatalogclienset.NewForConfig(config)
	if err != nil {
		log.Fatal("error creating service catalog Clientset", err.Error())
	}
	return serviceCatalogV1Client
}

// GetClusterServiceClasses queries the service service catalog to get available clusterServiceClasses
func getClusterServiceClasses(scc *servicecatalogclienset.ServicecatalogV1beta1Client) ([]scv1beta1.ClusterServiceClass, error) {
	classList, err := scc.ClusterServiceClasses().List(metav1.ListOptions{FieldSelector: "spec.clusterServiceBrokerName=template-service-broker"})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list cluster service classes")
	}
	return classList.Items, nil
}
