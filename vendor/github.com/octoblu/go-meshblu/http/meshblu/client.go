package meshblu

import (
	"github.com/octoblu/go-meshblu/config"
)

// NewClient creates a new device struct
func NewClient(configPath string) (Meshblu, string, error) {
	config, err := config.ReadFromConfig(configPath)
	if err != nil {
		return nil, "", err
	}
	url, err := config.ToURL()
	if err != nil {
		return nil, "", err
	}
	meshbluClient, err := Dial(url)
	if err != nil {
		return nil, "", err
	}
	meshbluClient.SetAuth(config.UUID, config.Token)
	return meshbluClient, config.UUID, nil
}
