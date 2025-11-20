package main

import (
	"github.com/fclairamb/go-log"
	"github.com/kardianos/service"
)

// program implements the service.Interface
type program struct {
	confFile string
	onlyConf bool
	logger   log.Logger
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
	err := runServer(p.confFile, p.onlyConf, p.logger)
	if err != nil {
		p.logger.Error("Server exited with error", "err", err)
	}
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	p.logger.Info("Stopping FTP server service")
	stop()
	return nil
}
