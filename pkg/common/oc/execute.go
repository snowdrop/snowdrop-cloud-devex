package oc

import (
	"os"
	"os/exec"
	log "github.com/sirupsen/logrus"
)

var Client struct {
	Path string
	Pwd    string
}

type Command struct {
	Args   []string
	Data   *string
	Format string
}

func getClientPath() string {
	// Search for oc client
	ocpath, err := exec.LookPath("oc")
	if err != nil {
		log.Error("Can't find oc client")
	}
	return ocpath
}

func init() {
	Client.Path   = getClientPath()
	Client.Pwd, _ = os.Getwd()
}

func ExecCommand(command Command) {
	cmd := exec.Command(Client.Path, command.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}