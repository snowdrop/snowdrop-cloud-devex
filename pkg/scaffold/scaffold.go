package scaffold

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"

	"strings"
	"text/template"

	"github.com/gobuffalo/packr"
	log "github.com/sirupsen/logrus"
)

const (
	templateDirName  = "tmpl"
	dummyDirName     = "dummy"
)

var (
	files            []string
	pathtemplateDir	 string
	templates        = make(map[string]template.Template)
	box 		     packr.Box
)

type Project struct {
	GroupId            string
	ArtifactId         string
	Version            string
	PackageName        string
	OutDir             string

	SnowdropBomVersion string
	SpringVersion      string

	UrlService  	   string
}

func CollectBoxTemplates(t, pathTemplateDir string) {
	if pathTemplateDir == "" {
		pathTemplateDir = "../scaffold"
	}
	pathTempl := strings.Join([]string{pathTemplateDir, templateDirName, t},"/")
	log.Infof("Path to the template directory : %s",pathTempl)
	box = packr.NewBox(pathTempl)
	log.Infof("List of files :",box.List())

	for _, tmpl:= range box.List() {
		log.Debug("File : " + tmpl)

		t := template.New(tmpl)
		t, _ = t.Parse(box.String(tmpl))
		templates[tmpl] = *t
	}
}

func ParseTemplates(dir string, outDir string, project Project) {
	for _, t := range templates {

		log.Debugf("Template : %s", t.Name())
		var b bytes.Buffer
		err := t.Execute(&b, project)
		if err != nil {
			log.Error(err.Error())
		}
		log.Debugf("Generated : %s", b.String())

		// Convert Path
		tFileName := t.Name()
		// TODO Use filepath.Join
		path := dir + outDir + path.Dir(tFileName)
		log.Debugf("Path : ",path)
		pathConverted := strings.Replace(path,dummyDirName,convertPackageToPath(project.PackageName),-1)
		log.Debugf("Path converted: ",path)

		// convert FileName
		// TODO Use filepath.Join
		fileName := dir + outDir + tFileName
		log.Debugf("File name : ",fileName)
		fileNameConverted := strings.Replace(fileName,dummyDirName,convertPackageToPath(project.PackageName),-1)
		log.Debugf("File name converted : ",fileNameConverted)

		// Create missing folders
		log.Infof("Path to generated file : ",pathConverted)
		os.MkdirAll(pathConverted, os.ModePerm)

		err = ioutil.WriteFile(fileNameConverted, b.Bytes(),0644)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func convertPackageToPath(p string) string {
	c := strings.Replace(p,".","/",-1)
	c = "src/main/java/" + c
	log.Debugf("Converted path : ",c)
	return c
}
