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
	// Read Supervisord Template file
	currentDir, err := os.Getwd()
	f, err := os.Open(currentDir + "/" + templateFile)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer f.Close()

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

	//create a supervisord.conf template
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

	// Write bytes to file
	error := t.Execute(outFile, m)
	if error != nil {
		log.Fatal(error)
	}

	// app := "echo"

	// arg0 := "-e"
	// arg1 := "Hello world"
	// arg2 := "\n\tfrom"
	// arg3 := "golang"

	// cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	// stdout, err := cmd.Output()

	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }

	// print(string(stdout))
}
