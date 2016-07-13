package runner

import (
	"fmt"
	"runtime"

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

// NeedsUpdate will check to see what version is installed and running
func (uc *UpdateConnector) NeedsUpdate() (bool, error) {
	return true, nil
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

func (uc *UpdateConnector) getConnectorURI() string {
	config := uc.prg.config
	baseURI := fmt.Sprintf("https://github.com/%s/releases/download", config.GithubSlug)
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	fileName := fmt.Sprintf("%s-%s-%s.%s", config.ConnectorName, runtime.GOOS, runtime.GOARCH, ext)
	return fmt.Sprintf("%s/%s/%s", baseURI, uc.prg.connector.VersionWithV(), fileName)
}
