package logger

import (
	"os"

	"github.com/mattn/go-isatty"
)

// IsTerminal returns a bool if it is stderr
func IsTerminal() bool {
	return isatty.IsTerminal(os.Stderr.Fd())
}
