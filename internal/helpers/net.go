package helpers

import "net"

// MyIPAddresses returns slice of active ip addresses on PC.
func MyIPAddresses() []net.IP {
	ips := make([]net.IP, 0)
	ips = append(ips, net.IPv4(127, 0, 0, 1), net.IPv6loopback)
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, er := iface.Addrs()
		if er != nil {
			continue
		}
		for _, addr := range addrs {
			ip := addr.(*net.IPNet).IP
			ips = append(ips, ip)
		}
	}
	return ips
}
