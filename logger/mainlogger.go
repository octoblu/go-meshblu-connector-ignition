package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/osext"
)

var mainLogger MainLogger

// MainClient defines the mainlogger struct
type MainClient struct {
	file           *os.File
	fileStream     io.Writer
	stderr         io.Writer
	currentVersion string
}

// MainLogger defines the interface for logging to stderr or stdout
type MainLogger interface {
	Clear() error
	Close() error
	Info(key, msg string)
	Error(key, msg string, err error)
}

// GetMainLogger gets the global instance of the MainLogger
func GetMainLogger() MainLogger {
	return mainLogger
}

// InitMainLogger creates a global instance of the main logger
func InitMainLogger(currentVersion string) error {
	filePath, err := getMainLogFilePath()
	if err != nil {
		return err
	}
	file, err := getFileFromPath(filePath)
	if err != nil {
		return err
	}
	fileStream := getFileStream(file)
	stderr := getStderrStream()
	mainLogger = &MainClient{
		file:           file,
		fileStream:     fileStream,
		stderr:         stderr,
		currentVersion: currentVersion,
	}
	return nil
}

// Info log a message
func (client *MainClient) Info(key, msg string) {
	timestamp := time.Now()
	logMessage := fmt.Sprintf("( %s )[info][v%s][%s] %s", timestamp, client.currentVersion, key, msg)
	prettyMessage := fmt.Sprintf("%s[info]%s[%s][v%s][%s%s%s] %s", cyan, reset, timestamp.Format("15:04:05.000"), client.currentVersion, magenta, key, reset, msg)
	fmt.Fprintln(client.fileStream, logMessage)
	if IsTerminal() {
		fmt.Fprintln(client.stderr, prettyMessage)
	}
}

// Error log a message
func (client *MainClient) Error(key, msg string, err error) {
	timestamp := time.Now()
	logMessage := fmt.Sprintf("( %s )[error][v%s][%s] %s %s", timestamp, client.currentVersion, key, msg, err.Error())
	prettyMessage := fmt.Sprintf("%s[error]%s[%s][v%s][%s%s%s] %s %s", red, reset, timestamp.Format("15:04:05.000"), client.currentVersion, cyan, key, reset, msg, err.Error())
	fmt.Fprintln(client.fileStream, logMessage)
	if IsTerminal() {
		fmt.Fprintln(client.stderr, prettyMessage)
	}
}

// Clear the stream
func (client *MainClient) Clear() error {
	return client.file.Truncate(0)
}

// Close the stream
func (client *MainClient) Close() error {
	err := client.Clear()
	if err != nil {
		return err
	}
	return client.file.Close()
}

// get the main log file path
func getMainLogFilePath() (string, error) {
	fullPath, err := osext.Executable()
	if err != nil {
		return "", err
	}
	dir, _ := filepath.Split(fullPath)
	return filepath.Join(dir, "log", "ignition.log"), nil
}

func getFileStream(file *os.File) io.Writer {
	return io.Writer(file)
}

func getStderrStream() io.Writer {
	return os.Stderr
}
