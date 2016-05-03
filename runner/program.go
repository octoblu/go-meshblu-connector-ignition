package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/kardianos/service"
)

// Program inteface that is real
type Program struct {
	config *Config
	srv    service.Service
	logger service.Logger
	cmd    *exec.Cmd
	exit   chan struct{}
}

// Start service but really
func (prg *Program) Start(srv service.Service) error {
	commandPath := prg.getCommandPath()
	if prg.config.Legacy {
		err := prg.npmInstall()
		if err != nil {
			return err
		}
		commandPath = prg.getLegacyCommandPath()
	}
	prg.cmd = exec.Command(commandPath, prg.config.Args...)
	prg.cmd.Env = prg.getEnv()

	go prg.run()

	return nil
}

func (prg *Program) run() {
	prg.logger.Info("Starting ", prg.config.DisplayName)
	defer func() {
		if service.Interactive() {
			prg.Stop(prg.srv)
		} else {
			prg.srv.Stop()
		}
	}()

	if prg.config.Stderr != "" {
		stdErrFile, err := os.OpenFile(prg.config.Stderr, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			prg.logger.Warningf("Failed to open std err %q: %v", prg.config.Stderr, err)
			return
		}
		defer stdErrFile.Close()
		prg.cmd.Stderr = stdErrFile
	}
	if prg.config.Stdout != "" {
		stdOutFile, err := os.OpenFile(prg.config.Stdout, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			prg.logger.Warningf("Failed to open std out %q: %v", prg.config.Stdout, err)
			return
		}
		defer stdOutFile.Close()
		prg.cmd.Stdout = stdOutFile
	}

	err := prg.cmd.Run()
	if err != nil {
		prg.logger.Warningf("Error running: %v", err)
	}

	return
}

// Stop service but really
func (prg *Program) Stop(srv service.Service) error {
	close(prg.exit)
	prg.logger.Info("Stopping ", prg.config.DisplayName)
	if prg.cmd.ProcessState.Exited() == false {
		prg.cmd.Process.Kill()
	}
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func (prg *Program) getCommandPath() string {
	return path.Join("node_modules", "meshblu-connector-runner", "command.js")
}

func (prg *Program) getLegacyCommandPath() string {
	return path.Join("node_modules", prg.getFullConnectorName(), "command.js")
}

func (prg *Program) npmInstall() error {
	cmd := exec.Command("npm", "install", prg.getFullConnectorName())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func (prg *Program) getFullConnectorName() string {
	return fmt.Sprintf("meshblu-%s", prg.config.ConnectorName)
}

func (prg *Program) getEnv() []string {
	debug := os.Getenv("DEBUG")
	if debug == "" {
		if prg.config.Legacy {
			debug = fmt.Sprintf("DEBUG=%s", prg.getFullConnectorName())
		} else {
			debug = fmt.Sprintf("DEBUG=%s", "meshblu-connector-*")
		}
	}
	env := append(os.Environ(), debug)
	return append(env, prg.config.Env...)
}
