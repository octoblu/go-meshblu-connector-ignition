package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	isterminal "github.com/azer/is-terminal"
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
	interactive := isterminal.IsTerminal(syscall.Stderr)
	stdout, err := logger.NewLogger(config.Stdout, interactive, false)
	if err != nil {
		return nil, err
	}
	stderr, err := logger.NewLogger(config.Stderr, interactive, true)
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
				log.Error("Error fetching connector %v", err.Error())
				return err
			}
			log.Info("Device stopped, waiting for connector to change")
			if prg.connector.Stopped() {
				time.Sleep(30 * time.Second)
			} else {
				log.Info("State Changed to Started")
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
	log.Info("Starting %v", prg.config.DisplayName)
	prg.running = true
	prg.cmd.Stderr = prg.stderr.Stream()
	prg.cmd.Stdout = prg.stdout.Stream()

	prg.uc.WritePID()

	prg.checkForChangesInterval()

	prg.timeStarted = time.Now()
	err := prg.cmd.Run()
	if err != nil {
		log.Error("Error running: %v", err)
		prg.running = false
	}
	prg.updateErrors()
	err = prg.tryAgain()
	if err != nil {
		log.Error("Error running: %v", err)
	}
}

func (prg *Program) tryAgain() error {
	timeSinceStarted := time.Since(prg.timeStarted)
	if timeSinceStarted > time.Minute {
		log.Info("Program ran for %v minutes, resetting backoff", timeSinceStarted)
		prg.b.Reset()
		log.Info("Restarting now")
		return prg.internalStart(false)
	}
	duration := prg.b.Duration()
	log.Info("Restarting in %v seconds", duration)
	time.Sleep(duration)
	return prg.internalStart(false)
}

// Stop service but really
func (prg *Program) Stop(srv service.Service) error {
	close(prg.exit)
	log.Info("Stopping %v", prg.config.DisplayName)
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
		log.Error("Error updating errors %v", err.Error())
	} else {
		log.Info("Updated status device with errors")
	}
	return nil
}

func (prg *Program) internalStop() error {
	log.Info("Internal Stopping %v", prg.config.DisplayName)
	if prg.cmd != nil {
		log.Info("Killing process")
		prg.cmd.Process.Kill()
	}
	log.Info("Internal Stopped")
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
	log.Info("Checking for changes")
	err := prg.connector.Fetch()
	if err != nil {
		log.Error("Device Update Error: %v", err.Error())
		return err
	}
	versionChange := prg.connector.DidVersionChange()
	if versionChange {
		log.Info("Device Version Change %v", prg.connector.Version())
		err := prg.update()
		if err != nil {
			return err
		}
	}
	stopChange := prg.connector.DidStopChange()
	if stopChange {
		log.Info("Device Stop Change")
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
		log.Error("needsUpdate -> Error %v", err.Error())
		return err
	}
	if !needsUpdate {
		log.Error("no update needed (%s)", tag)
		return nil
	}
	if prg.running {
		err = prg.internalStop()
		if err != nil {
			log.Error("internalStop -> Error %v", err.Error())
			return err
		}
	}
	err = prg.uc.Do(tag)
	if err != nil {
		log.Error("update connector error (%s) %v", tag, err.Error())
		return err
	}
	if !prg.running {
		err = prg.internalStart(false)
		if err != nil {
			log.Error("internalStop -> Error %v", err.Error())
			return err
		}
	}
	return nil
}

func (prg *Program) checkForChangesInterval() {
	duration := time.Minute
	if prg.ticker != nil {
		log.Info("changes interval already exists, canceling it now")
		prg.ticker.Stop()
	}
	prg.ticker = time.NewTicker(duration)

	go func() {
		for {
			select {
			case <-prg.ticker.C:
				prg.checkForChanges()
				log.Info("Will check for meshblu device changes in %v", duration)
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
		log.Error("Failed to get Executable File, %s - Error %v", file, err)
		return "", err
	}
	log.Info("got executable %s", file)
	return file, nil
}
