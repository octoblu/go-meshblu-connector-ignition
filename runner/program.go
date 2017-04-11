package runner

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jpillora/backoff"
	"github.com/kardianos/service"
	"github.com/octoblu/go-meshblu-connector-ignition/connector"
	"github.com/octoblu/go-meshblu-connector-ignition/interval"
	"github.com/octoblu/go-meshblu-connector-ignition/logger"
	"github.com/octoblu/go-meshblu-connector-ignition/status"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"
	"github.com/octoblu/process"
	uuid "github.com/satori/go.uuid"
)

// Program inteface that is real
type Program struct {
	config        *Config
	cmd           *exec.Cmd
	cmdGroup      *process.Group
	connector     connector.Connector
	currentRun    string
	status        status.Status
	uc            updateconnector.UpdateConnector
	boff          *backoff.Backoff
	timeStarted   time.Time
	errLog        logger.Logger
	outLog        logger.Logger
	interval      interval.Interval
	started       bool
	shouldRestart bool
	restartChan   chan bool
}

// NewProgram creates a new program cient
func NewProgram(config *Config) (*Program, error) {
	if mainLogger == nil {
		mainLogger = logger.GetMainLogger()
	}

	outLog, err := logger.NewLogger(config.Stdout, false)
	if err != nil {
		return nil, err
	}

	errLog, err := logger.NewLogger(config.Stderr, true)
	if err != nil {
		return nil, err
	}

	return &Program{
		config: config,
		boff: &backoff.Backoff{
			Min: time.Second,
			Max: time.Minute,
		},
		errLog:        errLog,
		outLog:        outLog,
		started:       false,
		shouldRestart: true,
		restartChan:   make(chan bool, 1),
	}, nil
}

// Start service but really
func (prg *Program) Start(_ service.Service) error {
	mainLogger.Info("program.Start", fmt.Sprintf("starting %v", prg.config.DisplayName))
	err := prg.uc.WritePID()
	if err != nil {
		mainLogger.Error("program.Start", "error writing PID", err)
		return err
	}
	prg.shouldRestart = true
	go prg.restartLoop()
	prg.restartChan <- true
	return nil
}

// Stop service but really
func (prg *Program) Stop(_ service.Service) error {
	mainLogger.Info("program.Stop", fmt.Sprintf("stopping %v", prg.config.DisplayName))
	defer prg.errLog.Close()
	defer prg.outLog.Close()
	prg.started = false
	prg.shouldRestart = false
	return nil
}

func (prg *Program) restart() {
	mainLogger.Info("program.restart", "restart called")
	prg.restartChan <- true
}

func (prg *Program) restartLoop() error {
	for range prg.restartChan {
		if !prg.shouldRestart {
			mainLogger.Info("program.restartLoop", "should not restart")
			return nil
		}
		if prg.started {
			mainLogger.Info("program.restartLoop", "restart signal received")
		}
		backoffDuration := prg.boff.Duration()
		currentRun := uuid.NewV4().String()
		prg.currentRun = currentRun

		if prg.started {
			mainLogger.Info("program.restartLoop", fmt.Sprintf("waiting for %v due to backoff", backoffDuration))
			time.Sleep(backoffDuration)
		}
		err := prg.stop()
		if err != nil {
			mainLogger.Error("program.restartLoop", "failed to stop existing child", err)
			return err
		}
		if prg.started {
			mainLogger.Info("program.restartLoop", "existing child stopped")
		}

		err = prg.update()
		if err != nil {
			mainLogger.Error("program.restartLoop", "failed to prg.update", err)
		} else {
			mainLogger.Info("program.restartLoop", "updated")
		}

		commandPath := prg.getCommandPath()
		nodeCommand, err := prg.TheExecutable("node")
		if err != nil {
			mainLogger.Error("program.restartLoop", "the executable error", err)
			return err
		}
		prg.cmd = exec.Command(nodeCommand, commandPath)
		prg.cmd.Dir = prg.config.Dir
		prg.cmd.Env = prg.getEnv()
		prg.cmd.SysProcAttr = sysProcAttrForOS()
		prg.cmd.Stderr = prg.errLog.Stream()
		prg.cmd.Stdout = prg.outLog.Stream()
		prg.cmdGroup, err = process.Background(prg.cmd)

		go func() {
			if waitErr := prg.cmdGroup.Wait(); waitErr != nil {
				mainLogger.Error("prg.cmd.Wait", "connector Runner error'd", waitErr)
				if currentRun != prg.currentRun {
					mainLogger.Info("prg.cmd.Wait", "not the currentRun, ignoring")
					return
				}
				prg.restart()
			}
		}()

		go func() {
			time.Sleep(time.Second * 30)
			if currentRun == prg.currentRun {
				mainLogger.Info("program.restartLoop", "ran for 30s without dying, resetting backoff")
				prg.boff.Reset()
			}
		}()

		prg.checkForChangesOnInterval()

		if err != nil {
			return err
		}
		if prg.started {
			mainLogger.Info("program.restartLoop", "restarted")
		}
		prg.started = true
	}

	return prg.stop()
}

