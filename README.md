# TeleTUN: A L3 VPN for Kubernetes

## Usage

 - `teletun client [-kubeconfig=FILE] -pod=PODNAME -port=PORT -server-pubkey=FILE -client-pubkey=FILE -client-privkey=FILE`
 - `teletun server-external -port=PORT -server-pubkey=FILE -server-privkey=FILE -client-pubkey=FILE -sockfile=FILE`
 - `teletun server-cluster -sockfile=FILE`
 - `teletun cleanup`

## Requirements:

General:
 - Go 1.11 (at compile time)
 - `socat`, compiled with OpenSSL DTLS support (at run time)
 - A functioning `resolvconf`(8) (at run time)

macOS:
 - macOS >= 10.6.8 ("Snow Leopard", 2009): For "utun"
 - macOS >= 10.10 ("Yosemite", 2014): Oldest version supported by Go 1.11

GNU/Linux:
 - Linux >= 2.2 (1999): For "tun" network driver
 - Linux >= 2.0 (YYYY): For sockio-base route control

## Positive aspects

 - Local cleanup is very robust--almost all of it is done
   automatically by the operating system kernel when the client
   exits.  The exception to this is that `/etc/resolv.conf` may be
   left in a bad state if we are killed unexpectedly; but we can
   robustly repair the state with `teletun cleanup`.
 - Works robustly with VPNs and other network configuration.
   Exception: assumes a functioning `resolvconf`(8); users with fancy
   DNS setups may have disabled or broken `resolvconf`--they knew what
   they were getting themselves in to.
 - Can have multiple instances connecting to multiple clusters at
   once.
 - Network performance should be good.

## Wire protocol

 - client: dials server on TCP+TLS, using the given certs
 - server: sends over that TCP+TLS connection an ASCII-text-formatted
   list IP addresses (1 IPv4, and any number of IPv6) separated by
   newline.  The list is terminated by a null-byte.
 - server: sends over that TCP+TLS connection the contents of
   `/etc/resolv.conf`.  This is terminated by a null-byte.
 - client: dials server on UDP+DTLS, using the given certs
 - client and server send raw L3 packets back and forth over that
   UDP+DTLS connection.  There is no signalling to indicate which L3
   protocol a packet is; Linux is smart enough to guess, and we are
   smart enough to guess on macOS.
 - Though the TCP+TLS connection is no longer used to exchange data,
   it is kept alive.  The client or server may disconnect by closing
   the connection.

## How it works

Client:

 - Creates a TUN device (a L3 virtual interface) that will show up in
   `ifconfig` just like any other network interface.  On Linux, it is
   named `tel0` (or `tel1`...); on macOS it is named `utun0` (or
   `utun`...) because macOS doesn't let us set the name.
 - Uses the Kubernetes client library to get a list of all cluster
   IPs, and add direct routes for those to the the TUN interface.
 - Sets the IP address of the TUN device to the cluster IP of the
   server.
 - Adds the `resolv.conf` from the server as an interface-specific
   `resolv.conf` for the TUN device.

Server:

 - The server comes in 2 parts: server-external and server-cluster.
   They run in separate pods on the same node, and communicate with
   eachother using an AF_UNIX socket.

Server-external:

 - Reads server-internal's IP and `resolv.conf` over AF_UNIX socket
 - Sends that information to the client over a TCP socket
 - Pumps data between the TCP socket and the AF_UNIX socket

Server-cluster:

 - The client will "think" that it is this pod in the cluster
 - Proxies data between the AF_UNIX socket and a TUN device that is
   bridged with the real ethernet device.
 - Because the client's VPN IP is the same as the real ethernet IP, we
   don't need to muck with promiscuous mode or munging IPs or
   anything.

## FAQ

### How is this different than sshuttle/teleproxy?

 - TeleTUN is a for-realsies VPN.  sshuttle (and teleproxy, which
   mimics it) isn't a real VPN.  Avery Pennarun (the author of
   sshuttle) is very careful to never say it's a VPN, it's a new
   category of thing that does VPN-ish things.  I'd say it's a "really
   weird userspace firewall, that acts like a VPN".

### How is this different than other VPN softwares?

 - With a typical VPN, connecting to the VPN adds the client host to
   the network.
 - With TeleTUN, Connecting to the VPN does NOT add the client as a
   host to the cluster network.  Instead, it effectively replaces the
   server on the cluster network.
   * This means that the client and the server cannot talk to
     eachother via the cluster network, because they are the *same* IP
     address on the cluster network.  This is why the server is a
     2-pod pair; one pod that the client can talk to, and the other
     pod that the client "replaces".
   * This is because we want Kubernetes (not the VPN server) to be
     responsible for adding hosts to the cluster network.

### Why does it have to handle traffic in both directions?

 - Because it acts at L3, not L4.  We don't have a concept of
   connections in order to know how to route the inbound part of an
   outbound TCP connection.

## Known problems (things that should be fixed):

 - Uses `socat` to accomplish DTLS, instead of binding to OpenSSL
 - With some `resolvconf` implementations, being killed unexpectedly
   can leave `/etc/resolv.conf` in a bad state.  This can be fixed by
   running `teletun cleanup`.  This is not a problem with
   systemd-resolvconf.
 - Wire protocol handles out-of-order UDP+DTLS packets very poorly
   when a forwarded L3 packet gets split across multiple DTLS packets.

## Problems declared to be out-of-scope:

 - Deploying the server to the Kubernetes cluster; that is the job of
   a higher-level tool.
 - Configuring Kubernetes to route cluster traffic to the server; that
   is the job of whatever deploys the server.
 - Removing the server pod when done.
 - Key generation
 - Key exchange; it is assumed that the server already has the
   client's pubkey, and that the client already has the server's
   pubkey.
 - The server can only serve 1 client at a time.  If multiple clients
   wish to connect to the cluster, they will each need to deploy their
   own server.
 - Overlapping/conflicting IP ranges when connecting to multiple
   clusters.

## Future directions

 - Consider using WireGuard protocol (but NOT userspace tools).  There
   are both kernelspace and userspace (in Go!) implementations.
   * Supports roaming (close laptop, go home, open laptop on home
     WiFi.  TeleTUN connection still works).
   * Would solve the problem of out-of-order UDP+DTLS packets.
 - This is a good candidate for testbench; very clear parameters to
   change: At least 3 different `resolvconf` implementations.

## Links:

 - Go TUN setup example: https://git.zx2c4.com/wireguard-go/
 - Tinc VPN uses a similar TCP for OOB signalling, UDP for traffic
   wire protocol.
