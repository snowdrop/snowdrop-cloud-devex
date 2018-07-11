package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	restclient "k8s.io/client-go/rest"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	imageclientsetv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"

	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	"io/ioutil"
	"os"
	"fmt"
	"text/template"
	"bytes"
	"k8s.io/apimachinery/pkg/util/intstr"
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

	appImagename         = "spring-boot-http"
	supervisordimagename = "copy-supervisord"

	builderpath          = "/builder/java/"
)

type Application struct {
	Name string
	Port int
	Image Image
}

type Image struct {
	Name string
	Repo string
}

var appConfig = Application{}

func main() {
	flag.Parse()

	log.Info("[Step 0] - Parse Application's Config")
	filename := os.Args[2]

	pwd, _ := os.Getwd()
	source, err := ioutil.ReadFile(pwd+"/"+filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &appConfig)
	if err != nil {
		panic(err)
	}
	log.Debug("Application's config")
	log.Debug("--------------------")
	log.Debug("Name : ", appConfig.Name)
	log.Debug("Port : ", appConfig.Port)

	log.Info("[Step 1] - Create Kube Client & Clientset")

	// Build kube config using kube config folder on the developer's machine
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}


	log.Info("[Step 2] - Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
	//createImageStreams(cfg)
	createImageStreamTemplate(cfg)

    log.Info("[Step 3] - Create DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
	dc := createDeploymentConfig(cfg)

	// log.Info("[Step 4] - Create Service and route")
	// createService(clientset, dc)
	// createRoute(cfg)

	clientset, errclientset := kubernetes.NewForConfig(cfg)
	if errclientset != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", errclientset.Error())
	}

	log.Info("[Step 4] - Create Service and route using Templates")
	createServiceTemplate(clientset, dc)
	createRouteTemplate(cfg)
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func createImageStreamTemplate(config *restclient.Config) {
	imageClient, err := imageclientsetv1.NewForConfig(config)
	if err != nil {
	}

	images := []Image{
		{
			Name: appImagename,
			Repo: "quay.io/snowdrop/spring-boot-s2i",
		},
		{
			Name: supervisordimagename,
			Repo: "quay.io/snowdrop/supervisord",
		},
	}

	for _, img := range images {

		cfg := &appConfig
		cfg.Image = img

		// Parse ImageStream Template
		var b = parseTemplate("imagestream_tmpl", cfg)

		// Create ImageStream struct using the generated ImageStream string
		img := imagev1.ImageStream{}
		errYamlParsing := yaml.Unmarshal(b.Bytes(), &img)
		if errYamlParsing != nil {
			panic(errYamlParsing)
		}

		_, errImages := imageClient.ImageStreams(namespace).Create(&img)
		if errImages != nil {
			glog.Fatal("Unable to create ImageStream for %s", errImages.Error())
		}
	}

}

func createServiceTemplate(clientset *kubernetes.Clientset, dc *appsv1.DeploymentConfig) {
	// Parse Service Template
	var b = parseTemplate("service_tmpl", &appConfig)

	// Create Service struct using the generated Service string
	svc := corev1.Service{}
	errYamlParsing := yaml.Unmarshal(b.Bytes(), &svc)
	if errYamlParsing != nil {
		panic(errYamlParsing)
	}

	// Create the Service
	_, errService := clientset.CoreV1().Services(namespace).Create(&svc)
	if errService != nil {
		glog.Fatal("Unable to create Service for %s", errService.Error())
	}
}

func createRouteTemplate(config *restclient.Config) {
	routeclientset, errrouteclientset := routeclientset.NewForConfig(config)
	if errrouteclientset != nil {
		glog.Fatal("error creating routeclientset", errrouteclientset.Error())
	}

	// Parse Route Template
	var b = parseTemplate("route_tmpl", &appConfig)

	// Create Route struct using the generated Route string
	route := routev1.Route{}
	errYamlParsing := yaml.Unmarshal(b.Bytes(), &route)
	if errYamlParsing != nil {
		panic(errYamlParsing)
	}

	// Create the route ...
	_, errRoute := routeclientset.Routes(namespace).Create(&route)
	if errRoute != nil {
		glog.Fatal("error creating route", errRoute.Error())
	}

}

func parseTemplate(tmpl string, cfg *Application) bytes.Buffer {
	// Create Template and parse it
	pwd, _ := os.Getwd()
	tfile, errFile := ioutil.ReadFile(pwd+builderpath+"/"+tmpl)
	if errFile != nil {
		fmt.Println("Err is ",errFile.Error())
	}

	var b bytes.Buffer
	t := template.New(tmpl)
	t, _ = t.Parse(string(tfile))
	err := t.Execute(&b, cfg)
	if err != nil {
		fmt.Println("There was an error:", err.Error())
	}
	log.Debug("Generated from ",tmpl,":",b.String())
	return b
}

func createDeploymentConfig(config *restclient.Config) *appsv1.DeploymentConfig {
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Can't get DeploymentConfig Clientset: %s", err.Error())
	}

	dc, errCreate := deploymentConfigV1client.DeploymentConfigs(namespace).Create(javaDeploymentConfig())
	if errCreate != nil {
		glog.Fatalf("DeploymentConfig not created: %s", errCreate.Error())
	}

	return dc
}

func javaDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: appConfig.Name,
			Labels: map[string]string{
				"app":              appConfig.Name,
				"io.openshift.odo": "inject-supervisord",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app":              appConfig.Name,
				"deploymentconfig": appConfig.Name,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: appConfig.Name,
					Labels: map[string]string{
						"app":              appConfig.Name,
						"deploymentconfig": appConfig.Name,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{*supervisordInitContainer()},
					Containers: []corev1.Container{
						{
							Image: appImagename + ":latest",
							Name:  appConfig.Name,
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
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							supervisordimagename,
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: supervisordimagename + ":latest",
						},
					},
				},
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							appConfig.Name,
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

func supervisordInitContainer() *corev1.Container {
	return &corev1.Container{
		Name:    "copy-supervisord",
		Image:   supervisordimagename + ":latest",
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

func createImageStreams(config *restclient.Config) {
	imageClient, err := imageclientsetv1.NewForConfig(config)
	if err != nil {
	}

	for _, v := range *imageStreams() {
		_, errImages := imageClient.ImageStreams(namespace).Create(&v)
		if errImages != nil {
		}
	}

}

func imageStreams() *[]imagev1.ImageStream {
	return &[]imagev1.ImageStream{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: appImagename,
				Labels: map[string]string{
					"app": appConfig.Name,
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
				Name: supervisordimagename,
				Labels: map[string]string{
					"app": appConfig.Name,
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

func createService(clientset *kubernetes.Clientset, dc *appsv1.DeploymentConfig) {
	// generate and create Service
	var svcPorts []corev1.ServicePort
	for _, containerPort := range dc.Spec.Template.Spec.Containers[0].Ports {
		svcPort := corev1.ServicePort{

			Name:       containerPort.Name,
			Port:       containerPort.ContainerPort,
			Protocol:   containerPort.Protocol,
			TargetPort: intstr.FromInt(int(containerPort.ContainerPort)),
		}
		svcPorts = append(svcPorts, svcPort)
	}
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: appConfig.Name,
			Labels: map[string]string{
				"app": appConfig.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"app":              appConfig.Name,
				"deploymentconfig": appConfig.Name,
			},
		},
	}

	_, errService := clientset.CoreV1().Services(namespace).Create(&svc)
	if errService != nil {
		glog.Fatal("unable to create Service for %s", appConfig.Name)
	}
}

// CreateRoute creates a route object for the given service and with the given
// labels
func createRoute(config *restclient.Config) {
	routeclientset, errrouteclientset := routeclientset.NewForConfig(config)
	if errrouteclientset != nil {
		glog.Fatal("error creating routeclientset", errrouteclientset.Error())
	}
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:   appConfig.Name,
			Labels: map[string]string{
				"app": appConfig.Name,
			},

		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: appConfig.Name,
			},
		},
	}
	_, errRoute := routeclientset.Routes(namespace).Create(route)
	if errRoute != nil {
		glog.Fatal("error creating route", errRoute.Error())
	}

}
