package main

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

type packetConnDev struct {
	conn net.PacketConn
	dst  net.Addr
	buf  []byte
}

func NewUDP4Device(port int) (L3Device, error) {
	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	ret := &packetConnDev{
		conn: conn,
		buf:  make([]byte, unix.Getpagesize()),
	}
	return ret, nil
}

func (c *packetConnDev) Close() error {
	return c.conn.Close()
}

func (c *packetConnDev) SendPacket(l3packet []byte) error {
	if c.dst == nil {
		return nil
	}
	_, err := c.conn.WriteTo(l3packet, c.dst)
	return err
}

func (c *packetConnDev) RecvPacket() ([]byte, error) {
	n, addr, err := c.conn.ReadFrom(c.buf)
	c.dst = addr
	return c.buf[:n], err
}

func (c *packetConnDev) String() string {
	return fmt.Sprintf("local:%v→remote:%v", c.conn.LocalAddr(), c.dst)
}
