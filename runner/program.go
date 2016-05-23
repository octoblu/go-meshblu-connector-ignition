package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jpillora/backoff"
	"github.com/kardianos/service"
	"github.com/octoblu/go-meshblu-connector-ignition/device"
)

// Program inteface that is real
type Program struct {
	config       *Config
	srv          service.Service
	logger       service.Logger
	cmd          *exec.Cmd
	device       device.Device
	uc           *UpdateConnector
	retrySeconds int
	b            *backoff.Backoff
	timeStarted  time.Time
	exit         chan struct{}
}

// NewProgram creates a new program cient
func NewProgram(config *Config) *Program {
	return &Program{
		config: config,
		b: &backoff.Backoff{
			Min: 5 * time.Second,
			Max: 30 * time.Minute,
		},
		exit: make(chan struct{}),
	}
}

// Start service but really
func (prg *Program) Start(srv service.Service) error {
	if prg.config.Legacy {
		err := prg.uc.DoLegacy()
		if err != nil {
			return err
		}
	}
	if prg.device.Stopped() {
		for {
			err := prg.device.Update()
			if err != nil {
				return err
			}
			prg.logger.Info("Device stopped, waiting for device to change")
			if prg.device.Stopped() {
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
	commandPath := prg.getCommandPath()
	if prg.config.Legacy {
		commandPath = prg.getLegacyCommandPath()
	}
	nodeCommand, err := prg.TheExecutable("node")
	if err != nil {
		return err
	}
	prg.cmd = exec.Command(nodeCommand, commandPath)
	prg.initCmd(prg.cmd)

	if fork {
		go prg.run()
	} else {
		prg.run()
	}

	return nil
}

func (prg *Program) run() {

	prg.logger.Info("Starting ", prg.config.DisplayName)

	if service.Interactive() {
		prg.cmd.Stderr = os.Stderr
		prg.cmd.Stdout = os.Stdout
	} else {
		if prg.config.Stderr != "" {
			stdErrFile, _ := prg.getStderrFile()
			defer stdErrFile.Close()
			prg.cmd.Stderr = stdErrFile
		}
		if prg.config.Stdout != "" {
			stdOutFile, _ := prg.getStdoutFile()
			defer stdOutFile.Close()
			prg.cmd.Stdout = stdOutFile
		}
	}

	prg.checkForChangesInterval()

	prg.timeStarted = time.Now()
	err := prg.cmd.Run()
	if err != nil {
		prg.logger.Warningf("Error running: %v", err)
	}
	prg.tryAgain()
}

func (prg *Program) tryAgain() {
	timeSinceStarted := time.Since(prg.timeStarted)
	if timeSinceStarted > time.Minute {
		prg.logger.Infof("Program ran for 1 minute, resetting backoff")
		prg.b.Reset()
	}
	duration := prg.b.Duration()
	prg.logger.Infof("Restarting in %v seconds", duration)
	time.Sleep(duration)
	prg.internalStart(false)
}

// Stop service but really
func (prg *Program) Stop(srv service.Service) error {
	close(prg.exit)
	prg.logger.Info("Stopping ", prg.config.DisplayName)

	err := prg.internalStop()
	if err != nil {
		return err
	}

	if service.Interactive() {
		os.Exit(0)
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

func (prg *Program) getLegacyCommandPath() string {
	return fmt.Sprintf(".%s%s", string(filepath.Separator), filepath.Join("node_modules", prg.getFullConnectorName(), "command.js"))
}

func (prg *Program) checkForChanges() error {
	err := prg.device.Update()
	if err != nil {
		prg.logger.Warningf("Device Update Error: %v", err.Error())
		return err
	}
	versionChange := prg.device.DidVersionChange()
	if versionChange {
		prg.logger.Infof("Device Version Change %v", prg.device.Version())
		err := prg.uc.DoBoth()
		if err != nil {
			return err
		}
	}
	stopChange := prg.device.DidStopChange()
	if stopChange {
		prg.logger.Infof("Device Stop Change")
		if prg.device.Stopped() {
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

func (prg *Program) checkForChangesInterval() {
	ticker := time.NewTicker(30 * time.Second)

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

func (prg *Program) getStderrFile() (*os.File, error) {
	file, err := os.OpenFile(prg.config.Stderr, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		prg.logger.Warningf("Failed to open stderr %q: %v", prg.config.Stderr, err)
		return nil, err
	}
	return file, nil
}

func (prg *Program) getStdoutFile() (*os.File, error) {
	file, err := os.OpenFile(prg.config.Stdout, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		prg.logger.Warningf("Failed to open stdout %q: %v", prg.config.Stdout, err)
		return nil, err
	}
	return file, nil
}

func (prg *Program) getEnv() []string {
	debug := SetEnv("DEBUG", "meshblu-*")
	pathEnv := GetPathEnv(prg.config.BinPath)
	return GetEnviron(debug, pathEnv)
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
