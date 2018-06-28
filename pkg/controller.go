package main

import (
	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
}

// NewController returns a new ODO controller
func NewController(kubeclientset kubernetes.Interface) *Controller {

	controller := &Controller{
		kubeclientset: kubeclientset,
	}
	glog.Info("Controller created")

	return controller
}
