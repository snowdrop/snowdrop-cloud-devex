package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	// "github.com/cmoulliard/k8s-odo-supervisor/pkg/signals"

	restclient "k8s.io/client-go/rest"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	imageclientsetv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	imagev1 "github.com/openshift/api/image/v1"
)

var (
	masterURL  string
	kubeconfig string

	filter = metav1.ListOptions{
		LabelSelector: "io.openshift.odo=inject-supervisord",
	}
)

const (
	namespace            = "k8s-supervisord"
	supervisordimage     = "172.30.1.1:5000/k8s-supervisord/copy-supervisord:1.0"
	appname              = "spring-boot-supervisord"
	appImagename         = "spring-boot-http"
	supervisordName      = "spring-boot-supervisord"
	supervisordImageName = "copy-supervisord"
)

func main() {
	flag.Parse()

	log.Info("[Step 1] - Create Kube Client & Clientset")

	// Build kube config using kube config folder on the developer's machine
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	_, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	log.Info("[Step 2] - Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
	createImageStreams(cfg)

	log.Info("[Step 3] - Create DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
	createDeploymentConfig(cfg)
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func createImageStreams(config *restclient.Config) {

	imageClient, err := imageclientsetv1.NewForConfig(config)
	if err != nil { }

	for _, v := range *imageStreams() {
		_, errImages := imageClient.ImageStreams(namespace).Create(&v)
		if errImages != nil {}
	}

}

func imageStreams() *[]imagev1.ImageStream{
	return &[]imagev1.ImageStream{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: appImagename,
				Labels: map[string]string{
					"app": appname,
				},
			},
			Spec: imagev1.ImageStreamSpec{
				LookupPolicy: imagev1.ImageLookupPolicy{
					Local: false,
				},
				Tags: []imagev1.TagReference{
					{
						Name: "latest",
						From: &corev1.ObjectReference{
							Name: "quay.io/snowdrop/spring-boot-s2i",
							Kind: "DockerImage",
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: supervisordImageName,
				Labels: map[string]string{
					"app": appname,
				},
			},
			Spec: imagev1.ImageStreamSpec{
				LookupPolicy: imagev1.ImageLookupPolicy{
					Local: false,
				},
				Tags: []imagev1.TagReference{
					{
						Name: "latest",
						From: &corev1.ObjectReference{
							Name: "quay.io/snowdrop/supervisord",
							Kind: "DockerImage",
						},
					},
				},
			},
		},
	}
}

func createDeploymentConfig(config *restclient.Config) {
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Can't get DeploymentConfig Clientset: %s", err.Error())
	}

	_, errCreate := deploymentConfigV1client.DeploymentConfigs(namespace).Create(javaDeploymentConfig())
	if errCreate != nil {
		glog.Fatalf("DeploymentConfig not created: %s", errCreate.Error())
	}
}

func javaDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: appname,
			Labels: map[string]string{
				"app":              appname,
				"io.openshift.odo": "inject-supervisord",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app":              appname,
				"deploymentconfig": appname,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: appname,
					Labels: map[string]string{
						"app":              appname,
						"deploymentconfig": appname,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{*supervisordInitContainer()},
					Containers: []corev1.Container{
						{
							Image: appImagename + ":latest",
							Name:  appname,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "JAVA_APP_DIR",
									Value: "/deployments",
								},
								{
									Name:  "JAVA_APP_JAR",
									Value: appImagename + "-1.0-exec.jar",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shared-data",
									MountPath: "/var/lib/supervisord",
								},
							},
							Command: []string{
								"/var/lib/supervisord/bin/supervisord",
							},
							Args: []string{
								"-c",
								"/var/lib/supervisord/conf/supervisor.conf",
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "shared-data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			Triggers: []appsv1.DeploymentTriggerPolicy{
				{
					Type: "ConfigChange",
				},
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							appname,
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: appImagename + ":latest",
						},
					},
				},
			},
		},
	}
}

func findDeploymentconfig(config *restclient.Config) {
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	deploymentList, err := deploymentConfigV1client.DeploymentConfigs(namespace).List(filter)
	fmt.Printf("Listing deployments in namespace %s: \n", namespace)
	if err != nil {
		glog.Error("Error to get Deployment Config !")
	}
	for _, d := range deploymentList.Items {
		fmt.Printf("%s\n", d.Name)

		// TODO : Check if deploymentConfig contains the initContainer for copy-supervisord. If no, we patch it

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			dc, err := deploymentConfigV1client.DeploymentConfigs(namespace).Get(d.Name, metav1.GetOptions{})
			if err != nil {
				glog.Error("Error to get the Deployment Config %s. Error is : %s\n", dc.Name, err)
			}

			// Add emptyDir volume shared by initContainer and Container
			dc.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: "shared-data",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			}

			// Add command to the s2i image in order to start the supervisord
			dc.Spec.Template.Spec.Containers[0].Command = append(dc.Spec.Template.Spec.Containers[0].Command, "/var/lib/supervisord/bin/supervisord")
			dc.Spec.Template.Spec.Containers[0].Args = append(dc.Spec.Template.Spec.Containers[0].Args, "-c", "/var/lib/supervisord/conf/supervisor.conf")
			dc.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
				{
					Name:      "shared-data",
					MountPath: "/var/lib/supervisord",
				},
			}

			// Inject an initcontainer containing binary go of the supervisord
			dc.Spec.Template.Spec.InitContainers = []corev1.Container{*supervisordInitContainer()}

			// Update the DeploymentConfig
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
		Args:    []string{"/usr/bin/cp", "-r", "/opt/supervisord", "/var/lib/"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/var/lib/supervisord",
			},
		},
		// TODO : The following list should be calculated based on the labels of the S2I image
		Env: []corev1.EnvVar{
			{
				Name:  "CMDS",
				Value: "echo:/var/lib/supervisord/conf/echo.sh;run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble",
			},
		},
	}
}
