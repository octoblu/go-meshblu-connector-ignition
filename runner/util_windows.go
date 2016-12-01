package runner

import "syscall"

func sysProcAttrForOS() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}
