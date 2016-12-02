package meshblu

import "fmt"

// RecoverableError is an error returned by meshblu
// methods that probably go away over time. Example errors
// include:
// * Meshblu is returing 5xx errors
// * the internet is down
type RecoverableError struct {
	// The original error, just in case
	OriginalError error
}

// IsRecoverable returns true if the error is an instance of
// RecoverableError
func IsRecoverable(err error) bool {
	_, isRecoverableError := err.(*RecoverableError)
	return isRecoverableError
}

// NewRecoverableError wraps the error so that it's easy
// to see if might be able to get away with just trying
// again later
func NewRecoverableError(err error) *RecoverableError {
	return &RecoverableError{OriginalError: err}
}

// Error represents the string representation of the original error
// prefixed by "RecoverableError: "
func (err *RecoverableError) Error() string {
	return fmt.Sprintln("RecoverableError:", err.OriginalError.Error())
}
