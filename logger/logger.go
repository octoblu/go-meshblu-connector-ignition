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
	memoryStream buffer.Buffer
	multi        io.Writer
	fileStream   buffer.Buffer
	file         *os.File
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
	fileStream := buffer.NewFile(100*1024*1024, file)
	memoryStream := buffer.NewUnboundedBuffer(32*1024, 100*1024*1024)
	both := buffer.NewMulti(fileStream, memoryStream)
	var multi io.Writer
	if IsTerminal() {
		if isErrorStream {
			multi = io.MultiWriter(both, os.Stderr)
		} else {
			multi = io.MultiWriter(both, os.Stdout)
		}
	} else {
		multi = both
	}
	return &Client{
		memoryStream: memoryStream,
		file:         file,
		fileStream:   fileStream,
		multi:        multi,
	}, nil
}

// Stream returns a Writer stream to multiple internal streams
func (client *Client) Stream() io.Writer {
	return client.multi
}

// Clear the streams
func (client *Client) Clear() error {
	client.memoryStream.Reset()
	client.fileStream.Reset()
	return nil
}

// Get the in-memory stream
func (client *Client) Get() []byte {
	return bufio.NewScanner(client.memoryStream).Bytes()
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
