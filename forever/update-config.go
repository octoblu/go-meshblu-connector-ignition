package forever

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kardianos/osext"
	"github.com/spf13/afero"
)

type updateJSON struct {
	PID    int  `json:"pid"`
	Locked bool `json:"locked"`
}

func readConfig(path string) (*updateJSON, error) {
	fs := afero.NewOsFs()
	updateConfig := &updateJSON{}
	exists, err := afero.Exists(fs, path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return updateConfig, nil
	}
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := json.NewDecoder(f)
	err = r.Decode(updateConfig)
	if err != nil {
		return nil, err
	}
	return updateConfig, nil
}

func writePID(pid int) error {
	fs := afero.NewOsFs()
	path, err := getUpdateConfigPath()
	if err != nil {
		return err
	}
	updateConfig, err := readConfig(path)
	if updateConfig.Locked {
		return fmt.Errorf("PID is locked! You must exit")
	}
	updateConfig.PID = pid
	updateConfig.Locked = true
	jsonBytes, err := json.Marshal(updateConfig)
	if err != nil {
		return err
	}
	return afero.WriteFile(fs, path, jsonBytes, 0644)
}

func unlockPID() error {
	fs := afero.NewOsFs()
	path, err := getUpdateConfigPath()
	if err != nil {
		return err
	}
	updateConfig, err := readConfig(path)
	updateConfig.Locked = false
	jsonBytes, err := json.Marshal(updateConfig)
	if err != nil {
		return err
	}
	return afero.WriteFile(fs, path, jsonBytes, 0644)
}

func isLocked() (bool, error) {
	path, err := getUpdateConfigPath()
	if err != nil {
		return false, err
	}
	updateConfig, err := readConfig(path)
	if err != nil {
		return false, err
	}
	return updateConfig.Locked, err
}

func getPID() (int, error) {
	path, err := getUpdateConfigPath()
	if err != nil {
		return 0, err
	}
	updateConfig, err := readConfig(path)
	if err != nil {
		return 0, err
	}
	return updateConfig.PID, err
}

func getUpdateConfigPath() (string, error) {
	fullexecpath, err := osext.Executable()
	if err != nil {
		return "", err
	}
	dir, _ := filepath.Split(fullexecpath)
	return filepath.Join(dir, "update.json"), nil
}
