package main

import (
	"github.com/snowdrop/k8s-supervisor/cmd"
	"github.com/snowdrop/k8s-supervisor/pkg/common/logger"
)

func main() {
	// Enable Debug if env var is defined
	logger.EnableLogLevelDebug()

	// Call commands
	cmd.Execute()
}