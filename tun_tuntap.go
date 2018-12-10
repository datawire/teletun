// tun_linux.go: Open a Tunnel (L3 virtual interface) using the Universal TUN/TAP driver.

// +build linux

package main

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func OpenTun() (io.ReadWriteCloser, error) {
	// https://www.kernel.org/doc/Documentation/networking/tuntap.txt

	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	name, err := IoctlTunSetInterfaceFlags(fd, "tel%d", unix.IFF_TUN|unix.IFF_NO_PI)
	if err != nil {
		_ = unix.Close(fd)
		return nil, err
	}

	wrapper := os.NewFile(uintptr(fd), name)
	return wrapper, nil
}
