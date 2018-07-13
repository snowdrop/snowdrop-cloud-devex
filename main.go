package main

import (
	"github.com/cmoulliard/k8s-supervisor/cmd"
	"github.com/cmoulliard/k8s-supervisor/pkg/common/logger"
)

func main() {
	// Enable Debug if env var is defined
	logger.EnableLogLevelDebug()

	// Call commands
	cmd.Execute()
}

