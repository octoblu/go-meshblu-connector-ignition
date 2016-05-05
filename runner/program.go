package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
	nodeCommand, err := prg.theExecutable("node")
	if err != nil {
		return err
	}
	prg.cmd = exec.Command(nodeCommand, commandPath)
	prg.initCmd(prg.cmd)

	go prg.run()

	return nil
}

func (prg *Program) run() {
	prg.logger.Info("Starting ", prg.config.DisplayName)

	if service.Interactive() {
		prg.cmd.Stderr = os.Stderr
		prg.cmd.Stdout = os.Stdout
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
	if prg.cmd != nil {
		prg.cmd.Process.Kill()
	}

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

func (prg *Program) getCommandPath() string {
	return fmt.Sprintf(".%s%s", string(filepath.Separator), filepath.Join("node_modules", "meshblu-connector-runner", "command.js"))
}

func (prg *Program) getLegacyCommandPath() string {
	return fmt.Sprintf(".%s%s", string(filepath.Separator), filepath.Join("node_modules", prg.getFullConnectorName(), "command.js"))
}

func (prg *Program) npmInstall() error {
	npmCommand, err := prg.theExecutable("npm")
	if err != nil {
		return err
	}
	cmd := exec.Command(npmCommand, "install", prg.getFullConnectorName())
	prg.initCmd(cmd)
	if service.Interactive() {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		return nil
	}
	if prg.config.Stderr != "" {
		stdErrFile, err := os.OpenFile(prg.config.Stderr, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			prg.logger.Warningf("Failed to open std err %q: %v", prg.config.Stderr, err)
			return nil
		}
		defer stdErrFile.Close()
		cmd.Stderr = stdErrFile
	}
	if prg.config.Stdout != "" {
		stdOutFile, err := os.OpenFile(prg.config.Stdout, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			prg.logger.Warningf("Failed to open std out %q: %v", prg.config.Stdout, err)
			return nil
		}
		defer stdOutFile.Close()
		cmd.Stdout = stdOutFile
	}

	return cmd.Run()
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
	binPathEnv := fmt.Sprintf("PATH=%s", prg.config.BinPath)
	return append(os.Environ(), binPathEnv, debug)
}

func (prg *Program) theExecutable(name string) (string, error) {
	file, err := exec.LookPath(filepath.Join(prg.config.BinPath, name))
	if err != nil {
		prg.logger.Warningf("Failed to get Executable File, %s - Error %v", file, err)
		return "", err
	}
	prg.logger.Infof("got executable %s", file)
	return file, nil
}
