package catalog

import (
	restclient "k8s.io/client-go/rest"

	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	log "github.com/sirupsen/logrus"

	"bytes"
	"encoding/json"
)

func GetClient(config *restclient.Config) *servicecatalogclienset.ServicecatalogV1beta1Client {
	serviceCatalogV1Client, err := servicecatalogclienset.NewForConfig(config)
	if err != nil {
		log.Fatal("error creating service catalog Clientset", err.Error())
	}
	return serviceCatalogV1Client
}

func PrettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

