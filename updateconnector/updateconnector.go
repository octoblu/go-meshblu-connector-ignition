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
	NeedsUpdate(tag string) (bool, error)
	Do(tag string) error
	WritePID() error
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
func New(githubSlug, connectorName, dir string, fs afero.Fs) (UpdateConnector, error) {
	if mainLogger == nil {
		mainLogger = logger.GetMainLogger()
	}
	if fs == nil {
		fs = afero.NewOsFs()
	}
	updateConfig, err := NewUpdateConfig(fs)
	if err != nil {
		mainLogger.Error("update-connector", "Error creating UpdateConfig", err)
		return nil, err
	}
	packageConfig, err := NewPackageConfig(fs)
	if err != nil {
		mainLogger.Error("update-connector", "Error creating PackgeConfig", err)
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
		mainLogger.Error("update-connector", "package.json exists", err)
		return false, err
	}
	if !exists {
		mainLogger.Error("update-connector", "package.json does not exist", err)
		return true, fmt.Errorf("package.json does not exist")
	}
	err = packageConfig.Load()
	if err != nil {
		mainLogger.Error("update-connector", "package.json load error", err)
		return false, err
	}
	if packageConfig.GetTag() == tag {
		mainLogger.Info("update-connector", fmt.Sprintf("package.version is the same (%s)", tag))
		return false, nil
	}
	mainLogger.Info("update-connector", fmt.Sprintf("package needs update %v != %v", packageConfig.GetTag(), tag))
	return true, nil
}

// WritePID updates the PID
func (u *updater) WritePID() error {
	pid := os.Getpid()
	mainLogger.Info("update-connector", fmt.Sprintf("WritePID - writing pid %v", pid))
	return u.updateConfig.Write(pid)
}

// ClearPID clears the PID
func (u *updater) ClearPID() error {
	mainLogger.Info("update-connector", "ClearPID")
	return u.updateConfig.Write(0)
}

// Do updates the connector
func (u *updater) Do(tag string) error {
	uri := u.getDownloadURI(tag)
	err := extractor.New().DoWithURI(uri, u.dir)
	if err != nil {
		mainLogger.Error("update-connector", "Extraction error", err)
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
	mainLogger.Info("update-connector", fmt.Sprintf("download uri %v", url))
	return url
}
