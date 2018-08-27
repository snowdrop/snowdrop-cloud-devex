package catalog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/pkg/errors"
)

const (
	POSTGRESQL_CLUSTER_EXT_NAME = "dh-postgresql-apb"
	POSTGRESQL_VERSION          = "9.4"
)

func Select(config *restclient.Config) {
	serviceCatalogClient := GetClient(config)
	createServiceInstance(serviceCatalogClient, "crud", "dh-postgresql", "", nil)
}

// CreateServiceInstance creates service instance from service catalog
func createServiceInstance(scc *servicecatalogclienset.ServicecatalogV1beta1Client, ns string, serviceName string, componentType string, labels map[string]string) error {
	// Creating Service Instance
	_, err := scc.ServiceInstances(ns).Create(
		&scv1beta1.ServiceInstance{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceInstance",
				APIVersion: "servicecatalog.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{"kubernetes-incubator/service-catalog"},
				Name:       serviceName,
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