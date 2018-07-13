package main

import (
	"io/ioutil"
	"os"
	"fmt"
	"text/template"
	"bytes"
	"flag"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/oc"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/logger"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	restclient "k8s.io/client-go/rest"
	
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	imageclientsetv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	routeclientsetv1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"

	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"encoding/json"
)

var (
	masterURL  string
	kubeconfig string

	podSelector = metav1.ListOptions{
		LabelSelector: "app=spring-boot-supervisord",
	}

	templateNames = []string{"imagestream","route","service"}
	templateBuilders = make(map[string]template.Template)
	appConfig types.Application
)

const (
	namespace            = "k8s-supervisord"

	appImagename         = "spring-boot-http"
	version              = "1.0"
	supervisordimagename = "copy-supervisord"

	builderpath          = "/builder/java/"
)

func main() {
	flag.Parse()

	log.Info("[Step 0] - Parse Application's Config")
	filename := os.Args[2]

	pwd, _ := os.Getwd()
	source, err := ioutil.ReadFile(pwd+"/"+filename)
	if err != nil {
		panic(err)
	}

	// Create an Application with default values
	appConfig = types.NewApplication()

	err = yaml.Unmarshal(source, &appConfig)
	if err != nil {
		panic(err)
	}
	log.Debug("Application's config")
	log.Debug("--------------------")
	appFormatted, _ := json.Marshal(appConfig)
	log.Debug(string(appFormatted))

	log.Info("[Step 1] - Create Kube Client & Clientset")

	// Build kube config using kube config folder on the developer's machine
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	clientset, errclientset := kubernetes.NewForConfig(cfg)
	if errclientset != nil {
		log.Fatalf("Error building kubernetes clientset: %s", errclientset.Error())
	}

	log.Info("[Step 2] - Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
	//createImageStreams(cfg)
	createImageStreamTemplate(cfg)

    log.Info("[Step 3] - Create DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
	dc := createDeploymentConfig(cfg)

	log.Info("[Step 4] - Create Service and route using Templates")
	createServiceTemplate(clientset, dc)
	createRouteTemplate(cfg)

	log.Info("[Step 5] - Watch about Development's pod ...")
	pod, err := WaitAndGetPod(clientset,podSelector)
	if err != nil {
		log.Error("Pod watch error",err)
	}

	podName := pod.Name

	log.Info("[Step 6] - Copy files from Development projects to the pod")
	oc.ExecCommand(oc.Command{Args: []string{"cp",oc.Client.Pwd+"/spring-boot/"+"pom.xml",podName+":/tmp/src/","-c","spring-boot-supervisord"}})
	oc.ExecCommand(oc.Command{Args: []string{"cp",oc.Client.Pwd+"/spring-boot/"+"src",podName+":/tmp/src/","-c","spring-boot-supervisord"}})

	log.Info("[Step 7] - Check status of the supervisord's daemon")
	oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,"/var/lib/supervisord/bin/supervisord","ctl","status"}})

	log.Info("[Step 7] - Start compilation")
	oc.ExecCommand(oc.Command{Args: []string{"rsh",podName,"/var/lib/supervisord/bin/supervisord","ctl","start","compile-java"}})
	oc.ExecCommand(oc.Command{Args: []string{"logs",podName,"-f"}})

}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	// Enable Debug if env var is defined
	logger.EnableLogLevelDebug()

	// Fill an array with our Builder's text/template
	for tmpl := range templateNames {
		pwd, _ := os.Getwd()
		// Create Template and parse it
		tfile, errFile := ioutil.ReadFile(pwd+builderpath+"/"+templateNames[tmpl])
		if errFile != nil {
			fmt.Println("Err is ",errFile.Error())
		}

		t := template.New(templateNames[tmpl])
		t, _ = t.Parse(string(tfile))
		templateBuilders[templateNames[tmpl]] = *t
	}
}

// WaitAndGetPod block and waits until pod matching selector is in in Running state
func WaitAndGetPod(c *kubernetes.Clientset, selector metav1.ListOptions) (*corev1.Pod, error) {
	log.Debugf("Waiting for %s pod", selector)

	w, err := c.CoreV1().Pods(namespace).Watch(selector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to watch pod")
	}
	defer w.Stop()
	for {
		val, ok := <-w.ResultChan()
		if !ok {
			break
		}
		if e, ok := val.Object.(*corev1.Pod); ok {
			log.Debugf("Status of %s pod is %s", e.Name, e.Status.Phase)
			switch e.Status.Phase {
			case corev1.PodRunning:
				log.Debugf("Pod %s is running.", e.Name)
				return e, nil
			case corev1.PodFailed, corev1.PodUnknown:
				return nil, errors.Errorf("pod %s status %s", e.Name, e.Status.Phase)
			}
		}
	}
	return nil, errors.Errorf("unknown error while waiting for pod matchin '%s' selector", selector)
}

