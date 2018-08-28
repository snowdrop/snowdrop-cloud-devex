package catalog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	restclient "k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

func Bind(config *restclient.Config) {
	serviceCatalogClient := GetClient(config)
	log.Infof("Creation of the secret containing the parameters of the service will be created ...")
	bind(serviceCatalogClient,NS,"",EXTERNAL_ID,INSTANCE_NAME,INSTANCE_NAME,nil,nil)
}

// Bind an instance to a secret.
func bind(scc *servicecatalogclienset.ServicecatalogV1beta1Client, namespace, bindingName, instanceName, externalID, secretName string,
	params interface{}, secrets map[string]string) error {

	// Manually defaulting the name of the binding
	// I'm not doing the same for the secret since the API handles defaulting that value.
	if bindingName == "" {
		bindingName = instanceName
	}

	_, err := scc.ServiceBindings(namespace).Create(
		&scv1beta1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bindingName,
				Namespace: namespace,
			},
			Spec: scv1beta1.ServiceBindingSpec{
				ExternalID: externalID,
				ServiceInstanceRef: scv1beta1.LocalObjectReference{
					Name: instanceName,
				},
				SecretName:     secretName,
				Parameters:     BuildParameters(params),
				ParametersFrom: BuildParametersFrom(secrets),
			},
		})

	if err != nil {
		return errors.Wrap(err, "binding is failing")
	}
	log.Infof("Binding instance created")
	return nil
}