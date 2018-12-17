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

	iface4, err := NewAFPacket(unix.ETH_P_IP)
	if err != nil {
		log.Fatalln("socket(AF_PACKET, …, IPv4):", err)
	}
	iface6, err := NewAFPacket(unix.ETH_P_IPV6)
	if err != nil {
		log.Fatalln("socket(AF_PACKET, …, IPv6):", iface6)
	}

	iface4packet, iface4done := DevPoller(iface4)
	iface6packet, iface6done := DevPoller(iface6)
	for {
		select {
		case l3packet := <-iface4packet:
			ipVersion := l3packet[0] >> 4
			src := net.IP(l3packet[12:16])
			dst := net.IP(l3packet[16:20])
			log.Println("l3packet", 4, ipVersion, src, dst)

			iface4done <- struct{}{}
		case l3packet := <-iface6packet:
			ipVersion := l3packet[0] >> 4
			src := net.IP(l3packet[8:24])
			dst := net.IP(l3packet[24:40])
			log.Println("l3packet", 6, ipVersion, src, dst)

			iface6done <- struct{}{}
		}
	}
}
