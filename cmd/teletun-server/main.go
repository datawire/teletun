// TeleTUN is a L3 VPN for Kubernetes
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// This is probably horribly slow.
// https://www.kernel.org/doc/Documentation/networking/packet_mmap.txt

func errUsage(a ...interface{}) {
	fmt.Fprintln(flag.CommandLine.Output(), a...)
	flag.Usage()
	os.Exit(2)
}

func hton16(x uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, x)
	return *(*uint16)(unsafe.Pointer(&buf[0]))
}

func main() {
	argPort := flag.String("port", "", "UDP port to listen on")
	flag.Parse()
	if flag.NArg() > 0 {
		errUsage("too many arguments:", flag.Args())
	}
	if *argPort == "" {
		errUsage("must specify a --port")
	}
	port, err := net.LookupPort("udp", *argPort)
	if err != nil {
		errUsage("invalid --port:", err)
	}

	func() {
		log.Printf("listening on UDP port %q (%d)...", *argPort, port)
		conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatal(err)
		}
		log.Println("conn", conn)
	}()

	func() {
		fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_DGRAM, int(hton16(unix.ETH_P_IP)))
		if err != nil {
			log.Fatalln("a", err)
		}
		f := os.NewFile(uintptr(fd), "any")
		defer f.Close()
		buf := make([]byte, unix.SizeofSockaddrLinklayer+1500)
		for {
			n, err := f.Read(buf)
			l3packet := buf[:n]
			ipVersion := l3packet[0] >> 4
			var src, dst net.IP
			switch ipVersion {
			case 4:
				src = net.IP(l3packet[12:16])
				dst = net.IP(l3packet[16:20])
			case 6:
				src = net.IP(l3packet[8:24])
				dst = net.IP(l3packet[24:40])
			default:
				log.Println("unknown packet", l3packet)
			}
			log.Println("l3packet", ipVersion, src, dst)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}()
}
