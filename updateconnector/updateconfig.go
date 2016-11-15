package updateconnector

import (
	"encoding/json"
	"path/filepath"

	"github.com/kardianos/osext"
	"github.com/spf13/afero"
)

// UpdateConfig defines the update configuration reader / writer
type UpdateConfig interface {
	Load() error
	Write(pid int) error
}

type updateJSON struct {
	Pid int
}

type config struct {
	fs   afero.Fs
	json *updateJSON
	path string
}

// NewUpdateConfig creates a new instance of UpdateConfig
func NewUpdateConfig(fs afero.Fs) (UpdateConfig, error) {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	json := &updateJSON{
		Pid: 0,
	}
	path, err := getUpdateConfigPath()
	if err != nil {
		return nil, err
	}
	return &config{
		json: json,
		fs:   fs,
		path: path,
	}, nil
}

// Load get the service config
func (c *config) Load() error {
	exists, err := afero.Exists(c.fs, c.path)
	if err != nil {
		return err
	}
	if !exists {
		return nil
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

func (c *config) Write(pid int) error {
	c.json.Pid = pid
	jsonBytes, err := json.Marshal(c.json)
	if err != nil {
		return err
	}
	return afero.WriteFile(c.fs, c.path, jsonBytes, 0644)
}

func getUpdateConfigPath() (string, error) {
	fullexecpath, err := osext.Executable()
	if err != nil {
		return "", err
	}

	dir, _ := filepath.Split(fullexecpath)

	return filepath.Join(dir, "update.json"), nil
}
