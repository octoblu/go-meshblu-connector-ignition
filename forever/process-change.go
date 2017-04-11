package forever

import (
	"os"

	"github.com/octoblu/go-meshblu-connector-ignition/updateconnector"
)

func sameProcess() (bool, error) {
	config, err := updateconnector.NewUpdateConfig(nil)
	if err != nil {
		return false, err
	}
	err = config.Load()
	if err != nil {
		return false, err
	}
	pid := os.Getpid()
	if config.ReadPID() == pid {
		return true, nil
	}
	return false, nil
}
