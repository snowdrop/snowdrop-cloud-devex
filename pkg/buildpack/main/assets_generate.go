package main

import (
  log "github.com/sirupsen/logrus"
  "github.com/shurcooL/vfsgen"

  "net/http"
)

func main() {
	var assets = http.Dir("tmpl")

	err := vfsgen.Generate(assets, vfsgen.Options{
		Filename:    "vfsdata.go",
		PackageName: "buildpack",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
