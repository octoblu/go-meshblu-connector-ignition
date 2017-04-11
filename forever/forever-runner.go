package forever

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	pid := os.Getpid()
	mainLogger.Info("forever", fmt.Sprintf("writing pid %v", pid))
	writePID(pid)
	client.waitForProcessChange()
	client.waitForSigterm()
	client.waitForUpdate()
	client.running = true
	for {
		if !client.running {
			return
		}
		err := client.runnerClient.Start()
		if err != nil {
			mainLogger.Error("forever", "Error running connector (will retry soon)", err)
			time.Sleep(10 * time.Second)
			continue
		}
		mainLogger.Info("forever", "running...")
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
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range sigChan {
			mainLogger.Info("forever", "Interrupt received, waiting to exit")
			client.Shutdown()
		}
	}()
}

func (client *Client) waitForUpdate() {
	go func() {
		firstTime := true
		for {
			if !firstTime {
				time.Sleep(time.Minute)
				// startNew()
				// return
			}
			firstTime = false
			latestVersion, err := resolveLatestVersion()
			if err != nil {
				mainLogger.Error("forever", "Cannot get latest version", err)
				continue
			}
			if !shouldUpdate(client.currentVersion, latestVersion) {
				continue
			}
			mainLogger.Info("forever", fmt.Sprintf("there is a new ignition version %s", latestVersion))
			err = doUpdate(latestVersion)
			if err != nil {
				mainLogger.Error("forever", "Error updating myself", err)
				continue
			}
			err = startNew()
			if err != nil {
				mainLogger.Error("forever", "start new error", err)
				continue
			}
			mainLogger.Info("forever", "I am updated and started new process")
			return
		}
	}()
}

func (client *Client) waitForProcessChange() {
	go func() {
		for {
			time.Sleep(time.Second)
			pid, err := getPID()
			if err != nil {
				mainLogger.Error("forever", "get pid error", err)
				continue
			}
			if pid == 0 {
				mainLogger.Info("forever", "pid is 0")
			}
			if pid == os.Getpid() {
				continue
			}
			mainLogger.Info("forever", "process changed")
			client.Shutdown()
			return
		}
	}()
}
