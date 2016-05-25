package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/kardianos/service"
	"github.com/octoblu/go-meshblu-connector-assembler/extractor"
)

// UpdateConnector defines the struct for updating the connector
type UpdateConnector struct {
	prg *Program
}

// NewUpdateConnector creates an instance of the UpdateConnector
func NewUpdateConnector(prg *Program) *UpdateConnector {
	return &UpdateConnector{prg}
}

// DoBoth handles either the legacy or new
func (uc *UpdateConnector) DoBoth() error {
	if uc.prg.config.Legacy {
		return uc.DoLegacy()
	}
	return uc.Do()
}

// Do downloads and extracts the update
func (uc *UpdateConnector) Do() error {
	err := uc.prg.internalStop()
	if err != nil {
		return err
	}
	uc.prg.logger.Info("Updating Connector")
	err = extractor.New().DoWithURI(uc.getConnectorURI(), uc.prg.config.Dir)
	if err != nil {
		return err
	}
	err = uc.prg.internalStart(false)
	if err != nil {
		return err
	}
	return nil
}

// DoLegacy downloads and extracts the update
func (uc *UpdateConnector) DoLegacy() error {
	uc.prg.logger.Info("Updating Legacy Connector")
	err := uc.prg.internalStop()
	if err != nil {
		return err
	}

	err = uc.runInstall()
	if err != nil {
		err = uc.clearModules()
		if err != nil {
			return err
		}
		err = uc.runInstall()
	}
	if err != nil {
		return err
	}

	err = uc.prg.internalStart(true)
	if err != nil {
		return err
	}
	return nil
}

func (uc *UpdateConnector) runInstall() error {
	prg := uc.prg
	npmCommand, err := prg.TheExecutable("npm")
	if err != nil {
		return err
	}
	version := prg.device.Version()
	connectorWithVersion := prg.getFullConnectorName()
	if version == "" {
		connectorWithVersion = fmt.Sprintf("%s@%s", connectorWithVersion, version)
	}

	cmd := exec.Command(npmCommand, "install", connectorWithVersion)
	prg.initCmd(cmd)
	if service.Interactive() {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	} else {
		if prg.config.Stderr != "" {
			stdErrFile, _ := prg.getStderrFile()
			defer stdErrFile.Close()
			cmd.Stderr = stdErrFile
		}
		if prg.config.Stdout != "" {
			stdOutFile, _ := prg.getStdoutFile()
			defer stdOutFile.Close()
			cmd.Stdout = stdOutFile
		}
	}

	err = cmd.Run()
	if err != nil {
		prg.logger.Warningf("Error running npm: %v", err)
		return err
	}
	return nil
}

func (uc *UpdateConnector) clearModules() error {
	uc.prg.logger.Info("Removing node_modules and trying again...")
	return os.RemoveAll(filepath.Join(uc.prg.config.Dir, "node_modules"))
}

// getConnectorURI gets the OS specific connector path
func (uc *UpdateConnector) getConnectorURI() string {
	config := uc.prg.config
	baseURI := fmt.Sprintf("https://github.com/%s/releases/download", config.GithubSlug)
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	fileName := fmt.Sprintf("%s-%s-%s.%s", config.ConnectorName, runtime.GOOS, runtime.GOARCH, ext)
	return fmt.Sprintf("%s/%s/%s", baseURI, uc.prg.device.VersionWithV(), fileName)
}
