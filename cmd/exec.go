package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"fmt"
	"github.com/snowdrop/k8s-supervisor/pkg/common/config"
	"github.com/snowdrop/k8s-supervisor/pkg/common/oc"
	"strings"
)

var exeCmd = &cobra.Command{
	Use:   "exec [options]",
	Short: "Stop, start or restart your SpringBoot application.",
	Long:  `Stop, start or restart your SpringBoot application.`,
	Example: fmt.Sprintf("%s\n%s\n%s",
		execStartCmd.Example,
		execStopCmd.Example,
		execRestartCmd.Example),
}

var execStartCmd = newCommand("start")
var execStopCmd = newCommand("stop")
var execRestartCmd = newCommandWith("restart", func(podName string, action string) {
	oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "stop", config.RunCmdName}})
	oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "start", config.RunCmdName}})
})

func newCommand(action string) *cobra.Command {
	return newCommandWith(action, execAction)
}

func newCommandWith(action string, toExec func(podName string, action string)) *cobra.Command {
	capitalizedAction := strings.Title(action)

	return &cobra.Command{
		Use:     action,
		Short:   capitalizedAction + " your SpringBoot application.",
		Long:    capitalizedAction + ` your SpringBoot application.`,
		Example: `  sb exec ` + action,
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			log.Infof("Exec %s command called", action)

			_, pod := SetupAndWaitForPod()
			podName := pod.Name

			log.Infof("%s the Spring Boot application ...", capitalizedAction)
			toExec(podName, action)
			oc.ExecCommand(oc.Command{Args: []string{"logs", podName, "-f"}})
		},
	}
}

func execAction(podName string, action string) {
	cmdArgs := []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, action, config.RunCmdName}
	log.Debug("Command :", cmdArgs)
	oc.ExecCommand(oc.Command{Args: cmdArgs})
}

func init() {
	exeCmd.AddCommand(execStartCmd)
	exeCmd.AddCommand(execStopCmd)
	exeCmd.AddCommand(execRestartCmd)

	exeCmd.Annotations = map[string]string{"command": "exec"}
	rootCmd.AddCommand(exeCmd)
}
