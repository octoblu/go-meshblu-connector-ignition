package updateconnector

import (
	"fmt"
	"os"
	"runtime"

	mainlogger "github.com/azer/logger"
	"github.com/octoblu/go-meshblu-connector-assembler/extractor"
	"github.com/spf13/afero"
)

var log = mainlogger.New("updateconnector")

// UpdateConnector is an interface to handing updating the connector files
type UpdateConnector interface {
	NeedsUpdate(tag string) (bool, error)
	Do(tag string) error
	WritePID() error
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
	if fs == nil {
		fs = afero.NewOsFs()
	}
	updateConfig, err := NewUpdateConfig(fs)
	if err != nil {
		log.Error("Error creating UpdateConfig")
		return nil, err
	}
	packageConfig, err := NewPackageConfig(fs)
	if err != nil {
		log.Error("Error creating PackgeConfig")
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
		log.Error("package.json exists err %v", err.Error())
		return false, err
	}
	if !exists {
		log.Error("package.json does not exist")
		return true, err
	}
	err = packageConfig.Load()
	if err != nil {
		log.Error("package.json load error %v", err.Error())
		return false, err
	}
	if packageConfig.GetTag() == tag {
		log.Info("package.version is the same (%s)", tag)
		return false, nil
	}
	log.Info("package needs update %v != %v", packageConfig.GetTag(), tag)
	return true, nil
}

// Dont updates the config file, but does not update the connector
func (u *updater) WritePID() error {
	pid := os.Getpid()
	log.Info("WritePID - writing pid %v", pid)
	return u.updateConfig.Write(pid)
}

// Do updates the connector
func (u *updater) Do(tag string) error {
	uri := u.getDownloadURI(tag)
	err := extractor.New().DoWithURI(uri, u.dir)
	if err != nil {
		log.Error("Do - extraction error %v", err.Error())
		return err
	}
	pid := os.Getpid()
	log.Info("Do - writing pid %v", pid)
	return u.updateConfig.Write(pid)
}

func (u *updater) getDownloadURI(tag string) string {
	baseURI := fmt.Sprintf("https://github.com/%s/releases/download", u.githubSlug)
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	fileName := fmt.Sprintf("%s-%s-%s.%s", u.connectorName, runtime.GOOS, runtime.GOARCH, ext)
	url := fmt.Sprintf("%s/%s/%s", baseURI, tag, fileName)
	log.Info("download uri %v", url)
	return url
}
