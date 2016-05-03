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
	prg.initCmd(prg.cmd)

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

	prg.setLogOnCmd(prg.cmd)

	err := prg.cmd.Start()
	if err != nil {
		prg.logger.Warningf("Error running (cmd.Start): %v", err)
	}
	err = prg.cmd.Wait()
	if err != nil {
		prg.logger.Warningf("Error running (cmd.Wait): %v", err)
	}
	return
}

// Stop service but really
func (prg *Program) Stop(srv service.Service) error {
	close(prg.exit)
	fmt.Println("Stopping")
	prg.logger.Info("Stopping ", prg.config.DisplayName)
	prg.cmd.Process.Kill()
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func (prg *Program) initCmd(cmd *exec.Cmd) {
	cmd.Dir = prg.config.Dir
	env := prg.getEnv()
	cmd.Env = env
}

func (prg *Program) setLogOnCmd(cmd *exec.Cmd) {
	if service.Interactive() {
		prg.cmd.Stderr = os.Stderr
		prg.cmd.Stdout = os.Stdout
		return
	}
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
}

func (prg *Program) getCommandPath() string {
	return path.Join("node_modules", "meshblu-connector-runner", "command.js")
}

func (prg *Program) getLegacyCommandPath() string {
	return path.Join("node_modules", prg.getFullConnectorName(), "command.js")
}

func (prg *Program) npmInstall() error {
	cmd := exec.Command("npm", "install", prg.getFullConnectorName())
	prg.initCmd(prg.cmd)
	prg.setLogOnCmd(prg.cmd)
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
	return append(prg.config.Env, debug)
}
