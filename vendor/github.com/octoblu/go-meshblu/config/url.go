package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// URL is a valid meshblu URL
type URL struct {
	hostName string
	port     int
	uri      *url.URL
}

// NewURL parses a url
func NewURL(hostName string, port int) (*URL, error) {
	schema := "http"
	if port == 443 {
		schema = "https"
	}
	urlStr := fmt.Sprintf("%s://%s:%v", schema, hostName, port)
	uri, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return &URL{hostName, port, uri}, nil
}

// ParseURL parses a url
func ParseURL(urlStr string) (*URL, error) {
	uri, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	hostName, port, err := getHostNameAndPort(uri)
	if err != nil {
		return nil, err
	}

	return &URL{hostName, port, uri}, nil
}

// HostName returns the hostname of the url
func (uri *URL) HostName() string {
	return uri.hostName
}

// Port returns the port of the url
func (uri *URL) Port() int {
	return uri.port
}

// SetPath sets the path of the uri
func (uri *URL) SetPath(path string) {
	uri.uri.Path = path
}

func (uri *URL) String() string {
	return uri.uri.String()
}

func getHostNameAndPort(uri *url.URL) (string, int, error) {
	parts := strings.Split(uri.Host, ":")

	if len(parts) > 1 {
		port, err := strconv.Atoi(parts[1])
		return parts[0], port, err
	}

	if uri.Scheme == "http" {
		return parts[0], 80, nil
	}

	if uri.Scheme == "https" {
		return parts[0], 443, nil
	}

	return "", 0, fmt.Errorf("URI contains invalid protocol: %v", uri.String())
}
