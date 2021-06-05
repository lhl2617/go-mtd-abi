package mtdabi

import (
	"golang.org/x/sys/unix"
)

// ioctl performs an ioctl operation specified by req and sets & gets the value
// on the device pointed by fd.
func ioctl(fd, req, value uintptr) error {
	_, _, err := unix.Syscall(unix.SYS_IOCTL, fd, req, value)
	if err != 0 {
		return err
	}
	return nil
}
