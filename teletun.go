// TeleTUN is a L3 VPN for Kubernetes
//
// Only supports routing for IPv4 and IPv6.  Transport of other L3
// protocols (like X.25) should work, but routing for them will need
// to be manually set up.
package main

import (
	"log"
	"os/exec"
)

func main() {
	tun, err := OpenTun()
	if err != nil {
		log.Fatal("Unable to open TUN interface:", err)
	}

	err := exec.Command("ifconfig", IPtun.Name(),
		"address", kubectl.getaddr4(),
		"add", kubectl.getaddr6()).Run()
	if err != nil {
		log.Fatalf("Unable to set IP address of %s interface: %v", name, err)
	}
}
