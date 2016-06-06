package connector

import (
	"fmt"
	"strings"

	"github.com/octoblu/go-meshblu/config"
	"github.com/octoblu/go-meshblu/http/meshblu"
)

// Client defines the meshblu info
type Client struct {
	config        *config.Config
	meshbluClient meshblu.Meshblu
	device        *MeshbluDevice
	lastDevice    *MeshbluDevice
	tag           string
}

// Connector defines the device management interface
type Connector interface {
	Fetch() error
	DidVersionChange() bool
	DidStopChange() bool
	Stopped() bool
	Version() string
	VersionWithV() string
}

// New creates a new device struct
func New(configPath, tag string) (Connector, error) {
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
		tag:           tag,
	}
	return device, nil
}

// Fetch updates the device
func (client *Client) Fetch() error {
	data, err := client.meshbluClient.GetDevice(client.config.UUID)
	if err != nil {
		return err
	}
	device, err := ParseMeshbluDevice(data, client.tag)
	if err != nil {
		return err
	}
	if client.device != nil {
		client.lastDevice = CopyMeshbluDevice(client.device)
	}
	client.device = device
	return nil
}

// DidVersionChange checks to see the version changed from the last fetch
func (client *Client) DidVersionChange() bool {
	if client.lastDevice == nil {
		return false
	}
	last := client.lastDevice.Metadata.Version
	current := client.device.Metadata.Version
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
	last := client.lastDevice.Metadata.Stopped
	current := client.device.Metadata.Stopped
	if last == current {
		return false
	}
	return true
}

// Stopped return the boolean true if the connector stopped
func (client *Client) Stopped() bool {
	return client.device.Metadata.Stopped
}

// Version return connector version
func (client *Client) Version() string {
	version := client.device.Metadata.Version
	return strings.Replace(version, "v", "", 1)
}

// VersionWithV return connector version
func (client *Client) VersionWithV() string {
	version := client.Version()
	return fmt.Sprintf("v%s", version)
}
