package catalog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pkg/errors"
	"fmt"
	"encoding/json"
)

const (
	CLASS_NAME                  = "dh-postgresql-apb"
	INSTANCE_NAME				= "my-postgresql-db"
	PLAN                        = "dev"
	POSTGRESQL_VERSION          = "9.6"
	NS                          = "crud"
)

var (
	PARAMS                      = map[string]string{
		"postgresql_user": "luke",
		"postgresql_password": "secret",
		"postgresql_database": "my_data",
		"postgresql_version": "9.6",
	}
)

func Create(config *restclient.Config) {
	serviceCatalogClient := GetClient(config)
	log.Infof("Service instance will be created ...")
	createServiceInstance(serviceCatalogClient, NS, INSTANCE_NAME, CLASS_NAME, PLAN, PARAMS)
}

// CreateServiceInstance creates service instance from service catalog
func createServiceInstance(scc *servicecatalogclienset.ServicecatalogV1beta1Client, ns string, instanceName string, className string, plan string, params interface{}) error {
	// Creating Service Instance
	_, err := scc.ServiceInstances(ns).Create(
		&scv1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:       instanceName,
				Namespace:  ns,
			},
			Spec: scv1beta1.ServiceInstanceSpec{
				PlanReference: scv1beta1.PlanReference{
					ClusterServiceClassExternalName: className,
					ClusterServicePlanExternalName:  plan,
				},
				Parameters: BuildParameters(params),
			},
		})

	if err != nil {
		return errors.Wrap(err, "unable to create service instance")
	}
	log.Infof("Service instance created")
	return nil
}

// BuildParameters converts a map of variable assignments to a byte encoded json document,
// which is what the ServiceCatalog API consumes.
func BuildParameters(params interface{}) *runtime.RawExtension {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		// This should never be hit because marshalling a map[string]string is pretty safe
		// I'd rather throw a panic then force handling of an error that I don't think is possible.
		panic(fmt.Errorf("unable to marshal the request parameters %v (%s)", params, err))
	}

	return &runtime.RawExtension{Raw: paramsJSON}
}