package meshblu

import (
	"fmt"
	"net"
	"strings"

	"github.com/octoblu/go-meshblu/config"
)

// NewClient creates a new device struct
func NewClient(configPath string) (Meshblu, string, error) {
	config, err := config.ReadFromConfig(configPath)
	if err != nil {
		return nil, "", err
	}

	if config.ResolveSRV() {
		return clientFromSRV(config)
	}
	return clientFromURL(config)
}

func clientFromSRV(cfg *config.Config) (Meshblu, string, error) {
	service := "meshblu"
	domain := cfg.Domain()
	protocol := "http"
	if cfg.Secure() {
		protocol = "https"
	}

	_, addrs, err := net.LookupSRV(service, protocol, domain)
	if err != nil {
		return nil, "", err
	}

	if len(addrs) == 0 {
		return nil, "", fmt.Errorf("Received an empty list of addresses attempting to resolve SRV")
	}

	hostname := strings.TrimSuffix(addrs[0].Target, ".")
	port := int(addrs[0].Port)
	url, err := config.NewURL(hostname, port)
	if err != nil {
		return nil, "", err
	}

	meshbluClient, err := Dial(url.String())
	if err != nil {
		return nil, "", err
	}

	meshbluClient.SetAuth(cfg.UUID(), cfg.Token())
	return meshbluClient, cfg.UUID(), nil
}

func clientFromURL(cfg *config.Config) (Meshblu, string, error) {
	url, err := cfg.ToURL()
	if err != nil {
		return nil, "", err
	}

	meshbluClient, err := Dial(url)
	if err != nil {
		return nil, "", err
	}

	meshbluClient.SetAuth(cfg.UUID(), cfg.Token())
	return meshbluClient, cfg.UUID(), nil
}
