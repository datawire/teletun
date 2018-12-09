// +build linux

package main

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

/*
#include <net/route.h>
#include <sys/socket.h>
#include <netinet/in.h>

struct sockaddr ipv4_to_sockaddr(in_addr_t ip) {
	union { struct sockaddr sa; struct sockaddr_in sin; } addr;
	addr.sin.sin_family = AF_INET;
	addr.sin.sin_addr.s_addr = (in_addr_t)ip;
	return addr.sa;
}
*/
import "C"

type sioRoutingTable struct {
	fd4    int
	fd4_ok bool

	fd6    int
	fd6_ok bool
}

func (rt *sioRoutingTable) getFD4() (int, error) {
	var err error
	if !rt.fd4_ok {
		rt.fd4, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_IP)
		if err != nil {
			return -1, err
		}
		rt.fd4_ok = true
	}
	return rt.fd4, nil
}

type IPv4 [4]byte

// IPv4Route mimics <linux/route.h>:struct rtentry.
type IPv4Route struct {
	Dst        IPv4 // target address
	Gateway    IPv4 // gateway addr (RTF_GATEWAY)
	Genmask    IPv4 // target network mask (IP)
	Flags      uint16
	Metric     int16  // +1 for binary compatibility!
	MTU        uint   // per-route MTU/window
	Window     uint   // window clamping
	InitialRTT uint16 // Initial RTT
}

func ipv4_to_sockaddr(ip IPv4) C.struct_sockaddr {
	return C.ipv4_to_sockaddr(C.in_addr_t(ip[3])<<24 + C.in_addr_t(ip[2])<<16 + C.in_addr_t(ip[1])<<8 + C.in_addr_t(ip[0]))
}

func (rt *sioRoutingTable) AddRoute(route IPv4Route) error {
	fd, err := rt.getFD4()
	if err != nil {
		return err
	}

	var rawRoute C.struct_rtentry
	rawRoute.rt_dst = ipv4_to_sockaddr(route.Dst)
	rawRoute.rt_gateway = ipv4_to_sockaddr(route.Gateway)
	rawRoute.rt_genmask = ipv4_to_sockaddr(route.Genmask)
	rawRoute.rt_flags = C.ushort(route.Flags)
	rawRoute.rt_metric = C.short(route.Metric)
	rawRoute.rt_mtu = C.ulong(route.MTU)
	rawRoute.rt_window = C.ulong(route.Window)
	rawRoute.rt_irtt = C.ushort(route.InitialRTT)

	return unix.IoctlSetInt(fd, unix.SIOCADDRT, int(uintptr(unsafe.Pointer(&rawRoute))))
}
