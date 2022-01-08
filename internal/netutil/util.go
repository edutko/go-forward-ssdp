package netutil

import (
	"bytes"
	"fmt"
	"net"
)

var privateIPv4CIDRs []*net.IPNet
var privateIPv6CIDRs []*net.IPNet

func init() {
	for _, s := range []string{"10.0.0.0/8", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.168.0.0/16"} {
		_, c, err := net.ParseCIDR(s)
		if err != nil {
			panic(err)
		}
		privateIPv4CIDRs = append(privateIPv4CIDRs, c)
	}

	for _, s := range []string{"::1/128", "fe80::/10", "fc00::/7"} {
		_, c, err := net.ParseCIDR(s)
		if err != nil {
			panic(err)
		}
		privateIPv6CIDRs = append(privateIPv6CIDRs, c)
	}
}

func isPrivateIPv4(i net.IP) bool {
	v4 := i.To4()
	if v4 != nil {
		for _, c := range privateIPv4CIDRs {
			if c.Contains(i) {
				return true
			}
		}
	}
	return false
}

func isPrivateIPv6(i net.IP) bool {
	for _, c := range privateIPv6CIDRs {
		if c.Contains(i) {
			return true
		}
	}
	return false
}

func parseMacs(macs []string) ([]net.HardwareAddr, error) {
	var parsedMacs []net.HardwareAddr
	for _, s := range macs {
		mac, err := net.ParseMAC(s)
		if err != nil {
			return nil, fmt.Errorf("listing interfaces: invalid hardware address (\"%s\") in query: %w", mac, err)
		}
		parsedMacs = append(parsedMacs, mac)
	}
	return parsedMacs, nil
}

func hasFlag(iface net.Interface, flag net.Flags) bool {
	return iface.Flags&flag != 0
}

func stringArrayContains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

func bytesArrayContains(a []net.HardwareAddr, b net.HardwareAddr) bool {
	for _, v := range a {
		if bytes.Compare(v, b) == 0 {
			return true
		}
	}
	return false
}
