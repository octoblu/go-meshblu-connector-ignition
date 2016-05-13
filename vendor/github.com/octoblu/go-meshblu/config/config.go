package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config interfaces with a remote meshblu server
type Config struct {
	UUID     string `json:"uuid"`
	Token    string `json:"token"`
	Server   string `json:"server"`
	HostName string `json:"hostname"`
	Port     int    `json:"port"`
}

// NewConfig constructs a new Meshblu instance
func NewConfig(UUID, Token, Server string, Port int) *Config {
	return &Config{UUID, Token, Server, "", Port}
}

// ReadFromConfig Reads the Config from a file
func ReadFromConfig(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseConfig(file)
}

// ParseConfig creates a config with the UUID and Token from a JSON byte array
func ParseConfig(data []byte) (*Config, error) {
	config := &Config{}
	err := json.Unmarshal(data, config)
	return config, err
}

// ToJSON serializes the object to the meshblu.json format
func (config *Config) ToJSON() ([]byte, error) {
	return json.Marshal(config)
}

// ToURL serializes the object to the meshblu.json format
func (config *Config) ToURL() (string, error) {
	hostName := config.HostName
	if hostName == "" {
		hostName = config.Server
	}
	meshbluURI, err := NewURL(hostName, config.Port)
	if err != nil {
		return "", err
	}
	return meshbluURI.String(), nil
}
