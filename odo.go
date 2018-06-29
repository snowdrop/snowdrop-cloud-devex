package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	// "github.com/cmoulliard/k8s-odo-supervisor/pkg/signals"

	restclient "k8s.io/client-go/rest"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	masterURL  string
	kubeconfig string

	filter = metav1.ListOptions{
		LabelSelector: "io.openshift.odo=inject-supervisord",
	}
)

const (
	namespace = "k8s-supervisord"
	supervisordimage = "172.30.1.1:5000/k8s-supervisord/copy-supervisord:1.0"
)

/*func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	// Build kube config using kube config folder on the developer's machine
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	fmt.Println("Kube config parsed correctly")

	controller := NewController(kubeClient,cfg, namespace)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}

}*/

func main() {
	flag.Parse()

	// Build kube config using kube config folder on the developer's machine
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	_, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	fmt.Println("Fetching about DC to be injected")
	findDeploymentconfig(cfg)
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func findDeploymentconfig(config *restclient.Config) {
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		glog.Error("")
	}

	deploymentList, err := deploymentConfigV1client.DeploymentConfigs(namespace).List(filter)
	fmt.Printf("Listing deployments in namespace %s: \n", namespace)
	if err != nil {
		glog.Error("Error to get Deployment Config !")
	}
	for _, d := range deploymentList.Items {
		fmt.Printf("%s\n", d.Name)

		// TODO : Check if deploymentConfig contains the initContainer for copy-supervisord. If no, we paytch it

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			dc, err := deploymentConfigV1client.DeploymentConfigs(namespace).Get(d.Name, metav1.GetOptions{})
			if err != nil {
				glog.Error("Error to get the Deployment Config %s. Error is : %s\n", dc.Name, err)
			}

			dc.Spec.Template.Spec.InitContainers = []corev1.Container { *supervisordInitContainer() }
			_, updatedErr := deploymentConfigV1client.DeploymentConfigs(namespace).Update(dc)
			return updatedErr
		})
		if retryErr != nil {
			panic(fmt.Errorf("Update failed: %v", retryErr))
		}
		fmt.Println("Updated deployment...")
		//fmt.Printf("Raw printout of the dc %+v\n", d)
	}
}

func supervisordInitContainer() *corev1.Container {
	return &corev1.Container{
		Name:    "copy-supervisord",
		Image:   supervisordimage,
		Command: []string{"/bin/busybox"},
		Args:    []string{"/usr/bin/cp","-r", "/opt/supervisord", "/var/lib/"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/var/lib/supervisord",
			},
		},
	}
}
