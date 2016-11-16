package updateconnector

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kardianos/osext"
	"github.com/spf13/afero"
)

// PackageConfig defines the update configuration reader / writer
type PackageConfig interface {
	Load() error
	GetTag() string
	Exists() (bool, error)
}

type packageJSON struct {
	Version string `json:"version"`
}

type pkgConfig struct {
	fs   afero.Fs
	json *packageJSON
	path string
}

// NewPackageConfig creates a new instance of PackageConfig
func NewPackageConfig(fs afero.Fs) (PackageConfig, error) {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	path, err := getPackgeConfigPath()
	if err != nil {
		return nil, err
	}
	return &pkgConfig{
		json: nil,
		fs:   fs,
		path: path,
	}, nil
}

// Load get the service config
func (c *pkgConfig) Load() error {
	exists, err := c.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("package.json does not exist")
	}
	f, err := c.fs.Open(c.path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := json.NewDecoder(f)
	err = r.Decode(&c.json)
	if err != nil {
		return err
	}
	return nil
}

func (c *pkgConfig) Exists() (bool, error) {
	exists, err := afero.Exists(c.fs, c.path)
	if err != nil {
		mainLogger.Error("package-config", "exists error", err)
		return false, err
	}
	if !exists {
		mainLogger.Info("package-config", "does not exist")
		return false, nil
	}
	return true, nil
}

func (c *pkgConfig) GetTag() string {
	return fmt.Sprintf("v%v", c.json.Version)
}

func getPackgeConfigPath() (string, error) {
	fullPath, err := osext.Executable()
	if err != nil {
		return "", err
	}
	dir, _ := filepath.Split(fullPath)
	return filepath.Join(dir, "package.json"), nil
}
