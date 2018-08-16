package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/shurcooL/vfsgen"
	"github.com/shurcooL/httpfs/vfsutil"
	target "github.com/snowdrop/k8s-supervisor/pkg/template"
)

var (
	templateFiles []string
)

func main() {
	generateVfsDataFile()
	//walkTemplatesTree(t)
}

func walkTemplatesTree(t string) {

	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}

		//fmt.Println(path)
		if fi.IsDir() {
			return nil
		}

		templateFiles = append(templateFiles,path)
		return nil
	}

	errW := vfsutil.Walk(target.Assets, t, walkFn)
	if errW != nil {
		panic(errW)
	}

	for i := range templateFiles {
		log.Info("Template file : ", templateFiles[i])
		b, err := vfsutil.ReadFile(target.Assets, templateFiles[i])
		log.Infof("Content : %q %v\n", string(b), err)
	}
}

func generateVfsDataFile() {
	err := vfsgen.Generate(target.Assets, vfsgen.Options{
		Filename:    "vfsdata.go",
		PackageName: "template",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
