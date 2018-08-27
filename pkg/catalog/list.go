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

// GetServiceInstanceList returns list service instances
func getServiceInstanceList(scc *servicecatalogclienset.ServicecatalogV1beta1Client, project string, selector string) ([]scv1beta1.ServiceInstance, error) {
	// List ServiceInstance according to given selectors
	svcList, err := scc.ServiceInstances(project).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list ServiceInstances")
	}

	return svcList.Items, nil
}

// CreateServiceInstance creates service instance from service catalog
func createServiceInstance(scc *servicecatalogclienset.ServicecatalogV1beta1Client, ns string, componentName string, componentType string, labels map[string]string) error {
	// Creating Service Instance
	_, err := scc.ServiceInstances(ns).Create(
		&scv1beta1.ServiceInstance{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceInstance",
				APIVersion: "servicecatalog.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{"kubernetes-incubator/service-catalog"},
				Name:       componentName,
				Namespace:  ns,
				Labels:     labels,
			},
			Spec: scv1beta1.ServiceInstanceSpec{
				PlanReference: scv1beta1.PlanReference{
					ClusterServiceClassExternalName: componentType,
				},
			},
			Status: scv1beta1.ServiceInstanceStatus{},
		})

	if err != nil {
		return errors.Wrap(err, "unable to create service instance")
	}
	return nil
}

func getClient(config *restclient.Config) *servicecatalogclienset.ServicecatalogV1beta1Client {
	serviceCatalogV1Client, err := servicecatalogclienset.NewForConfig(config)
	if err != nil {
		log.Fatal("error creating service catalog Clientset", err.Error())
	}
	return serviceCatalogV1Client
}
