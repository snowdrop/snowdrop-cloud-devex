package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/config"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

var (
	namespace string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sb",
	Short: "sb (ODO's Spring Boot prototype)",

	Long: `sb (ODO's Spring Boot prototype) is a prototype's project experimenting supervisord, MANIFEST's concepts`,

	Example: `    # Creating and deploying a spring Boot application
    git clone github.com/cmoulliard/k8s-supervisor && cd k8s-supervisor/spring-boot
    sb init -p namespace
    sb push
    sb compile
    sb run`,
}

func init() {
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "","Path to a kubeconfig ($HOME/.kube/config). Only required if out-of-cluster.")
	rootCmd.PersistentFlags().StringP("masterurl", "m", "","The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	// Global flag(s)
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace","n","",  "Namespace/project")
	//rootCmd.MarkPersistentFlagRequired("namespace")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		checkError(err,"Root execution")
	}
}

// checkError prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
func checkError(err error, context string, a ...interface{}) {
	if err != nil {
		log.Debugf("Error:\n%v", err)
		if context == "" {
			fmt.Println(errors.Cause(err))
		} else {
			fmt.Printf(fmt.Sprintf("%s\n", context), a...)
		}
		os.Exit(1)
	}
}

func parseManifest() types.Application {
	log.Info("Parse MANIFEST of the project if it exists")
	current, _ := os.Getwd()
	return buildpack.ParseManifest(current + "/MANIFEST")
}

func getK8Config(cmd cobra.Command) config.Kube {
	log.Info("Get K8s config file")
	var kubeCfg = config.Kube{}
	if cmd.Flag("kubeconfig").Value.String() == "" {
		kubeCfg.Config = config.HomeKubePath()
	} else {
		kubeCfg.Config = cmd.Flag("kubeconfig").Value.String()
	}
	log.Debug("Kubeconfig : ",kubeCfg)
	return kubeCfg
}

// Create Kube ClientSet
func createClientSet(kubeCfg config.Kube) *kubernetes.Clientset {
	kubeRestClient := createKubeRestconfig(kubeCfg)
	log.Info("Create Kube Clientset")
	clientset, errclientset := kubernetes.NewForConfig(kubeRestClient)
	if errclientset != nil {
		log.Fatalf("Error building kubernetes clientset: %s", errclientset.Error())
	}
	return clientset
}

// Create Kube Rest's Config Client
func createKubeRestconfig(kubeCfg config.Kube) *restclient.Config {
	log.Info("Create kube Rest config client using config's file of the developer's machine")
	kubeRestClient, err := clientcmd.BuildConfigFromFlags(kubeCfg.MasterURL, kubeCfg.Config)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	return kubeRestClient
}
