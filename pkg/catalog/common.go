package catalog

import (
	"github.com/pkg/errors"
	restclient "k8s.io/client-go/rest"
	"strings"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"

	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
)

func GetClient(config *restclient.Config) *servicecatalogclienset.ServicecatalogV1beta1Client {
	serviceCatalogV1Client, err := servicecatalogclienset.NewForConfig(config)
	if err != nil {
		log.Fatal("error creating service catalog Clientset", err.Error())
	}
	return serviceCatalogV1Client
}

// GetClusterServiceClasses queries the service service catalog to get available clusterServiceClasses
func GetClusterServiceClasses(scc *servicecatalogclienset.ServicecatalogV1beta1Client) ([]scv1beta1.ClusterServiceClass, error) {
	classList, err := scc.ClusterServiceClasses().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list cluster service classes")
	}
	itemsFromAutomationBroker := make([]scv1beta1.ClusterServiceClass, 0)
	for _, class := range classList.Items {
		// we are only interested in services from the automation-service-broker
		// this broker however could have various names depending on the setup
		if strings.Contains(class.Spec.ClusterServiceBrokerName, "ansible-service-broker") ||
			strings.Contains(class.Spec.ClusterServiceBrokerName, "openshift-automation-service-broker") ||
			strings.Contains(class.Spec.ClusterServiceBrokerName, "automation-broker") {

			itemsFromAutomationBroker = append(itemsFromAutomationBroker, class)
		}
	}
	return itemsFromAutomationBroker, nil
}

func GetServicePlanNames(stringMap map[string]scv1beta1.ClusterServicePlan) (keys []string) {
	keys = make([]string, len(stringMap))

	i := 0
	for k := range stringMap {
		keys[i] = k
		i++
	}

	return keys
}

func GetServiceClassesCategories(categories map[string][]scv1beta1.ClusterServiceClass) (keys []string) {
	keys = make([]string, len(categories))

	i := 0
	for k := range categories {
		keys[i] = k
		i++
	}

	return keys
}

func GetServiceClassesByCategory(scc *servicecatalogclienset.ServicecatalogV1beta1Client) (categories map[string][]scv1beta1.ClusterServiceClass, err error) {
	categories = make(map[string][]scv1beta1.ClusterServiceClass)
	classes, err := GetClusterServiceClasses(scc)

	for _, class := range classes {
		tags := class.Spec.Tags
		var meta map[string]interface{}
		json.Unmarshal(class.Spec.ExternalMetadata.Raw, &meta)
		category := "other"
		if len(tags) > 0 {
			category = tags[0]
		}
		if len(category) > 0 {
			categories[category] = append(categories[category], class)
		}
	}

	return categories, err
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

// BuildParametersFrom converts a map of secrets names to secret keys to the
// type consumed by the ServiceCatalog API.
func BuildParametersFrom(secrets map[string]string) []scv1beta1.ParametersFromSource {
	params := make([]scv1beta1.ParametersFromSource, 0, len(secrets))

	for secret, key := range secrets {
		param := scv1beta1.ParametersFromSource{
			SecretKeyRef: &scv1beta1.SecretKeyReference{
				Name: secret,
				Key:  key,
			},
		}

		params = append(params, param)
	}
	return params
}
