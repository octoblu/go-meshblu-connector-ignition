package status

import (
	"bytes"
	"strings"

	"github.com/octoblu/go-meshblu/http/meshblu"
)

// Client defines the meshblu info
type Client struct {
	meshbluClient meshblu.Meshblu
	meshbluDevice *MeshbluDevice
	uuid          string
}

// Status defines the device management interface
type Status interface {
	Fetch() error
	UpdateErrors(data []byte) error
	ResetErrors() error
}

// New creates a new status struct
func New(meshbluClient meshblu.Meshblu, uuid string) (Status, error) {
	device := &Client{
		meshbluClient: meshbluClient,
		uuid:          uuid,
	}
	return device, nil
}

// Fetch updates the device
func (client *Client) Fetch() error {
	if client.uuid == "" {
		return nil
	}
	data, err := client.meshbluClient.GetDevice(client.uuid)
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

// ResetErrors removes the error array from the device
func (client *Client) ResetErrors() error {
	if client.uuid == "" {
		return nil
	}
	body, err := NewUpdateErrorsBody([]string{})
	if err != nil {
		return err
	}
	_, err = client.meshbluClient.UpdateDevice(client.uuid, body)
	return err
}

// UpdateErrors updates the status device with the errors
func (client *Client) UpdateErrors(data []byte) error {
	if client.uuid == "" {
		return nil
	}
	if data == nil {
		return nil
	}
	buf := bytes.NewBuffer(data)
	errors := strings.Split(buf.String(), "\n")
	body, err := NewUpdateErrorsBody(errors)
	if err != nil {
		return err
	}
	_, err = client.meshbluClient.UpdateDevice(client.uuid, body)
	return err
}
