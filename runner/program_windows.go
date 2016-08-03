package runner

import (
	"os/exec"
	"syscall"
)

func (prg *Program) initCmdForOS(cmd *exec.Cmd) {
	prg.cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
