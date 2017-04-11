package forever

import (
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
	runnerClient runner.Runner
	running      bool
	restart      bool
}

// NewRunner creates a new instance of the forever runner
func NewRunner(serviceConfig *runner.Config) Forever {
	runnerClient := runner.New(serviceConfig)
	return &Client{
		runnerClient: runnerClient,
		running:      false,
		restart:      false,
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
		mainLogger.Info("forever", "starting runner...")
		if !client.running {
			return
		}
		err := client.runnerClient.Start()
		mainLogger.Info("forever", "started...")
		if err != nil {
			mainLogger.Error("forever", "Error running connector (will retry soon)", err)
			time.Sleep(10 * time.Second)
			continue
		}
		client.restart = false
		for {
			if !client.running {
				mainLogger.Info("forever", "forever is over, shutting down")
				return
			}
			if client.restart {
				mainLogger.Info("forever", "forever is going to restart")
				break
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
		time.Sleep(time.Second * 10)
		mainLogger.Info("forever", "I AM GOING TO UPDATE MYSELF")
		err := doUpdate()
		if err != nil {
			mainLogger.Error("forever", "Error updating myself", err)
			return
		}
		client.restart = true
		mainLogger.Info("forever", "UPDATED, restart set")
	}()
}
