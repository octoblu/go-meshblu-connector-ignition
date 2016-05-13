package device

import (
	"strings"

	"github.com/octoblu/go-meshblu/config"
	"github.com/octoblu/go-meshblu/http/meshblu"
)

// Client defines the meshblu info
type Client struct {
	config        *config.Config
	meshbluClient meshblu.Meshblu
	meshbluDevice *MeshbluDevice
	lastDevice    *MeshbluDevice
}

// Device defines the device management interface
type Device interface {
	Update() error
	DidVersionChange() bool
	DidStopChange() bool
	Stopped() bool
	Version() string
}

// New creates a new device struct
func New(configPath string) (Device, error) {
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
	if client.meshbluDevice != nil {
		client.lastDevice = CopyMeshbluDevice(client.meshbluDevice)
	}
	client.meshbluDevice = meshbluDevice
	return nil
}

// DidVersionChange checks to see the version changed from the last fetch
func (client *Client) DidVersionChange() bool {
	if client.lastDevice == nil {
		return false
	}
	last := client.lastDevice.Connector.Version
	current := client.meshbluDevice.Connector.Version
	if last == current {
		return false
	}
	return true
}

// DidStopChange checks to see the version changed from the last fetch
func (client *Client) DidStopChange() bool {
	if client.lastDevice == nil {
		return false
	}
	last := client.lastDevice.Connector.Stopped
	current := client.meshbluDevice.Connector.Stopped
	if last == current {
		return false
	}
	return true
}

// Stopped return the boolean true if the connector stopped
func (client *Client) Stopped() bool {
	return client.meshbluDevice.Connector.Stopped
}

// Version return connector version
func (client *Client) Version() string {
	version := client.meshbluDevice.Connector.Version
	return strings.Replace(version, "v", "", 1)
}
