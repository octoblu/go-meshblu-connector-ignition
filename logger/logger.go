package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Streams defines the streams supported by the logger
type Streams struct {
	memory *bytes.Buffer
	file   *os.File
}

// Client defines the logger struct
type Client struct {
	streams       *Streams
	isErrorStream bool
	interactive   bool
}

// Logger defines the interface for logging to mult-streams
type Logger interface {
	Stream() io.Writer
	Clear() error
	Get() []byte
	Close() error
}

// NewLogger creates an instance of a logger
func NewLogger(filePath string, interactive bool, isErrorStream bool) (Logger, error) {
	if filePath == "" {
		return nil, fmt.Errorf("Missing Log File Path %v", filePath)
	}
	streams := &Streams{}
	file, err := fileStream(filePath)
	if err != nil {
		return nil, err
	}
	streams.file = file
	streams.memory = memoryStream()
	return &Client{
		streams:       streams,
		isErrorStream: isErrorStream,
		interactive:   interactive,
	}, nil
}

// Stream returns a Writer stream to multiple internal streams
func (client *Client) Stream() io.Writer {
	file := client.streams.file
	memory := client.streams.memory
	if client.interactive {
		if client.isErrorStream {
			return io.MultiWriter(file, memory, os.Stderr)
		}
		return io.MultiWriter(file, memory, os.Stdout)
	}
	return io.MultiWriter(file, memory)
}

// Clear the streams
func (client *Client) Clear() error {
	client.streams.memory.Truncate(0)
	return client.streams.file.Truncate(0)
}

// Get the in-memory stream
func (client *Client) Get() []byte {
	return client.streams.memory.Bytes()
}

// Close the streams
func (client *Client) Close() error {
	err := client.Clear()
	if err != nil {
		return err
	}
	return client.streams.file.Close()
}

func fileStream(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0777)
}

func memoryStream() *bytes.Buffer {
	return &bytes.Buffer{}
}
