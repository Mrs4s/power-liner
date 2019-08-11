// +build windows

package power_liner

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
	procFillConsoleOutputAttribute = kernel32.NewProc("FillConsoleOutputAttribute")
)

type (
	coord struct {
		x int16
		y int16
	}
	smallRect struct {
		left   int16
		top    int16
		right  int16
		bottom int16
	}
	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        int16
		window            smallRect
		maximumWindowSize coord
	}
)

func GetTermSize() (width, height int, err error) {
	fd, _ := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	var info consoleScreenBufferInfo
	_, _, e := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return 0, 0, error(e)
	}
	return int(info.size.x), int(info.size.y), nil
}

func ClearScreen() error {
	fd, _ := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	var info consoleScreenBufferInfo
	_, _, e := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return error(e)
	}
	var (
		cursor coord
		w      int16
		total  = info.size.x * info.size.y
		handle = syscall.Handle(os.Stdout.Fd())
	)
	_, _, _ = procFillConsoleOutputCharacter.Call(
		uintptr(handle),
		uintptr(' '),
		uintptr(total),
		*(*uintptr)(unsafe.Pointer(&cursor)),
		uintptr(unsafe.Pointer(&w)),
	)
	_, _, _ = procFillConsoleOutputAttribute.Call(
		uintptr(handle),
		uintptr(info.attributes),
		uintptr(total), *(*uintptr)(unsafe.Pointer(&cursor)),
		uintptr(unsafe.Pointer(&w)),
	)
	return nil
}
