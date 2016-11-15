package runner

import (
	"path/filepath"

	mainlogger "github.com/azer/logger"
	"github.com/kardianos/service"
	"github.com/octoblu/go-meshblu-connector-ignition/connector"
	"github.com/octoblu/go-meshblu-connector-ignition/meshblu"
	"github.com/octoblu/go-meshblu-connector-ignition/status"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"
)

var log = mainlogger.New("runner")

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
		log.Error("Error creating new program %v", err.Error())
		return err
	}

	srvConfig := &service.Config{
		Name:        client.config.ServiceName,
		DisplayName: client.config.DisplayName,
		Description: client.config.Description,
	}

	srv, err := service.New(prg, srvConfig)
	if err != nil {
		log.Error("Error getting service %v", err.Error())
		return err
	}

	prg.srv = srv

	errs := make(chan error, 5)

	go func() {
		for {
			err = <-errs
			if err != nil {
				log.Error("Error captured in channel %v", err.Error())
			}
		}
	}()

	meshbluConfigPath := filepath.Join(client.config.Dir, "meshblu.json")
	meshbluClient, uuid, err := meshblu.NewClient(meshbluConfigPath)
	if err != nil {
		log.Error("Error getting meshblu client %v", err.Error())
		return err
	}
	connectorClient, err := connector.New(meshbluClient, uuid, client.config.Tag)
	if err != nil {
		log.Error("Error connector client new %v", err.Error())
		return err
	}
	err = connectorClient.Fetch()
	if err != nil {
		log.Error("Error connector client fetch %v", err.Error())
		return err
	}
	prg.connector = connectorClient

	status, err := status.New(meshbluClient, connectorClient.StatusUUID())
	if err != nil {
		log.Error("Error getting status device %v", err.Error())
		return err
	}
	err = status.ResetErrors()
	if err != nil {
		log.Error("Error resetting errors on status device %v", err.Error())
	}
	log.Info("Resetting errors on status device")
	prg.status = status

	githubSlug := prg.config.GithubSlug
	connectorName := prg.config.ConnectorName
	dir := prg.config.Dir
	uc, err := updateconnector.New(githubSlug, connectorName, dir, nil)
	if err != nil {
		log.Error("Error getting update connector %v", err.Error())
		return err
	}
	prg.uc = uc
	client.prg = prg

	err = srv.Run()
	if err != nil {
		log.Error("Error running %v", err.Error())
		return err
	}
	return nil
}

// Shutdown will kill the connector process
func (client *Client) Shutdown() error {
	return client.prg.Stop(client.prg.srv)
}
