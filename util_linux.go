package beego

import (
	"os"
	"syscall"
)

func (this xgoUtil) SetDaemonMode(nochdir, noclose int) int {
	if syscall.Getppid() == 1 {
		return 0
	}
	ret, ret2, err := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		return -1
	}
	if ret2 < 0 {
		os.Exit(-1)
	}
	if ret > 0 {
		os.Exit(0)
	}
	syscall.Umask(0)
	s_ret, s_errno := syscall.Setsid()
	if s_ret < 0 {
		return -1
	}
	if nochdir == 0 {
		os.Chdir("/")
	}
	if noclose == 0 {
		f, e := os.OpenFile("/dev/null", os.O_RDWR, 0)
		if e == nil {
			fd := f.Fd()
			syscall.Dup2(int(fd), int(os.Stdin.Fd()))
			syscall.Dup2(int(fd), int(os.Stdout.Fd()))
			syscall.Dup2(int(fd), int(os.Stderr.Fd()))
		}
	}
	return 0
}
