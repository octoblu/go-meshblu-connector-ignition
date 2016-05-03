package runner

import (
	"log"

	"github.com/kardianos/service"
)

// Runner defines the interface to run a Cmd
type Runner interface {
	Start() error
	Shutdown() error
}

// Client defines the stucture of the client
type Client struct {
	config *Config
	prg    *Program
}

// New creates a new instance of runner
func New(config *Config) Runner {
	return &Client{config, nil}
}

// Start runs the connector
func (client *Client) Start() error {
	client.prg = &Program{
		config: client.config,
		exit:   make(chan struct{}),
	}

	srvConfig := &service.Config{
		Name:        client.config.ServiceName,
		DisplayName: client.config.DisplayName,
		Description: client.config.Description,
	}

	srv, err := service.New(client.prg, srvConfig)
	if err != nil {
		return err
	}

	client.prg.srv = srv

	errs := make(chan error, 5)
	logger, err := srv.Logger(errs)
	if err != nil {
		return err
	}
	client.prg.logger = logger

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	err = srv.Run()
	if err != nil {
		return err
	}
	return nil
}

// Shutdown will kill the connector process
func (client *Client) Shutdown() error {
	return client.prg.Stop(client.prg.srv)
}
