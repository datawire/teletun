Both macOS utun and Linux TUN/TAP (unless IFF_NO_PI) use a 4-byte
header.  But they aren't the same!

Linux:
 - int16: Flags (<linux/if_tun.h>:TUN_PKT_STRIP is the only flag right
   now; it indicates that the packet was truncated because the buffer
   was too small)
 - int16_be: IEEE 802 protocol number (as in
   https://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml)

macOS:
 - int32_be: address family (as in AF_INET or AF_INET6)

https://github.com/kubernetes/kubernetes/issues/47862

https://github.com/flix-tech/k8s-mdns/
