package main

// CMDS="echo:/var/lib/supervisord/conf/echo.sh;run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble" go run main.go

import (
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"strings"
	"os/exec"
	"fmt"
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

	// Write template result to file
	error := t.Execute(outFile, m)
	if error != nil {
		log.Fatal(error)
	}

	// // Launch Supervisord ....
	// cmd := exec.Command("/opt/supervisord/bin/supervisord", "-c", "/opt/supervisord/conf/supervisor.conf")
	// // Combine stdout and stderr
	// printCommand(cmd)
	// output, err := cmd.CombinedOutput()
	// printError(err)
	// printOutput(output)
}

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}
