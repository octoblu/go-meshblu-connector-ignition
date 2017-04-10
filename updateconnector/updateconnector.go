package updateconnector

import (
	"fmt"
	"os"
	"runtime"

	"github.com/octoblu/go-meshblu-connector-assembler/extractor"
	"github.com/octoblu/go-meshblu-connector-ignition/logger"
	"github.com/spf13/afero"
)

var mainLogger logger.MainLogger

// UpdateConnector is an interface to handing updating the connector files
type UpdateConnector interface {
	// NeedsUpdate returns true if the connector needs to updated
	NeedsUpdate(tag string) (bool, error)

	// Do updates the connector
	Do(tag string) error

	// WritePID updates the PID
	WritePID() error

	// ClearPID clears the PID
	ClearPID() error
}

type updater struct {
	githubSlug    string
	connectorName string
	dir           string
	updateConfig  UpdateConfig
	packageConfig PackageConfig
	fs            afero.Fs
}

// New returns an instance of the UpdateConnector
func New(githubSlug, connectorName, dir string, fs afero.Fs, fakeMainLogger logger.MainLogger) (UpdateConnector, error) {
	if mainLogger == nil {
		if fakeMainLogger != nil {
			mainLogger = fakeMainLogger
		} else {
			mainLogger = logger.GetMainLogger()
		}
	}
	if fs == nil {
		fs = afero.NewOsFs()
	}
	updateConfig, err := NewUpdateConfig(fs)
	if err != nil {
		mainLogger.Error("updateconnector", "Error creating UpdateConfig", err)
		return nil, err
	}
	packageConfig, err := NewPackageConfig(fs)
	if err != nil {
		mainLogger.Error("updateconnector", "Error creating PackgeConfig", err)
		return nil, err
	}
	return &updater{
		githubSlug:    githubSlug,
		connectorName: connectorName,
		dir:           dir,
		fs:            fs,
		updateConfig:  updateConfig,
		packageConfig: packageConfig,
	}, nil
}

// NeedsUpdate returns if the connector needs to updated
func (u *updater) NeedsUpdate(tag string) (bool, error) {
	packageConfig := u.packageConfig
	exists, err := packageConfig.Exists()
	if err != nil {
		mainLogger.Error("updateconnector", "package.json exists error", err)
		return false, err
	}
	if !exists {
		mainLogger.Info("updateconnector", "package.json does not exist")
		return true, nil
	}
	err = packageConfig.Load()
	if err != nil {
		mainLogger.Error("updateconnector", "package.json load error", err)
		return true, err
	}
	if packageConfig.GetTag() == tag {
		mainLogger.Info("updateconnector", fmt.Sprintf("package.version is the same (%s)", tag))
		return false, nil
	}
	mainLogger.Info("updateconnector", fmt.Sprintf("package needs update %v != %v", packageConfig.GetTag(), tag))
	return true, nil
}

// WritePID updates the PID
func (u *updater) WritePID() error {
	pid := os.Getpid()
	mainLogger.Info("updateconnector.WritePID", fmt.Sprintf("writing pid %v", pid))
	err := u.updateConfig.Write(pid)
	if err != nil {
		mainLogger.Error("updateconnector.WritePID", "Error", err)
		return err
	}
	return nil
}

// ClearPID clears the PID
func (u *updater) ClearPID() error {
	mainLogger.Info("updateconnector.ClearPID", "clearing pid")
	err := u.updateConfig.Write(0)
	if err != nil {
		mainLogger.Error("updateconnector.ClearPID", "Error", err)
		return err
	}
	return nil
}

// Do updates the connector
func (u *updater) Do(tag string) error {
	uri := u.getDownloadURI(tag)
	err := extractor.New().DoWithURI(uri, u.dir)
	if err != nil {
		mainLogger.Error("updateconnector", "Extraction error", err)
		return err
	}
	return u.WritePID()
}

func (u *updater) getDownloadURI(tag string) string {
	baseURI := fmt.Sprintf("https://github.com/%s/releases/download", u.githubSlug)
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	fileName := fmt.Sprintf("%s-%s-%s.%s", u.connectorName, runtime.GOOS, runtime.GOARCH, ext)
	url := fmt.Sprintf("%s/%s/%s", baseURI, tag, fileName)
	mainLogger.Info("updateconnector", fmt.Sprintf("download uri %v", url))
	return url
}
