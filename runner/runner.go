package runner

import (
	"fmt"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/octoblu/go-meshblu-connector-ignition/device"
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
	client.prg = NewProgram(client.config)

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
				fmt.Println(err)
			}
		}
	}()

	meshbluConfigPath := filepath.Join(client.config.Dir, "meshblu.json")
	deviceClient, err := device.New(meshbluConfigPath, client.config.Tag)
	if err != nil {
		return err
	}
	err = deviceClient.Update()
	if err != nil {
		return err
	}
	client.prg.device = deviceClient

	client.prg.uc = NewUpdateConnector(client.prg)

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
