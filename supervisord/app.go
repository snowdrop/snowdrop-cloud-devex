package main

import (
	"os"
	"text/template"
	"io/ioutil"
	"log"
)

const (
	templateFile = "conf/supervisord.tmpl"
	outFile      = "conf/supervisor.conf"
)

type Program struct {
	Name string
	Command string
}

func main() {
	f, err := os.Open(templateFile)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	t := template.New("Supervisord template") //create a new template
	t, _ = t.Parse(string(data)) //parse it
	p := map[string][]Program{
		"cmd-1": {{ "echo", "/var/lib/supervisord/conf/echo.sh"}},
		"cmd-2": {{"run-java", "/usr/local/s2i/run"}},
		"cmd-3": {{"compile-java", "/usr/local/s2i/assemble"}},
	} //define an instance with required field

	// Open a new file for writing only
	outFile, err := os.OpenFile(
		outFile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	// Write bytes to file
	error := t.Execute(outFile, p)
	if error != nil {
		log.Fatal(error)
	}
}
