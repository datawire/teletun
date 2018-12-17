// TeleTUN is a L3 VPN for Kubernetes
package main

import (
	"log"
	"os"
)

func main() {
	tun, err := OpenTun()
	if err != nil {
		log.Fatal("Unable to open TUN interface:", err)
	}
	log.Print(tun.(*os.File).Name())

	select {}

	// 1. Get IP address from kubectl/remote
	// 2. Get list of IPs from kubectl (and subscribe to updates)
	// 3. Get resolv.conf from remote
	// 4. Pass traffic over DTLS to
}
