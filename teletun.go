// TeleTUN is a L3 VPN for Kubernetes
package main

import (
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
