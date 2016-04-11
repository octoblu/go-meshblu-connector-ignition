package runner

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Runner defines the interface to run a Cmd
type Runner interface {
	Execute() error
	Shutdown() error
}

// Client defines the stucture of the client
type Client struct {
	legacy    bool
	connector string
	cmd       *exec.Cmd
}

// New creates a new instance of runner
func New(legacy bool, connector string) Runner {
	return &Client{legacy, connector, nil}
}

// Execute runs the connector
func (client *Client) Execute() error {
	if client.legacy {
		err := client.npmInstall()
		if err != nil {
			return err
		}
		client.setupLegacyCommand()
	} else {
		client.setupCommand()
	}
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

func (client *Client) setupCommand() {
	client.cmd = exec.Command("node", "./node_modules/.bin/meshblu-connector-runner", ".")
	client.cmd.Env = getEnv("meshblu-connector-*")
}

func (client *Client) setupLegacyCommand() {
	client.cmd = exec.Command("node", fmt.Sprintf("./node_modules/%s/command.js", client.getFullConnectorName()))
	client.cmd.Env = getEnv(client.getFullConnectorName())
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
	return fmt.Sprintf("meshblu-%s", client.connector)
}

func getEnv(debugVar string) []string {
	env := os.Environ()
	if os.Getenv("DEBUG") == "" {
		env = append(env, fmt.Sprintf("DEBUG=%s", debugVar))
	}
	return env
}
