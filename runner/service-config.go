package runner

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
)

// ServiceConfig is the runner connector config structure.
type ServiceConfig struct {
	ConnectorName string
	Legacy        bool
	Dir           string
	Args          []string
	Env           []string

	Stderr, Stdout string
}

// GetServiceConfig get the service config
func GetServiceConfig() (*ServiceConfig, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf := &ServiceConfig{}

	r := json.NewDecoder(f)
	err = r.Decode(&conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func getConfigPath() (string, error) {
	fullexecpath, err := osext.Executable()
	if err != nil {
		return "", err
	}

	dir, execname := filepath.Split(fullexecpath)
	ext := filepath.Ext(execname)
	name := execname[:len(execname)-len(ext)]

	return filepath.Join(dir, name+".json"), nil
}
