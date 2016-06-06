package status

import (
	"bytes"
	"encoding/json"
	"io"
	"time"
)

// MeshbluDevice defines the meshblu device
type MeshbluDevice struct {
}

// UpdateDeviceErrors defines the update properties
type UpdateDeviceErrors struct {
	UpdateErrorsAt int64    `json:"updateErrorsAt"`
	Errors         []string `json:"errors"`
}

// ParseMeshbluDevice creates a device from a JSON byte array
func ParseMeshbluDevice(data []byte) (*MeshbluDevice, error) {
	device := &MeshbluDevice{}
	err := json.Unmarshal(data, device)
	return device, err
}

// NewUpdateErrorsBody returns the json body for updating the device
func NewUpdateErrorsBody(errors []string) (io.Reader, error) {
	updateDeviceErrors := &UpdateDeviceErrors{
		Errors:         errors,
		UpdateErrorsAt: time.Now().UnixNano() / int64(time.Millisecond),
	}
	data, err := json.Marshal(updateDeviceErrors)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}
