package updateconnector

import (
	"fmt"
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
}

type updater struct {
	githubSlug    string
	connectorName string
	dir           string
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

// Do updates the connector
func (u *updater) Do(tag string) error {
	uri := u.getDownloadURI(tag)
	return extractor.New().DoWithURI(uri, u.dir)
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
