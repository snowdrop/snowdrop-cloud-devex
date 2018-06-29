package main

// CMDS="echo:/var/lib/supervisord/conf/echo.sh;run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble" go run main.go

import (
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"strings"
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
	// Read Supervisord.tmpl file
	currentDir, err := os.Getwd()
	f, err := os.Open(currentDir + "/" + templateFile)
	if err != nil {
		panic(err)
	}
	// Close file on exit
	defer f.Close()

	// Read file content
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	// Recuperate ENV vars and split them / command
	m := make(map[string][]Program)
	if cmdsEnv := os.Getenv("CMDS"); cmdsEnv != "" {
		cmds := strings.Split(cmdsEnv,";")
		for i := range cmds {
			cmd := strings.Split(cmds[i], ":")
			p := Program{cmd[0], cmd[1]}
			m["cmd-" + string(i)] = append(m["cmd-" + string(i)],p)
		}
	} else {
		panic("No commands provided !")
	}

	// Create a template
	t := template.New("Supervisord template")
	t, _ = t.Parse(string(data)) //parse it

	// Open a new file to save the result
	outFile, err := os.OpenFile(
		outFile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	// Write template result to the supervisord.conf
	error := t.Execute(outFile, m)
	if error != nil {
		log.Fatal(error)
	}
}