// TeleTUN is a L3 VPN for Kubernetes
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"golang.org/x/sys/unix"
)

type L3Device interface {
	SendPacket([]byte) error
	RecvPacket() ([]byte, error)
	Close() error
}

func DevPoller(dev L3Device) (<-chan []byte, chan<- struct{}) {
	chanPacket := make(chan []byte)
	chanDone := make(chan struct{})
	go func() {
		for {
			packet, err := dev.RecvPacket()
			if err != nil {
				log.Fatalln(dev, err)
			}
			chanPacket <- packet
			<-chanDone
		}
	}()
	return chanPacket, chanDone
}

func errUsage(a ...interface{}) {
	fmt.Fprintln(flag.CommandLine.Output(), a...)
	flag.Usage()
	os.Exit(2)
}

func debugLogPacket(msg string, dev L3Device, l3packet []byte) {
	ipVersion := l3packet[0] >> 4
	var src, dst net.IP
	switch ipVersion {
	case 4:
		src = net.IP(l3packet[12:16])
		dst = net.IP(l3packet[16:20])
	case 6:
		src = net.IP(l3packet[8:24])
		dst = net.IP(l3packet[24:40])
	}

	log.Printf("%s: %v: IPv%d src=%v dst=%v\n", msg, dev, ipVersion, src, dst)
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

	udp, err := NewUDP4Device(port)
	if err != nil {
		log.Fatalf("ListenPacket(\"udp\", \":%d\"): %v", port, err)
	}
	iface4, err := NewAFPacket(unix.ETH_P_IP)
	if err != nil {
		log.Fatalln("socket(AF_PACKET, …, IPv4):", err)
	}
	iface6, err := NewAFPacket(unix.ETH_P_IPV6)
	if err != nil {
		log.Fatalln("socket(AF_PACKET, …, IPv6):", iface6)
	}

	udpPacket, udpDone := DevPoller(udp)
	iface4packet, iface4done := DevPoller(iface4)
	iface6packet, iface6done := DevPoller(iface6)
	for {
		select {
		case l3packet := <-udpPacket:
			debugLogPacket("handle", udp, l3packet)
			ipVersion := l3packet[0] >> 4
			switch ipVersion {
			case 4:
				err := iface4.SendPacket(l3packet)
				if err != nil {
					log.Fatalln("send", iface4, err)
				}
			case 6:
				err := iface6.SendPacket(l3packet)
				if err != nil {
					log.Fatalln("send", iface6, err)
				}
			default:
				log.Fatal("Unknown packet type")
			}
			udpDone <- struct{}{}
		case l3packet := <-iface4packet:
			skip := false

			// L3 bypass: Look at L4/L5 to decide if we
			// want to handle this packet at L3 or bypass
			// the L3 handling.  Currently does L4
			// filtering, rhs says L5 would be be useful,
			// but that's a lot harder to implement.
			ihl := l3packet[0] & 0x0f
			l4proto := l3packet[9]
			l4packet := l3packet[ihl*4:]
			if l4proto == unix.IPPROTO_UDP {
				dstPort := uint16(l4packet[2])<<8 | uint16(l4packet[3])
				if int(dstPort) == port {
					skip = true
				}
			}

			if skip {
				debugLogPacket("skip", iface4, l3packet)
			} else {
				debugLogPacket("handle", iface4, l3packet)
				err := udp.SendPacket(l3packet)
				if err != nil {
					log.Fatalln("send", udp, err)
				}
			}
			iface4done <- struct{}{}
		case l3packet := <-iface6packet:
			skip := false

			// Insert L3 bypass here

			if skip {
				debugLogPacket("skip", iface6, l3packet)
			} else {
				debugLogPacket("handle", iface6, l3packet)
				err := udp.SendPacket(l3packet)
				if err != nil {
					log.Fatalln("send", udp, err)
				}
			}
			iface6done <- struct{}{}
		}
	}
}
