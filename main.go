package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-semver/semver"
	"github.com/octoblu/go-meshblu-connector-ignition/forever"
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
	mainLogger = logger.GetMainLogger()
	mainLogger.Info("main", fmt.Sprintf("starting %v...", version()))
	defer mainLogger.Clear()
	defer mainLogger.Close()

	serviceConfig, err := runner.GetConfig()
	fatalIfErr(err, "Error getting service config")

	foreverClient := forever.NewRunner(serviceConfig, version())
	foreverClient.Start()

	os.Exit(0)
}

func fatalIfErr(err error, msg string) {
	if err == nil {
		return
	}

	mainLogger.Error("main", msg, err)
	time.Sleep(time.Second * 1)
	os.Exit(1)
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
