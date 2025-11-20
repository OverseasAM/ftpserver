// ftpserver allows to create your own FTP(S) server
package main

import (
	"flag"
	"os"

	gkwrap "github.com/fclairamb/go-log/gokit"
	"github.com/kardianos/service"
)

func main() {
	// Arguments vars
	var confFile string
	var onlyConf bool
	var svcAction string

	// Parsing arguments
	flag.StringVar(&confFile, "conf", "", "Configuration file")
	flag.BoolVar(&onlyConf, "conf-only", false, "Only create the conf")
	flag.StringVar(&svcAction, "service", "", "Control the system service (install, uninstall, start, stop, restart)")
	flag.Parse()

	// Setting up the logger
	logger := gkwrap.New()

	// Service configuration
	svcConfig := &service.Config{
		Name:        "ftpserver",
		DisplayName: "FTP Server",
		Description: "FTP/FTPS server with multiple storage backend support",
	}

	// Add config file argument if specified
	if confFile != "" {
		svcConfig.Arguments = []string{"-conf", confFile}
	}

	prg := &program{
		confFile: confFile,
		onlyConf: onlyConf,
		logger:   logger,
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		logger.Error("Failed to create service", "err", err)
		os.Exit(1)
	}

	// Handle service control commands
	if svcAction != "" {
		err := service.Control(s, svcAction)
		if err != nil {
			logger.Error("Service control action failed", "action", svcAction, "err", err)
			os.Exit(1)
		}
		logger.Info("Service control action completed", "action", svcAction)
		return
	}

	// Run the service
	err = s.Run()
	if err != nil {
		logger.Error("Service run failed", "err", err)
		os.Exit(1)
	}
}
