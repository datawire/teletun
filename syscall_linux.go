// +build linux

package main

import (
	"bytes"
	"unsafe"

	"golang.org/x/sys/unix"
)

// IoctlTunSetInterfaceFlags wraps the TUNSETIFF ioctl
func IoctlTunSetInterfaceFlags(fd int, name string, flags int16) (string, error) {
	var ifreq struct {
		name  [unix.IFNAMSIZ]byte
		flags int16
	}

	if len(name) > unix.IFNAMSIZ {
		return "", unix.EINVAL
	}
	for i, b := range []byte(name) {
		ifreq.name[i] = b
	}

	ifreq.flags = flags

	// <linux/if.h> declares TUNSETIFF as taking an 'int', not a
	// pointer, so I guess casting the pointer to an int and
	// calling IoctlSetInt is the "right thing".
	err := unix.IoctlSetInt(fd, unix.TUNSETIFF, int(uintptr(unsafe.Pointer(&ifreq))))

	return string(bytes.SplitN(ifreq.name[:], []byte{0}, 2)[0]), err
}
