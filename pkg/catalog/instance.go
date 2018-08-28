package catalog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/uuid"
	"github.com/pkg/errors"
)

const (
	CLASS_NAME                  = "dh-postgresql-apb"
	INSTANCE_NAME				= "my-postgresql-db"
	BINDING_NAME				= "my-postgresql-db-binding"
	PLAN                        = "dev"
	EXTERNAL_ID                 = "a7c00676-4398-11e8-842f-0ed5f89f718b"
	POSTGRESQL_VERSION          = "9.6"
	NS                          = "crud"
)

var (
	PARAMS = map[string]string{
		"postgresql_user": "luke",
		"postgresql_password": "secret",
		"postgresql_database": "my_data",
		"postgresql_version": "9.6",
	}
)

func Create(config *restclient.Config) {
	serviceCatalogClient := GetClient(config)
	log.Infof("Service instance will be created ...")
	createServiceInstance(serviceCatalogClient, NS, INSTANCE_NAME, CLASS_NAME, PLAN, EXTERNAL_ID, PARAMS)
	log.Infof("Service instance created")

	// Generate UUID otherwise binding will fail if we use same id as the instanceName, bidingName
	UUID := string(uuid.NewUUID())

	log.Infof("Let's generate a secret containing the parameters to be used by the application")
	bind(serviceCatalogClient,NS,BINDING_NAME,INSTANCE_NAME,UUID,INSTANCE_NAME, nil,nil)
}

// CreateServiceInstance creates service instance from service catalog
func createServiceInstance(scc *servicecatalogclienset.ServicecatalogV1beta1Client, ns string, instanceName string, className string, plan string, externalID string, params interface{}) error {
	// Creating Service Instance
	_, err := scc.ServiceInstances(ns).Create(
		&scv1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:       instanceName,
				Namespace:  ns,
			},
			Spec: scv1beta1.ServiceInstanceSpec{
				ExternalID: externalID,
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
	return nil
}