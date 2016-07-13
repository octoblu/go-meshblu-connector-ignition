package updateconnector

import "github.com/octoblu/go-meshblu-connector-ignition/runner"

// UpdateConnector is an interface to handing updating the connector files
type UpdateConnector interface {
	NeedsUpdate() (bool, error)
}

type updater struct {
	config *runner.Config
}

// New returns an instance of the UpdateConnector
func New(config *runner.Config) (UpdateConnector, error) {
	return &updater{config}, nil
}

func (u *updater) NeedsUpdate() (bool, error) {
	return true, nil
}
