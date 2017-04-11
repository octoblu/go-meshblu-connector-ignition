package forever

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/octoblu/go-meshblu-connector-ignition/logger"
	"github.com/octoblu/go-meshblu-connector-ignition/runner"
)

var mainLogger logger.MainLogger

// Forever defines the interface to run the runner for forever
type Forever interface {
	Start()
	Shutdown()
}

// Client defines the stucture of the client
type Client struct {
	runnerClient   runner.Runner
	running        bool
	currentVersion string
}

// NewRunner creates a new instance of the forever runner
func NewRunner(serviceConfig *runner.Config, currentVersion string) Forever {
	runnerClient := runner.New(serviceConfig)
	return &Client{
		runnerClient:   runnerClient,
		running:        false,
		currentVersion: currentVersion,
	}
}

// Start runs the connector forever
func (client *Client) Start() {
	if mainLogger == nil {
		mainLogger = logger.GetMainLogger()
	}
	client.waitForSigterm()
	client.waitForUpdate()
	client.running = true
	for {
		if !client.running {
			return
		}
		mainLogger.Info("forever", "starting runner...")
		err := client.runnerClient.Start()
		mainLogger.Info("forever", "started...")
		if err != nil {
			mainLogger.Error("forever", "Error running connector (will retry soon)", err)
			time.Sleep(10 * time.Second)
			continue
		}
		for {
			if !client.running {
				mainLogger.Info("forever", "forever is over, shutting down")
				return
			}
			time.Sleep(time.Second)
			continue
		}
	}
}

// Shutdown will give tell the connector runner it is time to shutdown
func (client *Client) Shutdown() {
	mainLogger.Info("forever", "forever is going to Shutdown")
	client.running = false
}

func (client *Client) waitForSigterm() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		for range sigChan {
			mainLogger.Info("forever", "Interrupt received, waiting to exit")
			client.Shutdown()
		}
	}()
}

func (client *Client) waitForUpdate() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			latestVersion, err := resolveLatestVersion()
			if err != nil {
				mainLogger.Error("forever", "Cannot get latest version", err)
				continue
			}
			if !shouldUpdate(client.currentVersion, latestVersion) {
				continue
			}
			mainLogger.Info("forever", fmt.Sprintf("I AM GOING TO UPDATE MYSELF TO %s", latestVersion))
			err = doUpdate(latestVersion)
			if err != nil {
				mainLogger.Error("forever", "Error updating myself", err)
				continue
			}
			client.waitForProcessChange()
			err = startNew()
			if err != nil {
				mainLogger.Error("forever", "Error starting new process", err)
				continue
			}
			return
		}
	}()
}

func (client *Client) waitForProcessChange() {
	go func() {
		for {
			same, err := sameProcess()
			if err != nil {
				mainLogger.Error("forever", "Error checking PID", err)
				continue
			}
			if !same {
				mainLogger.Info("forever", "PROCESS CHANGED SHUTTING DOWN")
				client.runnerClient.Shutdown()
				client.Shutdown()
				return
			}
			time.Sleep(time.Second)
		}
	}()
}
