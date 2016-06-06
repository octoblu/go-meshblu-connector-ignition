package meshblu

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/octoblu/go-meshblu/config"
)

// Meshblu interfaces with a remote meshblu server
type Meshblu interface {
	SetAuth(uuid, token string)
	GetDevice(uuid string) ([]byte, error)
	UpdateDevice(uuid string, body io.Reader) ([]byte, error)
}

// Client interfaces with a remote meshblu server
type Client struct {
	uri, uuid, token string
}

// Dial constructs a new Meshblu instance and creates a connection
func Dial(uri string) (Meshblu, error) {
	return &Client{
		uri: uri,
	}, nil
}

// SetAuth sets the authentication
func (client *Client) SetAuth(uuid, token string) {
	client.uuid = uuid
	client.token = token
}

// GetDevice returns a byte response of the meshblu device
func (client *Client) GetDevice(uuid string) ([]byte, error) {
	return client.request("GET", fmt.Sprintf("/v2/devices/%s", uuid), nil)
}

// UpdateDevice returns a byte response of the meshblu device
func (client *Client) UpdateDevice(uuid string, body io.Reader) ([]byte, error) {
	return client.request("PATCH", fmt.Sprintf("/v2/devices/%s", uuid), body)
}

func (client *Client) request(method, path string, body io.Reader) ([]byte, error) {
	meshbluURL, err := config.ParseURL(client.uri)
	if err != nil {
		return nil, err
	}
	meshbluURL.SetPath(path)

	httpClient := &http.Client{}
	request, err := http.NewRequest(method, meshbluURL.String(), body)
	request.SetBasicAuth(client.uuid, client.token)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Connection", "Keep-Alive")
	request.Header.Add("Keep-Alive", "timeout=30,max=15")
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode > 299 {
		return nil, fmt.Errorf("Meshblu returned invalid response code: %v", response.StatusCode)
	}

	return ioutil.ReadAll(response.Body)
}
