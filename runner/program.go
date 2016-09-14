package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jpillora/backoff"
	"github.com/kardianos/service"
	"github.com/octoblu/go-meshblu-connector-ignition/connector"
	"github.com/octoblu/go-meshblu-connector-ignition/logger"
	"github.com/octoblu/go-meshblu-connector-ignition/status"
	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"
)

// Program inteface that is real
type Program struct {
	config       *Config
	srv          service.Service
	logger       service.Logger
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
}

// NewProgram creates a new program cient
func NewProgram(config *Config) (*Program, error) {
	stdout, err := logger.NewLogger(config.Stdout, service.Interactive(), false)
	if err != nil {
		return nil, err
	}
	stderr, err := logger.NewLogger(config.Stderr, service.Interactive(), true)
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
	if prg.connector.Stopped() {
		for {
			err := prg.connector.Fetch()
			if err != nil {
				return err
			}
			prg.logger.Info("Device stopped, waiting for connector to change")
			if prg.connector.Stopped() {
				time.Sleep(30 * time.Second)
			} else {
				prg.logger.Info("State Changed to Started")
				break
			}
		}
	}

	return prg.internalStart(true)
}

func (prg *Program) internalStart(fork bool) error {
	tag := prg.connector.VersionWithV()
	needsUpdate, err := prg.uc.NeedsUpdate(tag)
	if err != nil {
		return err
	}
	if needsUpdate {
		prg.uc.Do(tag)
	}
	commandPath := prg.getCommandPath()
	nodeCommand, err := prg.TheExecutable("node")
	if err != nil {
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
	prg.logger.Info("Starting ", prg.config.DisplayName)
	prg.running = true
	prg.cmd.Stderr = prg.stderr.Stream()
	prg.cmd.Stdout = prg.stdout.Stream()
	tag := prg.connector.VersionWithV()
	prg.uc.Dont(tag)

	prg.checkForChangesInterval()

	prg.timeStarted = time.Now()
	err := prg.cmd.Run()
	if err != nil {
		prg.logger.Warningf("Error running: %v", err)
		prg.running = false
	}
	prg.updateErrors()
	err = prg.tryAgain()
	if err != nil {
		prg.logger.Warningf("Error running: %v", err)
	}
}

func (prg *Program) tryAgain() error {
	timeSinceStarted := time.Since(prg.timeStarted)
	if timeSinceStarted > time.Minute {
		prg.logger.Infof("Program ran for 1 minute, resetting backoff")
		prg.b.Reset()
	}
	duration := prg.b.Duration()
	prg.logger.Infof("Restarting in %v seconds", duration)
	time.Sleep(duration)
	return prg.internalStart(false)
}

// Stop service but really
func (prg *Program) Stop(srv service.Service) error {
	close(prg.exit)
	prg.logger.Info("Stopping ", prg.config.DisplayName)
	defer prg.stderr.Close()
	defer prg.stdout.Close()
	err := prg.internalStop()
	if err != nil {
		return err
	}

	if service.Interactive() {
		os.Exit(0)
	}

	return nil
}

func (prg *Program) updateErrors() error {
	err := prg.status.UpdateErrors(prg.stderr.Get())
	if err != nil {
		prg.logger.Warningf("Error updating errors")
	} else {
		prg.logger.Info("Updated status device with errors")
	}
	return nil
}

func (prg *Program) internalStop() error {
	prg.logger.Info("Internal Stopping ", prg.config.DisplayName)
	if prg.cmd != nil {
		prg.cmd.Process.Kill()
	}
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
	err := prg.connector.Fetch()
	if err != nil {
		prg.logger.Warningf("Device Update Error: %v", err.Error())
		return err
	}
	versionChange := prg.connector.DidVersionChange()
	if versionChange {
		prg.logger.Infof("Device Version Change %v", prg.connector.Version())
		err := prg.update()
		if err != nil {
			return err
		}
	}
	stopChange := prg.connector.DidStopChange()
	if stopChange {
		prg.logger.Infof("Device Stop Change")
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
		return err
	}
	if !needsUpdate {
		return nil
	}
	if prg.running {
		err := prg.internalStop()
		if err != nil {
			return err
		}
	}
	err = prg.uc.Do(tag)
	if err != nil {
		return err
	}
	if !prg.running {
		err = prg.internalStart(false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (prg *Program) checkForChangesInterval() {
	ticker := time.NewTicker(time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				prg.checkForChanges()
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
		prg.logger.Warningf("Failed to get Executable File, %s - Error %v", file, err)
		return "", err
	}
	prg.logger.Infof("got executable %s", file)
	return file, nil
}
