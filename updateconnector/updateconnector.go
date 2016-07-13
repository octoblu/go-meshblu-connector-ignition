package updateconnector

import (
	"fmt"

	"github.com/octoblu/go-meshblu-connector-ignition/runner"
	"github.com/spf13/afero"
)

// UpdateConnector is an interface to handing updating the connector files
type UpdateConnector interface {
	NeedsUpdate(tag string) (bool, error)
}

type updater struct {
	config       *runner.Config
	updateConfig UpdateConfig
	fs           afero.Fs
}

// New returns an instance of the UpdateConnector
func New(config *runner.Config, fs afero.Fs) (UpdateConnector, error) {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	updateConfig, err := NewUpdateConfig(fs)
	if err != nil {
		fmt.Println("err config", err)
		return nil, err
	}
	err = updateConfig.Load()
	if err != nil {
		fmt.Println("err load", err)
		return nil, err
	}
	return &updater{
		config:       config,
		fs:           fs,
		updateConfig: updateConfig,
	}, nil
}

// NeedsUpdate returns if the connector needs to updated
func (u *updater) NeedsUpdate(tag string) (bool, error) {
	if u.updateConfig.GetTag() == tag {
		return false, nil
	}
	return true, nil
}