func createImageStreamTemplate(config *restclient.Config) {
	imageClient, err := imageclientsetv1.NewForConfig(config)
	if err != nil {
	}

	images := []types.Image{
		{
			Name: appImagename,
			Repo: "quay.io/snowdrop/spring-boot-s2i",
		},
		{
			Name: supervisordimagename,
			Repo: "quay.io/snowdrop/supervisord",
		},
	}

	appCfg := &appConfig
	for _, img := range images {

		appCfg.Image = img

		// Parse ImageStream Template
		var b = parseTemplate("imagestream",appCfg)

		// Create ImageStream struct using the generated ImageStream string
		img := imagev1.ImageStream{}
		errYamlParsing := yaml.Unmarshal(b.Bytes(), &img)
		if errYamlParsing != nil {
			panic(errYamlParsing)
		}

		_, errImages := imageClient.ImageStreams(namespace).Create(&img)
		if errImages != nil {
			log.Fatal("Unable to create ImageStream for %s", errImages.Error())
		}
	}

}

func createServiceTemplate(clientset *kubernetes.Clientset, dc *appsv1.DeploymentConfig) {
	// Parse Service Template
	var b = parseTemplate("service",&appConfig)

	// Create Service struct using the generated Service string
	svc := corev1.Service{}
	errYamlParsing := yaml.Unmarshal(b.Bytes(), &svc)
	if errYamlParsing != nil {
		panic(errYamlParsing)
	}

	// Create the Service
	_, errService := clientset.CoreV1().Services(namespace).Create(&svc)
	if errService != nil {
		log.Fatal("Unable to create Service for %s", errService.Error())
	}
}

func createRouteTemplate(config *restclient.Config) {
	routeclientsetv1, errrouteclientsetv1 := routeclientsetv1.NewForConfig(config)
	if errrouteclientsetv1 != nil {
		log.Fatal("error creating routeclientsetv1", errrouteclientsetv1.Error())
	}

	// Parse Route Template
	var b = parseTemplate("route", &appConfig)

	// Create Route struct using the generated Route string
	route := routev1.Route{}
	errYamlParsing := yaml.Unmarshal(b.Bytes(), &route)
	if errYamlParsing != nil {
		panic(errYamlParsing)
	}

	// Create the route ...
	_, errRoute := routeclientsetv1.Routes(namespace).Create(&route)
	if errRoute != nil {
		log.Fatal("error creating route", errRoute.Error())
	}

}

// Parse the file's template using the Application struct
func parseTemplate(tmpl string, cfg *types.Application) bytes.Buffer {
	// Create Template and parse it
	var b bytes.Buffer
	t := templateBuilders[tmpl]
	err := t.Execute(&b, cfg)
	if err != nil {
		fmt.Println("There was an error:", err.Error())
	}
	log.Debug("Generated :",b.String())
	return b
}

func createDeploymentConfig(config *restclient.Config) *appsv1.DeploymentConfig {
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		log.Fatalf("Can't get DeploymentConfig Clientset: %s", err.Error())
	}

	dc, errCreate := deploymentConfigV1client.DeploymentConfigs(namespace).Create(javaDeploymentConfig())
	if errCreate != nil {
		log.Fatalf("DeploymentConfig not created: %s", errCreate.Error())
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
									ContainerPort: appConfig.Port,
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
									Value: appImagename + "-" + version + ".jar",
								},
							},
/*							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse(appConfig.Cpu),
									corev1.ResourceMemory: resource.MustParse(appConfig.Memory),
								},
							},*/
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
				Value: "echo:/var/lib/supervisord/conf/echo.sh;run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp",
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
		log.Fatal("unable to create Service for %s", appConfig.Name)
	}
}

// CreateRoute creates a route object for the given service and with the given
// labels
func createRoute(config *restclient.Config) {
	routeclientsetv1, errrouteclientsetv1 := routeclientsetv1.NewForConfig(config)
	if errrouteclientsetv1 != nil {
		log.Fatal("error creating routeclientsetv1", errrouteclientsetv1.Error())
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
	_, errRoute := routeclientsetv1.Routes(namespace).Create(route)
	if errRoute != nil {
		log.Fatal("error creating route", errRoute.Error())
	}

}
