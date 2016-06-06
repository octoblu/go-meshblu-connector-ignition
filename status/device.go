package status

import (
	"github.com/octoblu/go-meshblu/config"
	"github.com/octoblu/go-meshblu/http/meshblu"
)

// Client defines the meshblu info
type Client struct {
	config        *config.Config
	meshbluClient meshblu.Meshblu
	meshbluDevice *MeshbluDevice
}

// Device defines the device management interface
type Device interface {
	Update() error
}

// New creates a new device struct
func New(configPath, tag string) (Device, error) {
	config, err := config.ReadFromConfig(configPath)
	if err != nil {
		return nil, err
	}
	url, err := config.ToURL()
	if err != nil {
		return nil, err
	}
	meshbluClient, err := meshblu.Dial(url)
	if err != nil {
		return nil, err
	}
	meshbluClient.SetAuth(config.UUID, config.Token)

	device := &Client{
		config:        config,
		meshbluClient: meshbluClient,
	}
	return device, nil
}

// Update updates the device
func (client *Client) Update() error {
	data, err := client.meshbluClient.GetDevice(client.config.UUID)
	if err != nil {
		return err
	}
	meshbluDevice, err := ParseMeshbluDevice(data)
	if err != nil {
		return err
	}
	client.meshbluDevice = meshbluDevice
	return nil
}
