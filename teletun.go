// TeleTUN is a L3 VPN for Kubernetes
//
// Only supports routing for IPv4 and IPv6.  Transport of other L3
// protocols (like X.25) should work, but routing for them will need
// to be manually set up.
package main

import (
	"log"
	"net"
)

// FIXME: This calls socat(1) because I could not find a good DTLS
// implementation in Go, and I didn't want to muck with OpenSSL
// bindings yet.  I don't like this because socat does not allow us to
// deal with errors robustly, has bad buffering characteristics, and
// might not actually be compiled with OpenSSL support.
func DialDTLS(address string) (net.PacketConn, error) {
	// socat STDIO OPENSSL:{address},method=DTLS1
}

func main() {
	tun, err := OpenTun()
	if err != nil {
		log.Fatal("Unable to open TUN interface:", err)
	}

	// 1. Get IP address from kubectl/remote
	// 2. Get list of IPs from kubectl (and subscribe to updates)
	// 3. Get resolv.conf from remote
	// 4. Pass traffic over DTLS to
}
