package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	isterminal "github.com/azer/is-terminal"
	mainlogger "github.com/azer/logger"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-semver/semver"
	"github.com/kardianos/osext"
	"github.com/octoblu/go-meshblu-connector-ignition/runner"
)

var logMain = mainlogger.New("main")

func main() {
	initMainLogger()
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
		logMain.Error("Error getting service config", err.Error())
		log.Fatalln("Error getting service config", err.Error())
		return
	}

	runnerClient := runner.New(serviceConfig)
	err = runnerClient.Start()
	if err != nil {
		logMain.Error("Error getting service config", err.Error())
		log.Fatalln("Error executing connector", err.Error())
		return
	}

	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, syscall.SIGTERM)

	sigTermReceived := false

	go func() {
		<-sigTerm
		logMain.Info("SIGTERM received, waiting to exit")
		sigTermReceived = true
	}()

	for {
		if sigTermReceived {
			logMain.Info("SIGTERM received, shutting down...")
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

func getIgnitionLogFile() (*os.File, error) {
	fullPath, err := osext.Executable()
	if err != nil {
		return nil, err
	}
	dir, _ := filepath.Split(fullPath)
	logFilePath := filepath.Join(dir, "log", "ignition.log")
	return os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0777)
}

func initMainLogger() {
	mainlogger.AllEnabled = true
	if isterminal.IsTerminal(syscall.Stderr) {
		fmt.Println("interactive mode")
		return
	}
	logFile, err := getIgnitionLogFile()
	if err != nil {
		log.Panicln("Error getting log file", err.Error())
		return
	}
	mainlogger.SetOutput(logFile)
}
