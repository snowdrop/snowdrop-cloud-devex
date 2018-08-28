package catalog

import (
	appsv1 "github.com/openshift/api/apps/v1"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	corev1 "k8s.io/api/core/v1"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/snowdrop/k8s-supervisor/pkg/buildpack/types"
	"github.com/snowdrop/k8s-supervisor/pkg/common/oc"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func Bind(config *restclient.Config, application types.Application) {
	serviceCatalogClient := GetClient(config)

	// Generate UUID otherwise the binding's creation will fail if we use the same id as the instanceName, bindingName
	UUID := string(uuid.NewUUID())

	log.Infof("Let's generate a secret containing the parameters to be used by the application")
	bind(serviceCatalogClient, application.Namespace, BINDING_NAME, INSTANCE_NAME, UUID, SECRET_NAME, nil, nil)
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

//Mount the secret as EnvFrom to the DeploymentConfig of the Application
func MountSecretAsEnvFrom(config *restclient.Config, application types.Application, secretName string) error {

	// Retrieve the DeploymentConfig
	deploymentConfigV1client := getAppsClient(config)
	deploymentConfigs := deploymentConfigV1client.DeploymentConfigs(application.Namespace)

	var dc *appsv1.DeploymentConfig
	var err error
	if oc.Exists("dc", application.Name) {
		dc, err = deploymentConfigs.Get(application.Name, metav1.GetOptions{})
		log.Infof("'%s' DeploymentConfig exists, got it", application.Name)
	}
	if err != nil {
		log.Fatalf("DeploymentConfig does not exist : %s", err.Error())
	}

	// Add the Secret as EnvVar to the container
	appcontainer := dc.Spec.Template.Spec.Containers[0]
	appcontainer.EnvFrom = append(appcontainer.EnvFrom, addSecretAsEnvFromSource(secretName))

	// Update the DeploymentConfig
	_, errUpdate := deploymentConfigs.Update(dc)
	if errUpdate != nil {
		log.Fatalf("DeploymentConfig not updated : %s", errUpdate.Error())
	}

	// Redeploy it
	return nil
}

func addSecretAsEnvFromSource(secretName string) corev1.EnvFromSource {
	return corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
		},
	}
}

func getAppsClient(config *restclient.Config) *appsocpv1.AppsV1Client {
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		log.Fatalf("Can't get DeploymentConfig Clientset: %s", err.Error())
	}
	return deploymentConfigV1client
}
