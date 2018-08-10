package scaffold

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr"
	log "github.com/sirupsen/logrus"
)

const (
	trimPrefix   = "tmpl/"
	templateDir  = "./tmpl/"
	outDir       = "/generated/"
	javaPattern  = "*.java"
	pomPattern   = "pom.xml"
)

var (
	files       []string
	templates   = make(map[string]template.Template)
	box 		= packr.NewBox(templateDir)
)

type Project struct {
	GroupId string
	ArtifactId string
	Version string
	PackageName string

	SnowdropBomVersion string
	SpringVersion string
}

func GenerateProjectFiles(p Project) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	CollectBoxTemplates()
	ParseTemplates(currentDir,"/",p)
	fmt.Println("Done !!")
}

func CollectBoxTemplates() {
	log.Infof("List of files :",box.List())
	for _, tmpl:= range box.List() {
		log.Debug("File : " + tmpl)

		t := template.New(tmpl)
		t, _ = t.Parse(box.String(tmpl))
		templates[tmpl] = *t
	}
}
func collectTemplates(dir string, pattern string) {
	filepath.Walk(dir, visitFile(pattern))

	for _, f := range files {
		fmt.Println("File : " + f)
		b, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		name := strings.TrimPrefix(f, trimPrefix)
		log.Info("Template File Name :", name)

		t := template.New(name)
		t, _ = t.Parse(string(b))
		templates[name] = *t
	}
}
func visitFile(pattern string) filepath.WalkFunc {
	return func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil   // but continue walking elsewhere
		}
		if fi.IsDir() {
			return nil // not a file.  ignore.
		}
		matched, err := filepath.Match(pattern, fi.Name())
		if err != nil {
			fmt.Println(err) // malformed pattern
			return err   // this is fatal.
		}
		if matched {
			files = append(files,fp)
		}
		return nil
	}
}
func ParseTemplates(dir string, outDir string, p Project) {
	for _, t := range templates {
		log.Debug("##### Template : ", t.Name())
		var b bytes.Buffer
		err := t.Execute(&b, p)
		if err != nil {
			fmt.Println("There was an error:", err.Error())
		}
		log.Debug("##### Generated : ", b.String())
		os.MkdirAll(dir + outDir + path.Dir(t.Name()), os.ModePerm);
		err = ioutil.WriteFile(dir + outDir + t.Name(), b.Bytes(),0644)
		if err != nil {
			fmt.Println("There was an error:", err.Error())
		}
	}
}
func packageDirectory() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("No caller information")
	}
	return path.Dir(filename)
}
