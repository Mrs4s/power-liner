// +build !windows

package power_liner

import (
	"fmt"
	"syscall"
	"unsafe"
)

func GetTermSize() (int, int, error) {
	fd, _ := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	var dimensions [4]uint16
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0)
	if err != 0 {
		return 0, 0, err
	}
	return int(dimensions[1]), int(dimensions[0]), nil
}

func ClearScreen() error {
	fmt.Print("\033[2J")
	return nil
}
