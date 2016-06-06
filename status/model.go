package status

import "encoding/json"

// MeshbluDevice defines the meshblu device
type MeshbluDevice struct {
}

// ParseMeshbluDevice creates a device from a JSON byte array
func ParseMeshbluDevice(data []byte) (*MeshbluDevice, error) {
	device := &MeshbluDevice{}
	err := json.Unmarshal(data, device)
	return device, err
}
