package main

import (
	"github.com/snowdrop/k8s-supervisor/cmd"
	"github.com/snowdrop/k8s-supervisor/pkg/common/logger"
)

var (
	// VERSION is set during build
	VERSION string
)

func main() {
	// Enable Debug if env var is defined
	logger.EnableLogLevelDebug()

	// Call commands
	cmd.Execute(VERSION)
}