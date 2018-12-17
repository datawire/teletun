package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func hton16(x uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, x)
	return *(*uint16)(unsafe.Pointer(&buf[0]))
}

// This wrapper around AF_PACKET is probably horribly slow.  Use the
// PACKET_RX_RING sockopt to speed it up.
//
// https://www.kernel.org/doc/Documentation/networking/packet_mmap.txt
// https://codemonkeytips.blogspot.com/2011/07/asynchronous-packet-socket-reading-with.html

type afpacket struct {
	sock *os.File
	buf  []byte
}

func NewAFPacket(ethertype uint16) (L3Device, error) {
	fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_DGRAM, int(hton16(ethertype)))
	if err != nil {
		return nil, err
	}
	ret := &afpacket{
		sock: os.NewFile(uintptr(fd), fmt.Sprintf("AF_PACKET:%04x:any", ethertype)),
		buf:  make([]byte, unix.Getpagesize()),
	}
	return ret, nil
}

func (afp *afpacket) Close() error {
	return afp.sock.Close()
}

func (afp *afpacket) SendPacket(l3packet []byte) error {
	_, err := afp.sock.Write(l3packet)
	return err
}

func (afp *afpacket) RecvPacket() ([]byte, error) {
	n, err := afp.sock.Read(afp.buf)
	return afp.buf[:n], err
}

func (afp *afpacket) String() string {
	return afp.sock.Name()
}
