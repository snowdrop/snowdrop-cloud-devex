package main

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	"github.com/golang/glog"

	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	config *restclient.Config
	namespace string
}

// NewController returns a new ODO controller
func NewController(
	kubeclientset kubernetes.Interface,
	config *restclient.Config,
    namespace string) *Controller {

	controller := &Controller{
		kubeclientset: kubeclientset,
		config: config,
		namespace: namespace,
	}
	glog.Info("Controller created")

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	//defer runtime.HandleCrash()
	//defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting ODO controller")

	deploymentConfigV1client, err := appsv1.NewForConfig(c.config)
	if err != nil {
		glog.Error("")
	}

	deploymentList, err := deploymentConfigV1client.DeploymentConfigs(c.namespace).List(metav1.ListOptions{})
	fmt.Printf("Listing deployments in namespace %s: \n", c.namespace)
	if err != nil {
		panic(err)
	}
	for _, d := range deploymentList.Items {
		fmt.Printf("%s\n", d.Name)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}
