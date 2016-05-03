package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-semver/semver"
	"github.com/octoblu/go-meshblu-connector-ignition/runner"
	De "github.com/tj/go-debug"
)

var debug = De.Debug("meshblu-connector-ignition:main")

func main() {
	app := cli.NewApp()
	app.Name = "meshblu-connector-ignition"
	app.Version = version()
	app.Action = run
	app.Flags = []cli.Flag{}
	app.Run(os.Args)
}

func run(context *cli.Context) {
	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, syscall.SIGTERM)

	serviceConfig, err := runner.GetConfig()
	if err != nil {
		log.Fatalln("Error executing connector", err.Error())
	}

	runnerClient := runner.New(serviceConfig)

	go func() {
		<-sigTerm
		fmt.Println("SIGTERM received, waiting to exit")
		err := runnerClient.Shutdown()
		if err != nil {
			log.Fatalln("Error SIGTERMing connector", err.Error())
		}
		fmt.Println("Connector shutdown. I'll be back.")
		os.Exit(0)
	}()

	err = runnerClient.Start()
	if err != nil {
		log.Fatalln("Error executing connector", err.Error())
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
