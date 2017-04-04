package logger

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/djherbis/buffer"
)

// Client defines the logger struct
type Client struct {
	stream buffer.Buffer
	multi  io.Writer
	file   *os.File
}

// Logger defines the interface for logging to mult-streams
type Logger interface {
	Stream() io.Writer
	Clear() error
	Get() []byte
	Close() error
}

// NewLogger creates an instance of a logger
func NewLogger(filePath string, isErrorStream bool) (Logger, error) {
	if filePath == "" {
		return nil, fmt.Errorf("Missing Log File Path %v", filePath)
	}
	file, err := getFileFromPath(filePath)
	if err != nil {
		return nil, err
	}
	stream := buffer.NewUnboundedBuffer(32*1024, 1024*1024)
	var multi io.Writer
	if IsTerminal() {
		if isErrorStream {
			multi = buffer.NewMulti(stream, buffer.NewFile(1024*1024, os.Stderr))
		} else {
			multi = buffer.NewMulti(stream, buffer.NewFile(1024*1024, os.Stdout))
		}
	} else {
		multi = stream
	}
	return &Client{
		file:   file,
		stream: stream,
		multi:  multi,
	}, nil
}

// Stream returns a Writer stream to multiple internal streams
func (client *Client) Stream() io.Writer {
	return client.multi
}

// Clear the streams
func (client *Client) Clear() error {
	client.stream.Reset()
	return nil
}

// Get the in-memory stream
func (client *Client) Get() []byte {
	return bufio.NewScanner(client.stream).Bytes()
}

// Close the streams
func (client *Client) Close() error {
	err := client.Clear()
	if err != nil {
		return err
	}
	return client.file.Close()
}

func getFileFromPath(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
}
