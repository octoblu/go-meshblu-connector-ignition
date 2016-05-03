package runner

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
)

// Config is the runner connector config structure.
type Config struct {
	ServiceName   string
	DisplayName   string
	Description   string
	ConnectorName string
	Legacy        bool
	Dir           string
	Env           []string

	Stderr, Stdout string
}

// GetConfig get the service config
func GetConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf := &Config{}

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

	dir, _ := filepath.Split(fullexecpath)

	return filepath.Join(dir, "service.json"), nil
}
