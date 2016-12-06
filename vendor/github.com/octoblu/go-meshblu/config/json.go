package config

import "encoding/json"

// JSON is for serializing/deserializing JSON data
type JSON struct {
	UUID  string `json:"uuid"`
	Token string `json:"token"`

	// Deprecated values, the SRV configuration should be used instead where possible
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`

	// SRV configuration
	ResolveSRV bool   `json:"resolveSrv"`
	Secure     bool   `json:"secure"`
	Domain     string `json:"domain"`
}

// Parse parses a byte array into a JSON object
func Parse(data []byte) (*JSON, error) {
	parsed := &JSON{}
	err := json.Unmarshal(data, parsed)

	if err != nil {
		return nil, err
	}

	if parsed.ResolveSRV {
		parsed.Secure, err = parseSecure(data)
		if err != nil {
			return nil, err
		}
	}

	return parsed, err
}

func parseSecure(data []byte) (bool, error) {
	var parsed interface{}
	err := json.Unmarshal(data, &parsed)
	if err != nil {
		return false, err
	}

	obj := parsed.(map[string]interface{})
	parsedSecure, ok := obj["secure"]
	if !ok {
		return true, nil
	}

	secure, ok := parsedSecure.(bool)
	if !ok {
		return true, nil
	}

	return secure, nil
}
