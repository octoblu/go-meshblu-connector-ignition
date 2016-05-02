package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// Runner defines the interface to run a Cmd
type Runner interface {
	Execute() error
	Shutdown() error
}

// Client defines the stucture of the client
type Client struct {
	serviceConfig *ServiceConfig
	cmd           *exec.Cmd
}

// New creates a new instance of runner
func New(serviceConfig *ServiceConfig) Runner {
	return &Client{serviceConfig, nil}
}

// Execute runs the connector
func (client *Client) Execute() error {
	commandPath := client.getCommandPath()
	if client.serviceConfig.Legacy {
		err := client.npmInstall()
		if err != nil {
			return err
		}
		commandPath = client.getLegacyCommandPath()
	}
	client.cmd = exec.Command(commandPath, client.serviceConfig.Args...)
	client.cmd.Env = client.getEnv()
	client.cmd.Stdout = os.Stdout
	client.cmd.Stderr = os.Stderr
	err := client.cmd.Start()
	if err != nil {
		return err
	}
	return client.cmd.Wait()
}

// Shutdown will send a SIGTERM to the connector process
func (client *Client) Shutdown() error {
	return client.cmd.Process.Signal(syscall.SIGTERM)
}

func (client *Client) getCommandPath() string {
	return path.Join("node_modules", "meshblu-connector-runner", "command.js")
}

func (client *Client) getLegacyCommandPath() string {
	return path.Join("node_modules", client.getFullConnectorName(), "command.js")
}

func (client *Client) npmInstall() error {
	cmd := exec.Command("npm", "install", client.getFullConnectorName())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func (client *Client) getFullConnectorName() string {
	return fmt.Sprintf("meshblu-%s", client.serviceConfig.ConnectorName)
}

func (client *Client) getEnv() []string {
	debug := os.Getenv("DEBUG")
	if debug == "" {
		if client.serviceConfig.Legacy {
			debug = fmt.Sprintf("DEBUG=%s", client.getFullConnectorName())
		} else {
			debug = fmt.Sprintf("DEBUG=%s", "meshblu-connector-*")
		}
	}
	env := append(os.Environ(), debug)
	return append(env, client.serviceConfig.Env...)
}
