package buildpack

import (
	"text/template"
	"io/ioutil"
	log "github.com/sirupsen/logrus"

	"github.com/cmoulliard/k8s-supervisor/pkg/buildpack/types"
	"bytes"
	"fmt"
	"runtime"
	"path"
)

var (
	templateNames = []string{"imagestream","route","service"}
	templateBuilders = make(map[string]template.Template)
)

const (
	builderpath = "tmpl/java/"
)

func init() {
	// Fill an array with our Builder's text/template
	for tmpl := range templateNames {
		buildPackDir := packageDirectory()
		// Create Template and parse it
		tfile, errFile := ioutil.ReadFile( buildPackDir + "/" + builderpath + templateNames[tmpl])
		log.Debug("Template File :",tfile)
		if errFile != nil {
			log.Error("Err is ",errFile.Error())
		}

		t := template.New(templateNames[tmpl])
		t, _ = t.Parse(string(tfile))
		templateBuilders[templateNames[tmpl]] = *t
	}
}

// Parse the file's template using the Application struct
func ParseTemplate(tmpl string, cfg types.Application) bytes.Buffer {
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

func packageDirectory() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return path.Dir(filename)
}

