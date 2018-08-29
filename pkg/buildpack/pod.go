package buildpack

import (
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/snowdrop/k8s-supervisor/pkg/buildpack/types"
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

	const timeoutInSeconds = 10
	duration := timeoutInSeconds * time.Second
	select {
	case val := <-w.ResultChan():
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
	case <-time.After(duration):
		bytes, e := json.Marshal(selector)
		if e != nil {
			return nil, errors.Errorf("Couldn't marshall pod selector to JSON: %s", e)
		}
		return nil, errors.Errorf("Waited %s but couldn't find pod matching '%s' selector", duration, string(bytes))
	}

	bytes, e := json.Marshal(selector)
	if e != nil {
		return nil, errors.Errorf("Couldn't marshall pod selector to JSON in unknown error code-path. JSON error is: %s", e)
	}
	return nil, errors.Errorf("Unknown error while waiting for pod matching '%s' selector", string(bytes))
}

func podSelector(application types.Application) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "app=" + application.Name,
	}
}
