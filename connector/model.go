package connector

import "encoding/json"

// Metadata defines the connector metadata
type Metadata struct {
	Stopped bool   `json:"stopped"`
	Version string `json:"version"`
}

// MeshbluDevice defines the meshblu device
type MeshbluDevice struct {
	*Metadata `json:"connectorMetadata"`
}

// ParseMeshbluDevice creates a device from a JSON byte array
func ParseMeshbluDevice(data []byte, defaultVersion string) (*MeshbluDevice, error) {
	device := &MeshbluDevice{}
	err := json.Unmarshal(data, device)
	if device.Metadata == nil {
		device.Metadata = &Metadata{
			Version: defaultVersion,
			Stopped: false,
		}
	}
	return device, err
}

// CopyMeshbluDevice creates a copy of the device passed in
func CopyMeshbluDevice(orgMeshbluDevice *MeshbluDevice) *MeshbluDevice {
	connector := &Metadata{
		Stopped: orgMeshbluDevice.Metadata.Stopped,
		Version: orgMeshbluDevice.Metadata.Version,
	}
	device := &MeshbluDevice{
		connector,
	}
	return device
}
