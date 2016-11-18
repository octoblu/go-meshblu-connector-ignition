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
	uuid          string
	statusUUID    string
}

// Connector defines the device management interface
type Connector interface {
	Fetch() error
	DidVersionChange() bool
	DidStopChange() bool
	StatusUUID() string
	Stopped() bool
	Version() string
	VersionWithV() string
}

// New creates a new device struct
func New(meshbluClient meshblu.Meshblu, uuid, tag string) (Connector, error) {
	device := &Client{
		meshbluClient: meshbluClient,
		uuid:          uuid,
		tag:           tag,
		statusUUID:    "",
	}
	return device, nil
}

// Fetch updates the local device with latest from remote
func (client *Client) Fetch() error {
	data, err := client.meshbluClient.GetDevice(client.uuid)
	if err != nil {
		return err
	}
	device, err := ParseMeshbluDevice(data, client.tag)
	if err != nil {
		return err
	}
	statusUUID, err := ParseMeshbluDeviceForStatusUUID(data)
	if err != nil {
		return err
	}
	client.statusUUID = statusUUID
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

// StatusUUID gets the uuid of the status device attached to the connector
func (client *Client) StatusUUID() string {
	return client.statusUUID
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
