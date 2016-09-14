package updateconnector

import (
	"fmt"
	"os"
	"runtime"

	"github.com/octoblu/go-meshblu-connector-assembler/extractor"
	"github.com/spf13/afero"
)

// UpdateConnector is an interface to handing updating the connector files
type UpdateConnector interface {
	NeedsUpdate(tag string) (bool, error)
	Do(tag string) error
}

type updater struct {
	githubSlug    string
	connectorName string
	dir           string
	updateConfig  UpdateConfig
	fs            afero.Fs
}

// New returns an instance of the UpdateConnector
func New(githubSlug, connectorName, dir string, fs afero.Fs) (UpdateConnector, error) {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	updateConfig, err := NewUpdateConfig(fs)
	if err != nil {
		return nil, err
	}
	return &updater{
		githubSlug:    githubSlug,
		connectorName: connectorName,
		dir:           dir,
		fs:            fs,
		updateConfig:  updateConfig,
	}, nil
}

// NeedsUpdate returns if the connector needs to updated
func (u *updater) NeedsUpdate(tag string) (bool, error) {
	err := u.updateConfig.Load()
	if err != nil {
		return false, err
	}
	if u.updateConfig.GetTag() == tag {
		return false, nil
	}
	return true, nil
}

// Dont updates the config file, but does not update the connector
func (u *updater) Dont(tag string, pid int) error {
	return u.updateConfig.Write(tag, pid)
}

// Do updates the connector
func (u *updater) Do(tag string) error {
	uri := u.getDownloadURI(tag)
	err := extractor.New().DoWithURI(uri, u.dir)
	if err != nil {
		return err
	}
	pid := os.Getpid()
	return u.updateConfig.Write(tag, pid)
}

func (u *updater) getDownloadURI(tag string) string {
	baseURI := fmt.Sprintf("https://github.com/%s/releases/download", u.githubSlug)
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	fileName := fmt.Sprintf("%s-%s-%s.%s", u.connectorName, runtime.GOOS, runtime.GOARCH, ext)
	return fmt.Sprintf("%s/%s/%s", baseURI, tag, fileName)
}
