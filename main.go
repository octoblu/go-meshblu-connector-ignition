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
	"github.com/octoblu/go-meshblu-connector-ignition/runner"
)

func main() {
	app := cli.NewApp()
	app.Name = "meshblu-connector-ignition"
	app.Version = version()
	app.Action = run
	app.Flags = []cli.Flag{}
	app.Run(os.Args)
}

func run(context *cli.Context) {
	serviceConfig, err := runner.GetConfig()
	if err != nil {
		log.Fatalln("Error getting service config", err.Error())
	}

	runnerClient := runner.New(serviceConfig)
	err = runnerClient.Start()
	if err != nil {
		log.Fatalln("Error executing connector", err.Error())
	}

	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, syscall.SIGTERM)

	sigTermReceived := false

	go func() {
		<-sigTerm
		fmt.Println("SIGTERM received, waiting to exit")
		sigTermReceived = true
	}()

	for {
		if sigTermReceived {
			fmt.Println("SIGTERM received, shutting down.")
			runnerClient.Shutdown()
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
	}
	return version.String()
}
