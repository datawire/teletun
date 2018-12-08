// +build linux

package main

import (
	"io"
	"strings"
	"syscall"
)

func IoctlInterfaceSetFlags(fd int, name string, flags int16) (string, error) {
	var req struct {
		name  [syscall.IFNAMESIZE]byte
		flags int16
	}

	if len(name) > syscall.IFNAMSIZE {
		panic("It is invalid to call IoctlInterfaceSetFlags() with a name > 16 bytes")
	}
	for i, b := range []byte(name) {
		req.name[i] = b
	}

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, syscall.TUNSETIFF, uintptr(action), uintptr(unsafe.Pointer(&req)))
	return strings.TrimRight(string(req.name), "\x00"), err
}

func OpenTun() (io.ReadWriteCloser, error) {
	// https://www.kernel.org/doc/Documentation/networking/tuntap.txt

	fd, err := syscall.Open("/dev/net/tun", syscall.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	name, err := IoctlInterfaceSetFlags(fd.Fd(), "tel%d", syscall.IFF_TUN|syscall.IFF_NO_PI)
	if err != nil {
		_ = syscall.Close(fd)
		return nil, err
	}

	wrapper := os.Newfile(fd, name)
	return wrapper, nil
}
