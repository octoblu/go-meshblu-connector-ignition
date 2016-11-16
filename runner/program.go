package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jpillora/backoff"
	"github.com/kardianos/service"
	"github.com/mattn/go-isatty"
	"github.com/octoblu/go-meshblu-connector-ignition/connector"
	"github.com/octoblu/go-meshblu-connector-ignition/logger"
	"github.com/octoblu/go-meshblu-connector-ignition/status"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"
)

// Program inteface that is real
type Program struct {
	config       *Config
	srv          service.Service
	cmd          *exec.Cmd
	connector    connector.Connector
	status       status.Status
	uc           updateconnector.UpdateConnector
	retrySeconds int
	b            *backoff.Backoff
	timeStarted  time.Time
	running      bool
	stderr       logger.Logger
	stdout       logger.Logger
	exit         chan struct{}
	ticker       *time.Ticker
}

// NewProgram creates a new program cient
func NewProgram(config *Config) (*Program, error) {
	if mainLogger == nil {
		mainLogger = logger.GetMainLogger()
	}
	stdout, err := logger.NewLogger(config.Stdout, false)
	if err != nil {
		return nil, err
	}
	stderr, err := logger.NewLogger(config.Stderr, true)
	if err != nil {
		return nil, err
	}
	return &Program{
		config: config,
		b: &backoff.Backoff{
			Min: 5 * time.Second,
			Max: 30 * time.Minute,
		},
		running: false,
		stderr:  stderr,
		stdout:  stdout,
		exit:    make(chan struct{}),
	}, nil
}

// Start service but really
func (prg *Program) Start(srv service.Service) error {
	err := prg.uc.ClearPID()
	if err != nil {
		mainLogger.Error("program.Start", "Error clearing PID", err)
		return err
	}
	if prg.connector.Stopped() {
		mainLogger.Info("program.Start", "connector is stopped, fetching")
		for {
			err = prg.connector.Fetch()
			if err != nil {
				mainLogger.Error("program.Start", "Error fetching connector", err)
				return err
			}
			mainLogger.Info("program.Start", "Device stopped, waiting for connector to change")
			if prg.connector.Stopped() {
				time.Sleep(30 * time.Second)
			} else {
				mainLogger.Info("program.Start", "State Changed to Started")
				break
			}
		}
	}
	mainLogger.Info("program.Start", "running internal start")
	return prg.internalStart(true)
}

func (prg *Program) internalStart(fork bool) error {
	tag := prg.connector.VersionWithV()
	needsUpdate, err := prg.uc.NeedsUpdate(tag)
	if err != nil {
		mainLogger.Error("program.internalStart", "Needs Update Error", err)
		return err
	}
	if needsUpdate {
		mainLogger.Info("program.internalStart", "Updating")
		err = prg.uc.Do(tag)
		if err != nil {
			mainLogger.Error("program.internalStart", "update.Do Error", err)
			return err
		}
	}
	commandPath := prg.getCommandPath()
	nodeCommand, err := prg.TheExecutable("node")
	if err != nil {
		mainLogger.Error("program.internalStart", "the executable error", err)
		return err
	}
	prg.cmd = exec.Command(nodeCommand, commandPath)
	prg.initCmd(prg.cmd)
	prg.initCmdForOS(prg.cmd)

	if fork {
		go prg.run()
	} else {
		prg.run()
	}

	return nil
}

func (prg *Program) run() {
	mainLogger.Info("program.run", fmt.Sprintf("Starting %v", prg.config.DisplayName))
	prg.running = true
	prg.cmd.Stderr = prg.stderr.Stream()
	prg.cmd.Stdout = prg.stdout.Stream()

	prg.uc.WritePID()

	prg.checkForChangesInterval()

	prg.timeStarted = time.Now()
	err := prg.cmd.Run()
	if err != nil {
		mainLogger.Error("program.run", "Error running", err)
		prg.running = false
	}
	prg.updateErrors()
	err = prg.tryAgain()
	if err != nil {
		mainLogger.Error("program.run", "Error running again", err)
	}
}

