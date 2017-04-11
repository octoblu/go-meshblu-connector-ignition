package forever

import (
	"encoding/json"
	"path/filepath"

	"github.com/kardianos/osext"
	"github.com/spf13/afero"
)

type updateJSON struct {
	PID int `json:"Pid"`
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
	updateConfig.PID = pid
	jsonBytes, err := json.Marshal(updateConfig)
	if err != nil {
		return err
	}
	return afero.WriteFile(fs, path, jsonBytes, 0644)
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
