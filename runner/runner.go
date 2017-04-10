package runner

import (
	"path/filepath"
	"time"

	"github.com/kardianos/service"
	"github.com/octoblu/go-meshblu-connector-ignition/connector"
	"github.com/octoblu/go-meshblu-connector-ignition/logger"
	"github.com/octoblu/go-meshblu-connector-ignition/status"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"
	"github.com/octoblu/go-meshblu/http/meshblu"
)

var mainLogger logger.MainLogger

// Runner defines the interface to run a Cmd
type Runner interface {
	Start() error
	Shutdown() error
	IsRunning() bool
}

// Client defines the stucture of the client
type Client struct {
	config    *Config
	prg       *Program
	isRunning bool
}

// New creates a new instance of runner
func New(config *Config) Runner {
	return &Client{config, nil, false}
}

// Start runs the connector
func (client *Client) Start() error {
	if mainLogger == nil {
		mainLogger = logger.GetMainLogger()
	}
	prg, err := NewProgram(client.config)
	if err != nil {
		mainLogger.Error("runner", "Error creating new program", err)
		return err
	}

	srvConfig := &service.Config{
		Name:        client.config.ServiceName,
		DisplayName: client.config.DisplayName,
		Description: client.config.Description,
	}

	srv, err := service.New(prg, srvConfig)
	if err != nil {
		mainLogger.Error("runner", "Error getting service", err)
		return err
	}

	meshbluConfigPath := filepath.Join(client.config.Dir, "meshblu.json")
	meshbluClient, uuid, err := meshblu.NewClient(meshbluConfigPath)
	if err != nil {
		mainLogger.Error("runner", "Error getting meshblu client", err)
		return err
	}
	connectorClient, err := connector.New(meshbluClient, uuid, client.config.Tag)
	if err != nil {
		mainLogger.Error("runner", "Error connector client new", err)
		return err
	}

	for {
		meshbluErr := connectorClient.Fetch()
		if meshbluErr == nil {
			break
		}
		mainLogger.Error("runner", "Error connector client fetch", meshbluErr)
		if meshblu.IsRecoverable(meshbluErr) {
			mainLogger.Info("runner", "Error was recoverable, trying again in 10s")
			time.Sleep(10 * time.Second)
			continue
		}
		return meshbluErr
	}

	prg.connector = connectorClient

	status, err := status.New(meshbluClient, connectorClient.StatusUUID())
	if err != nil {
		mainLogger.Error("runner", "error getting status device", err)
		return err
	}
	err = status.ResetErrors()
	if err != nil {
		mainLogger.Error("runner", "error resetting errors on status device", err)
	} else {
		mainLogger.Info("runner", "reset errors on status device")
	}
	prg.status = status

	githubSlug := prg.config.GithubSlug
	connectorName := prg.config.ConnectorName
	dir := prg.config.Dir
	uc, err := updateconnector.New(githubSlug, connectorName, dir, nil, nil)
	if err != nil {
		mainLogger.Error("runner", "Error getting update connector", err)
		return err
	}
	prg.uc = uc
	client.prg = prg

	err = srv.Run()
	if err != nil {
		mainLogger.Error("runner", "Error running", err)
		return err
	}
	client.isRunning = true
	return nil
}

// Shutdown will kill the connector process
func (client *Client) Shutdown() error {
	err := client.prg.Stop(nil)
	client.isRunning = false
	return err
}

// IsRunning will return true if the connector is reported as running
func (client *Client) IsRunning() bool {
	return client.isRunning
}