func (prg *Program) tryAgain() error {
	timeSinceStarted := time.Since(prg.timeStarted)
	if timeSinceStarted > time.Minute {
		mainLogger.Info("program.tryAgain", fmt.Sprintf("Program ran for %v minutes, resetting backoff", timeSinceStarted))
		prg.b.Reset()
		mainLogger.Info("program.tryAgain", "Restarting now")
		return prg.internalStart(false)
	}
	duration := prg.b.Duration()
	mainLogger.Info("program.tryAgain", fmt.Sprintf("Restarting in %v seconds", duration))
	time.Sleep(duration)
	return prg.internalStart(false)
}

// Stop service but really
func (prg *Program) Stop(srv service.Service) error {
	close(prg.exit)
	mainLogger.Info("program.Stop", fmt.Sprintf("Stopping %v", prg.config.DisplayName))
	defer prg.stderr.Close()
	defer prg.stdout.Close()
	err := prg.internalStop()
	if err != nil {
		mainLogger.Error("program.Stop", "Error stopping", err)
		return err
	}
	if isatty.IsTerminal(os.Stdout.Fd()) {
		os.Exit(0)
	}
	return nil
}

func (prg *Program) updateErrors() error {
	err := prg.status.UpdateErrors(prg.stderr.Get())
	if err != nil {
		mainLogger.Error("program.updateErrors", "Error updating errors", err)
	} else {
		mainLogger.Info("program.updateErrors", "Updated status device with errors")
	}
	return nil
}

func (prg *Program) internalStop() error {
	mainLogger.Info("program.internalStop", fmt.Sprintf("Internal Stopping %v", prg.config.DisplayName))
	if prg.cmd != nil {
		mainLogger.Info("program.internalStop", "Killing process")
		prg.cmd.Process.Kill()
	}
	mainLogger.Info("program.intervalStop", "Internal Stopped")
	return nil
}

func (prg *Program) initCmd(cmd *exec.Cmd) {
	cmd.Dir = prg.config.Dir
	env := prg.getEnv()
	cmd.Env = env
}

func (prg *Program) getCommandPath() string {
	return fmt.Sprintf(".%s%s", string(filepath.Separator), filepath.Join("node_modules", "meshblu-connector-runner", "command.js"))
}

func (prg *Program) checkForChanges() error {
	mainLogger.Info("program.checkForChanges", "Checking for changes")
	err := prg.connector.Fetch()
	if err != nil {
		mainLogger.Error("program.checkForChanges", "Device Update Error", err)
		return err
	}
	versionChange := prg.connector.DidVersionChange()
	if versionChange {
		mainLogger.Info("program.checkForChanges", fmt.Sprintf("Device Version Change %v", prg.connector.Version()))
	}
	err = prg.update()
	if err != nil {
		return err
	}
	stopChange := prg.connector.DidStopChange()
	if stopChange {
		mainLogger.Info("program.checkForChanges", "Device Stop Change")
		if prg.connector.Stopped() {
			err := prg.internalStop()
			if err != nil {
				return err
			}
		} else {
			err := prg.internalStart(false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (prg *Program) update() error {
	tag := prg.connector.VersionWithV()
	needsUpdate, err := prg.uc.NeedsUpdate(tag)
	if err != nil {
		mainLogger.Error("program.update", "Failed to run needsUpdate", err)
		return err
	}
	if !needsUpdate {
		mainLogger.Info("program.update", fmt.Sprintf("no update needed (%s)", tag))
		return nil
	}
	if prg.running {
		err = prg.internalStop()
		if err != nil {
			mainLogger.Error("program.update", "Failed to run Stop", err)
			return err
		}
	}
	err = prg.uc.Do(tag)
	if err != nil {
		mainLogger.Error("program.update", "Failed to run Do", err)
		return err
	}
	if !prg.running {
		err = prg.internalStart(false)
		if err != nil {
			mainLogger.Error("program.update", "Failed to run Start", err)
			return err
		}
	}
	return nil
}

func (prg *Program) checkForChangesInterval() {
	duration := time.Minute
	if prg.ticker != nil {
		mainLogger.Info("program.checkForChangesInterval", "changes interval already exists, canceling it now")
		prg.ticker.Stop()
	}
	prg.ticker = time.NewTicker(duration)

	go func() {
		for {
			select {
			case <-prg.ticker.C:
				prg.checkForChanges()
				mainLogger.Info("program.checkForChangesInterval", fmt.Sprintf("Will check for meshblu device changes in %v", duration))
			}
		}
	}()
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
	mainLogger.Info("program.TheExectuable", fmt.Sprintf("got executable %s", file))
	return file, nil
}
