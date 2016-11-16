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
		fmt.Println(fmt.Sprintf("[update-connector] Error creating UpdateConfig"))
		return nil, err
	}
	packageConfig, err := NewPackageConfig(fs)
	if err != nil {
		fmt.Println(fmt.Sprintf("[update-connector] Error creating PackgeConfig"))
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
		fmt.Println(fmt.Sprintf("[update-connector] package.json exists err %v", err.Error()))
		return false, err
	}
	if !exists {
		fmt.Println(fmt.Sprintf("[update-connector] package.json does not exist"))
		return true, err
	}
	err = packageConfig.Load()
	if err != nil {
		fmt.Println(fmt.Sprintf("[update-connector] package.json load error %v", err.Error()))
		return false, err
	}
	if packageConfig.GetTag() == tag {
		fmt.Println(fmt.Sprintf("[update-connector] package.version is the same (%s)", tag))
		return false, nil
	}
	fmt.Println(fmt.Sprintf("[update-connector] package needs update %v != %v", packageConfig.GetTag(), tag))
	return true, nil
}

// Dont updates the config file, but does not update the connector
func (u *updater) WritePID() error {
	pid := os.Getpid()
	fmt.Println(fmt.Sprintf("[update-connector] WritePID - writing pid %v", pid))
	return u.updateConfig.Write(pid)
}

// Do updates the connector
func (u *updater) Do(tag string) error {
	uri := u.getDownloadURI(tag)
	err := extractor.New().DoWithURI(uri, u.dir)
	if err != nil {
		fmt.Println(fmt.Sprintf("[update-connector] Do - extraction error %v", err.Error()))
		return err
	}
	pid := os.Getpid()
	fmt.Println(fmt.Sprintf("[update-connector] Do - writing pid %v", pid))
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
	fmt.Println(fmt.Sprintf("[update-connector] download uri %v", url))
	return url
}
