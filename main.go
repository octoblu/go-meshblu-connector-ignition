package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-semver/semver"
	"github.com/octoblu/go-meshblu-connector-ignition/logger"
	"github.com/octoblu/go-meshblu-connector-ignition/runner"
)

var mainLogger logger.MainLogger

func main() {
	app := cli.NewApp()
	app.Name = "meshblu-connector-ignition"
	app.Version = version()
	app.Action = run
	app.Flags = []cli.Flag{}
	app.Run(os.Args)
}

func run(context *cli.Context) {
	err := logger.InitMainLogger()
	if err != nil {
		log.Panicln("Error initializing the main logger", err.Error())
		os.Exit(1)
		return
	}
	mainLogger := logger.GetMainLogger()

	serviceConfig, err := runner.GetConfig()
	if err != nil {
		mainLogger.Error("main", "Error getting service config", err)
		os.Exit(1)
		return
	}

	runnerClient := runner.New(serviceConfig)
	err = runnerClient.Start()
	if err != nil {
		mainLogger.Error("main", "Error getting service config", err)
		os.Exit(1)
		return
	}

	mainLogger.Info("main", "Starting...")

	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, syscall.SIGTERM)

	sigTermReceived := false

	go func() {
		<-sigTerm
		mainLogger.Info("main", "SIGTERM received, waiting to exit")
		sigTermReceived = true
	}()

	for {
		if sigTermReceived {
			mainLogger.Info("main", "SIGTERM received, shutting down...")
			runnerClient.Shutdown()
			mainLogger.Clear()
			mainLogger.Close()
			os.Exit(0)
		}

		time.Sleep(1 * time.Second)
	}
}

func version() string {
	version, err := semver.NewVersion(VERSION)
	if err != nil {
		errorMessage := fmt.Sprintf("Error with version number: %v", VERSION)
		log.Panicln(errorMessage, err.Error())
		os.Exit(1)
		return ""
	}
	return version.String()
}
