package runner

import (
	"fmt"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/octoblu/go-meshblu-connector-ignition/connector"
	"github.com/octoblu/go-meshblu-connector-ignition/meshblu"
	"github.com/octoblu/go-meshblu-connector-ignition/status"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"
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
	prg, err := NewProgram(client.config)
	if err != nil {
		return err
	}

	srvConfig := &service.Config{
		Name:        client.config.ServiceName,
		DisplayName: client.config.DisplayName,
		Description: client.config.Description,
	}

	srv, err := service.New(prg, srvConfig)
	if err != nil {
		return err
	}

	prg.srv = srv

	errs := make(chan error, 5)
	logger, err := srv.Logger(errs)
	if err != nil {
		return err
	}
	prg.logger = logger

	go func() {
		for {
			err := <-errs
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	meshbluConfigPath := filepath.Join(client.config.Dir, "meshblu.json")
	meshbluClient, uuid, err := meshblu.NewClient(meshbluConfigPath)
	if err != nil {
		return err
	}
	connectorClient, err := connector.New(meshbluClient, uuid, client.config.Tag)
	if err != nil {
		return err
	}
	err = connectorClient.Fetch()
	if err != nil {
		return err
	}
	prg.connector = connectorClient

	status, err := status.New(meshbluClient, connectorClient.StatusUUID())
	if err != nil {
		return err
	}
	err = status.ResetErrors()
	logger.Info("Resetting errors on status device")
	prg.status = status

	githubSlug := prg.config.GithubSlug
	connectorName := prg.config.ConnectorName
	dir := prg.config.Dir
	uc, err := updateconnector.New(githubSlug, connectorName, dir, nil)
	if err != nil {
		return err
	}
	prg.uc = uc
	client.prg = prg

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
