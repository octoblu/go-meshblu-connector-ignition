package logger

type fakeMainLoggerClient struct {
}

// FakeMainLogger is a fake interface of the main logger
type FakeMainLogger interface {
	Clear() error
	Close() error
	Info(key, msg string)
	Error(key, msg string, err error)
}

// NewFakeMainLogger creates a fake instance of the logger
func NewFakeMainLogger() FakeMainLogger {
	return &fakeMainLoggerClient{}
}

// Info log a message
func (client *fakeMainLoggerClient) Info(key, msg string) {
}

// Error log a message
func (client *fakeMainLoggerClient) Error(key, msg string, err error) {
}

// Clear the stream
func (client *fakeMainLoggerClient) Clear() error {
	return nil
}

// Close the stream
func (client *fakeMainLoggerClient) Close() error {
	return nil
}
