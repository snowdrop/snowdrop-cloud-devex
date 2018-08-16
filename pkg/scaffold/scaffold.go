package scaffold

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/shurcooL/httpfs/vfsutil"
	log "github.com/sirupsen/logrus"

	tmpl "github.com/snowdrop/k8s-supervisor/pkg/template"
)

const (
	configDirName    = "config"
	configYamlName   = "starters.yaml"
	dummyDirName     = "dummy"
)

var (
	templateFiles            []string
	config                   Config

	assetsJavaTemplates      = tmpl.Assets
	templates                = make(map[string]template.Template)
)

type Project struct {
	GroupId            string
	ArtifactId         string
	Version            string
	PackageName        string
	Dependencies	   []string
	OutDir             string

	SnowdropBomVersion string
	SpringVersion      string
	Modules            []Module

	UrlService  	   string
}

type Config struct {
	Modules      []Module
}

type Module struct {
	Name	     string
	Description  string
	Starters     []Starter
}

type Starter struct {
	GroupId	     string
	ArtifactId	 string
	Scope	     string
}

func ParseStartersConfigFile(pathTemplateDir string) {
	if pathTemplateDir == "" {
		pathTemplateDir = "../scaffold"
	}
	startersPath := strings.Join([]string{pathTemplateDir, configDirName, configYamlName},"/")
	log.Infof("Parsing Starters's Config at %s", startersPath)

	// Read file and parse it to create a Config's type
	if _, err := os.Stat(startersPath); err == nil {
		source, err := ioutil.ReadFile(startersPath)
		if err != nil {
			log.Fatal(err.Error())
		}

		err = yaml.Unmarshal(source, &config)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		log.Fatal("No Starters's config file detected !!!")
	}

	if log.GetLevel() == log.DebugLevel {
		log.Debug("Starters's config")
		log.Debug("--------------------")
		s, _ := json.Marshal(&config)
		log.Debug(string(s))
	}
}

func CollectVfsTemplates(t string) {

	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		log.Debug("Path of the file to be added as template : " + path)
		templateFiles = append(templateFiles,path)
		return nil
	}

	errW := vfsutil.Walk(assetsJavaTemplates, t, walkFn)
	if errW != nil {
		panic(errW)
	}

	for i := range templateFiles {
		log.Info("File template : " + templateFiles[i])

		// Create a new Template using the File name as key and add it to the array
		t := template.New(templateFiles[i])

		// Read Template's content
		data, err := vfsutil.ReadFile(assetsJavaTemplates,templateFiles[i])
		if err != nil {
			log.Error(err)
		}
		t, err = t.Parse(bytes.NewBuffer(data).String())
		if err != nil {
			log.Error(err)
		}
		templates[templateFiles[i]] = *t
	}
}

func ParseTemplates(dir string, outDir string, project Project) {
	for _, t := range templates {

		log.Infof("Template : %s", t.Name())
		var b bytes.Buffer

		// Enrich project with starters dependencies if they exist
		if strings.Contains(t.Name(),"pom.xml") {
			if project.Dependencies != nil {
				project = convertDependencyToModule(project.Dependencies, config.Modules, project)
			}
			log.Infof("Project enriched %+v ",project)
		}

		// Use template to generate the content
		err := t.Execute(&b, project)
		if err != nil {
			log.Error(err.Error())
		}

		// Convert Path
		tFileName := t.Name()
		pathF := strings.Join([]string{dir,outDir,path.Dir(tFileName)},"/")
		log.Debugf("## Path : %s",pathF)
		pathConverted := strings.Replace(pathF,dummyDirName,convertPackageToPath(project.PackageName),-1)
		log.Debugf("Path converted: ",pathF)

		// Convert FileName
		fileName := strings.Join([]string{dir,outDir,tFileName},"/")
		log.Debugf("## File name : %s",fileName)
		fileNameConverted := strings.Replace(fileName,dummyDirName,convertPackageToPath(project.PackageName),-1)
		log.Debugf("File name converted : ",fileNameConverted)

		// Create missing folders
		log.Debugf("Path to generated file : ",pathConverted)
		os.MkdirAll(pathConverted, os.ModePerm)

		// Content generated
		log.Debugf("Content generated : %s",b.Bytes())

		err = ioutil.WriteFile(fileNameConverted, b.Bytes(),0644)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func convertDependencyToModule(deps []string, modules []Module, p Project) Project {
	for _, dep := range deps {
		for _, module := range modules {
			if module.Name == dep {
				log.Infof("Match found for dep %s and starters %+v ",dep,module)
				p.Modules = append(p.Modules,module)
			}
		}
	}
	return p
}

func convertPackageToPath(p string) string {
	c := strings.Replace(p,".","/",-1)
	log.Debugf("Converted path : ",c)
	return c
}
