package scaffold

import (
	"text/template"
	"os"
	"path"
	"runtime"
	"fmt"
	"path/filepath"
	"io/ioutil"
	"strings"
	log "github.com/sirupsen/logrus"
	"bytes"
	"archive/zip"
)

const (
	trimPrefix   = "tmpl/"
	templateDir  = "/tmpl/"
	generatedDir = "./generated/"
	javaPattern  = "*.java"
	pomPattern   = "pom.xml"
)

var (
	files []string
	templates = make(map[string]template.Template)
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
	// Generate pom.xml
	collectTemplates(packageDirectory() + templateDir, pomPattern)
	parseTemplates(p)

	// Generate Java files
	collectTemplates(packageDirectory() + templateDir, javaPattern)
    parseTemplates(p)

	PopulateZip()

	fmt.Println("Done !!")
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
		log.Debug("Template File Name :", name)

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
func parseTemplates(p Project) {
	var b bytes.Buffer
	for _, t := range templates {
		log.Debug("##### Template : ", t.Name())
		err := t.Execute(&b, p)
		if err != nil {
			fmt.Println("There was an error:", err.Error())
		}
		log.Debug("##### Generated : ", b.String())
		os.MkdirAll(generatedDir + path.Dir(t.Name()), os.ModePerm);
		err = ioutil.WriteFile(generatedDir + t.Name(), b.Bytes(),0644)
		if err != nil {
			fmt.Println("There was an error:", err.Error())
		}
	}
}

func PopulateZip() {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Add some files to the archive.
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling licence.\nWrite more examples."},
	}
	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			log.Fatal(err)
		}
	}

	// Make sure to check the error on Close.
	err := w.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func packageDirectory() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return path.Dir(filename)
}
