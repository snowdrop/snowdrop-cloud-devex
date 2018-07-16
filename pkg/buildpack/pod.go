package buildpack

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
)

// WaitAndGetPod block and waits until pod matching selector is in in Running state
func WaitAndGetPod(c *kubernetes.Clientset, application types.Application) (*corev1.Pod, error) {

	selector := podSelector(application)
	log.Debugf("Waiting for %s pod", selector)

	w, err := c.CoreV1().Pods(application.Namespace).Watch(selector)
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
	return nil, errors.Errorf("unknown error while waiting for pod matchin '%s' selector", podSelector)
}

func podSelector(application types.Application) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "app=" + application.Name,
	}
}
