package buildpack_test

import (
	"text/template"
	"path"
	"runtime"
	"testing"
	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"bytes"
)

func TestServiceTemplate(t *testing.T) {

	builderpath := "tmpl/java/"

	const service =
`apiVersion: v1
kind: Service
metadata:
  name: service-test
  labels:
    app: service-test
    name: service-test
spec:
  ports:
  - port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    app: spring-boot-supervisord
    deploymentconfig: spring-boot-supervisord`

	application := types.Application{
		Name: "service-test",
		Port: 8080,
	}

	// Get package full path
	serviceFile := packageDirectory() + "/" + builderpath + "/service"
	templ, _ := template.New("service").ParseFiles(serviceFile)

	var b bytes.Buffer
	templ.Execute(&b,application)
	r := b.String()

	if service != r {
		t.Errorf("Result was incorrect, got: " +
			"%s" +
			", want: " +
			"%s", r, service)
	}
}

func packageDirectory() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return path.Dir(filename)
}