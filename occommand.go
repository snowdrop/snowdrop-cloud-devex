package main

import (
	"os/exec"
	log "github.com/sirupsen/logrus"
	"os"
)

var client struct {
	ocpath string
	pwd    string
}

type OcCommand struct {
	args   []string
	data   *string
	format string
}

func getOcClient() string {
	// Search for oc client
	ocpath, err := exec.LookPath("oc")
	if err != nil {
		log.Error("Can't find oc client")
	}
	return ocpath
}

func init() {
	client.ocpath = getOcClient()
	client.pwd, _ = os.Getwd()
}

func ExecOcCommand(command OcCommand) {
	cmd := exec.Command(client.ocpath, command.args...)
	output, errCmd := cmd.CombinedOutput()
	if errCmd != nil {
		log.Debug("Error running command")
	}
	log.Info(string(output))
}