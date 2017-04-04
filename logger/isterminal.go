package logger

import (
	"os"

	isatty "github.com/mattn/go-isatty"
)

// IsTerminal returns a bool if it is stderr
func IsTerminal() bool {
	return isatty.IsTerminal(os.Stderr.Fd())
}