func (prg *Program) updateErrors() error {
	err := prg.status.UpdateErrors(prg.errLog.Get())
	if err != nil {
		mainLogger.Error("program.updateErrors", "Error updating errors", err)
	} else {
		mainLogger.Info("program.updateErrors", "Updated status device with errors")
	}
	return nil
}

func (prg *Program) stop() error {
	if prg.started {
		mainLogger.Info("program.stop", "stopping connector")
	}
	if prg.cmdGroup == nil {
		return nil
	}
	err := prg.cmdGroup.Terminate(time.Second * 30)
	if _, isExitError := err.(*exec.ExitError); isExitError {
		return nil
	}
	return err
}

func (prg *Program) getCommandPath() string {
	return fmt.Sprintf(".%s%s", string(filepath.Separator), filepath.Join("node_modules", "meshblu-connector-runner", "command.js"))
}

func (prg *Program) checkForChanges() error {
	err := prg.connector.Fetch()
	if err != nil {
		mainLogger.Error("program.checkForChanges", "Device Update Error", err)
		return err
	}
	versionChange := prg.connector.DidVersionChange()
	if versionChange {
		mainLogger.Info("program.checkForChanges", fmt.Sprintf("Device Version Change %v", prg.connector.Version()))
		prg.boff.Reset()
		prg.restart()
	}
	return nil
}

func (prg *Program) update() error {
	err := prg.connector.Fetch()
	if err != nil {
		mainLogger.Error("program.update", "Failed to run prg.connector.Fetch", err)
		return err
	}

	tag := prg.connector.VersionWithV()
	needsUpdate, err := prg.uc.NeedsUpdate(tag)
	if err != nil {
		mainLogger.Error("program.update", "Failed to run prg.uc.needsUpdate", err)
		return err
	}
	if !needsUpdate {
		mainLogger.Info("program.update", fmt.Sprintf("no update needed (%s)", tag))
		return nil
	}
	err = prg.uc.Do(tag)
	if err != nil {
		mainLogger.Error("program.update", "Failed to run uc.Do", err)
		return err
	}
	return nil
}

func (prg *Program) checkForChangesOnInterval() {
	mainLogger.Info("program.checkForChangesOnInterval", "started")

	if prg.interval != nil {
		prg.interval.Clear()
	}

	duration := time.Minute
	prg.interval = interval.SetInterval(duration, func() {
		prg.checkForChanges()
	})
}

func (prg *Program) getFullConnectorName() string {
	return fmt.Sprintf("meshblu-%s", prg.config.ConnectorName)
}

func (prg *Program) getEnv() []string {
	pathEnv := GetPathEnv(prg.config.BinPath)
	return GetEnviron(pathEnv)
}

// TheExecutable should return the correct executable
func (prg *Program) TheExecutable(name string) (string, error) {
	thePath := filepath.Join(prg.config.BinPath, name)
	file, err := exec.LookPath(thePath)
	if err != nil {
		mainLogger.Error("program.TheExecutable", "Error getting executable", err)
		return "", err
	}
	mainLogger.Info("program.TheExectuable", fmt.Sprintf("using executable %s", file))
	return file, nil
}
