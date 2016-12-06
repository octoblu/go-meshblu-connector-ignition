package config

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// Config interfaces with a remote meshblu server
type Config struct {
	uuid, token        string
	protocol, hostname string
	port               int

	resolveSrv, secure bool
	domain             string
}

// ReadFromConfig Reads the Config from a file
func ReadFromConfig(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	jsonConfig, err := Parse(file)
	if err != nil {
		return nil, err
	}

	if jsonConfig.ResolveSRV {
		return newWithResolveSRV(jsonConfig)
	}
	return newWithURL(jsonConfig)
}

// Hostname returns the config's hostname
func (config *Config) Hostname() string {
	return config.hostname
}

// Port returns the config's port
func (config *Config) Port() int {
	return config.port
}

// Protocol returns the config's protocol
func (config *Config) Protocol() string {
	return config.protocol
}

// ResolveSRV returns true if resolve SRV is true
func (config *Config) ResolveSRV() bool {
	return config.resolveSrv
}

// Secure returns true if the config's secure was true or missing
func (config *Config) Secure() bool {
	return config.secure
}

// Domain returns the config's domain
func (config *Config) Domain() string {
	return config.domain
}

// Token returns the config's token
func (config *Config) Token() string {
	return config.token
}

// UUID returns the config's uuid
func (config *Config) UUID() string {
	return config.uuid
}

// ToURL serializes the object to the meshblu.json format
func (config *Config) ToURL() (string, error) {
	meshbluURI, err := NewURL(config.hostname, config.port)
	if err != nil {
		return "", err
	}

	return meshbluURI.String(), nil
}

// assertNoSRVProperties verifies that no domain or service
// values are present
func assertNoSRVProperties(parsed *JSON) error {
	var errorMessages []string

	if parsed.Domain != "" {
		errorMessages = append(errorMessages, "domain property only allowed when 'resolveSrv' is 'true'")
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf(strings.Join(errorMessages, ", "))
	}

	return nil
}

// assertNoURLProperties verifies that no protocol, hostname, or port
// values are present
func assertNoURLProperties(parsed *JSON) error {
	var errorMessages []string

	if parsed.Protocol != "" {
		errorMessages = append(errorMessages, "protocol property only allowed when 'resolveSrv' is 'false'")
	}
	if parsed.Hostname != "" {
		errorMessages = append(errorMessages, "hostname property only allowed when 'resolveSrv' is 'false'")
	}
	if parsed.Port != 0 {
		errorMessages = append(errorMessages, "port property only allowed when 'resolveSrv' is 'false'")
	}
	if len(errorMessages) > 0 {
		return fmt.Errorf(strings.Join(errorMessages, ", "))
	}

	return nil
}

func newWithResolveSRV(parsed *JSON) (*Config, error) {
	err := assertNoURLProperties(parsed)
	if err != nil {
		return nil, err
	}

	config := &Config{
		uuid:       parsed.UUID,
		token:      parsed.Token,
		resolveSrv: true,
		domain:     parsed.Domain,
		secure:     parsed.Secure,
	}

	if config.domain == "" {
		config.domain = "octoblu.com"
	}

	return config, nil
}

func newWithURL(parsed *JSON) (*Config, error) {
	err := assertNoSRVProperties(parsed)
	if err != nil {
		return nil, err
	}

	config := &Config{
		uuid:       parsed.UUID,
		token:      parsed.Token,
		resolveSrv: false,
		protocol:   parsed.Protocol,
		hostname:   parsed.Hostname,
		port:       parsed.Port,
	}

	if config.port == 0 {
		config.port = 443
	}

	if config.protocol == "" {
		if config.port == 443 {
			config.protocol = "https"
		} else {
			config.protocol = "http"
		}
	}

	if config.hostname == "" {
		config.hostname = "meshblu.octoblu.com"
	}

	return config, nil
}
