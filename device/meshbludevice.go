package device

import "encoding/json"

// Connector defines the connector metadata
type Connector struct {
	Stopped bool   `json:"stopped"`
	Version string `json:"version"`
}

// MeshbluDevice defines the meshblu device
type MeshbluDevice struct {
	*Connector `json:"connectorMetadata"`
}

// ParseMeshbluDevice creates a device from a JSON byte array
func ParseMeshbluDevice(data []byte) (*MeshbluDevice, error) {
	device := &MeshbluDevice{}
	err := json.Unmarshal(data, device)
	return device, err
}

// CopyMeshbluDevice creates a copy of the device passed in
func CopyMeshbluDevice(orgDevice *MeshbluDevice) *MeshbluDevice {
	connector := &Connector{
		Stopped: orgDevice.Connector.Stopped,
		Version: orgDevice.Connector.Version,
	}
	device := &MeshbluDevice{
		connector,
	}
	return device
}
