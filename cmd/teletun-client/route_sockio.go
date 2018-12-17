// route_sockio.go: Manipulate kernel routing tables with the SIOCADDRT/SIOCDELRT ioctl()s.

// +build linux

package main

import (
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

//#include <stdlib.h>     /* for free(3p) */
//#include <net/route.h>  /* for struct rtentry */
//
//#include <sys/socket.h> /* for struct sockaddr */
//#include <netinet/in.h> /* for struct sockaddr_in and in_addr_t */
//
//struct sockaddr ipv4_to_sockaddr(in_addr_t ip) {
//	union { struct sockaddr sa; struct sockaddr_in sin; } addr;
//	addr.sin.sin_family = AF_INET;
//	addr.sin.sin_addr.s_addr = (in_addr_t)ip;
//	return addr.sa;
//}
import "C"

type sockioRoutingTable struct {
	fd4    int
	fd4_ok bool

	fd6    int
	fd6_ok bool
}

var RoutingTable = sockioRoutingTable{}

func (rt *sockioRoutingTable) close() {
	if rt.fd4_ok {
		unix.Close(rt.fd4)
	}
	if rt.fd6_ok {
		unix.Close(rt.fd6)
	}
	runtime.SetFinalizer(rt, nil)
}

func (rt *sockioRoutingTable) getFD4() (int, error) {
	var err error
	if !rt.fd4_ok {
		rt.fd4, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_IP)
		if err != nil {
			return -1, err
		}
		rt.fd4_ok = true
		runtime.SetFinalizer(rt, (*sockioRoutingTable).close)
	}
	return rt.fd4, nil
}

func (rt *sockioRoutingTable) getFD6() (int, error) {
	var err error
	if !rt.fd6_ok {
		rt.fd6, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_IPV6)
		if err != nil {
			return -1, err
		}
		rt.fd6_ok = true
		runtime.SetFinalizer(rt, (*sockioRoutingTable).close)
	}
	return rt.fd6, nil
}

type IPv4 [4]byte

// IPv4Route mimics <linux/route.h>:struct rtentry.
type IPv4Route struct {
	Dst        IPv4 // target address
	Gateway    IPv4 // gateway addr (RTF_GATEWAY)
	Genmask    IPv4 // target network mask (IP)
	Flags      uint16
	Metric     int16  // +1 for binary compatibility!
	Dev        string // forcing the device to add
	MTU        uint   // per-route MTU/window
	Window     uint   // window clamping
	InitialRTT uint16 // Initial RTT
}

func ipv4_to_sockaddr(ip IPv4) C.struct_sockaddr {
	return C.ipv4_to_sockaddr(C.in_addr_t(ip[3])<<24 + C.in_addr_t(ip[2])<<16 + C.in_addr_t(ip[1])<<8 + C.in_addr_t(ip[0]))
}

func (rt *sockioRoutingTable) AddIPv4Route(route IPv4Route) error {
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
	rawRoute.rt_dev = C.CString(route.Dev)
	defer C.free(unsafe.Pointer(rawRoute.rt_dev))
	rawRoute.rt_mtu = C.ulong(route.MTU)
	rawRoute.rt_window = C.ulong(route.Window)
	rawRoute.rt_irtt = C.ushort(route.InitialRTT)

	return unix.IoctlSetInt(fd, unix.SIOCADDRT, int(uintptr(unsafe.Pointer(&rawRoute))))
}

func (rt *sockioRoutingTable) DelIPv4Route(route IPv4Route) error {
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

	return unix.IoctlSetInt(fd, unix.SIOCDELRT, int(uintptr(unsafe.Pointer(&rawRoute))))
}

type IPv6 [16]byte

// IPv6Route mimics <linux/ipv6_route.h>:struct in6_rtmsg.
type IPv6Route struct {
	Dst            IPv6
	Src            IPv6
	Gateway        IPv6
	Type           uint32
	DstLen         uint16
	SrcLen         uint16
	Metric         uint32
	info           int // C.ulong
	Flags          uint32
	InterfaceIndex int32 // C.int
}

func (rt *sockioRoutingTable) GetInterfaceIndex(ifname string) (int32, error) {
	fd, err := rt.getFD6()
	if err != nil {
		return -1, err
	}

	return IoctlGetInterfaceIndex(fd, ifname)
}

func (rt *sockioRoutingTable) AddIPv6Route(route IPv6Route) error {
	fd, err := rt.getFD6()
	if err != nil {
		return err
	}

	return unix.IoctlSetInt(fd, unix.SIOCADDRT, int(uintptr(unsafe.Pointer(&route))))
}

func (rt *sockioRoutingTable) DelIPv6Route(route IPv6Route) error {
	fd, err := rt.getFD6()
	if err != nil {
		return err
	}

	return unix.IoctlSetInt(fd, unix.SIOCDELRT, int(uintptr(unsafe.Pointer(&route))))
}
