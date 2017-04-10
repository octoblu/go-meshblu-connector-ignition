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
}

// NewRunner creates a new instance of the forever runner
func NewRunner(serviceConfig *runner.Config) Forever {
	runnerClient := runner.New(serviceConfig)
	return &Client{runnerClient: runnerClient, running: false}
}

// Start runs the connector forever
func (client *Client) Start() {
	if mainLogger == nil {
		mainLogger = logger.GetMainLogger()
	}
	client.waitForSigterm()
	client.running = true
	for {
		mainLogger.Info("forever", "starting runner...")
		if client.running == false {
			return
		}
		err := client.runnerClient.Start()
		if err != nil {
			mainLogger.Error("forever", "Error running connector (will retry soon)", err)
			time.Sleep(10 * time.Second)
			continue
		}
		for {
			if client.running == false {
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
