package netutil

import (
	"fmt"
	"net"
	"strings"
)

var InterfaceFlags = [...]net.Flags{
	net.FlagUp,
	net.FlagBroadcast,
	net.FlagLoopback,
	net.FlagPointToPoint,
	net.FlagMulticast,
}

func InterfaceToString(iface net.Interface) string {
	flags := make([]string, 0)
	for _, f := range InterfaceFlags {
		if iface.Flags&f != 0 {
			flags = append(flags, f.String())
		}
	}

	var unicast []string
	addrs, err := iface.Addrs()
	if err != nil {
		unicast = append(unicast, "error: " + err.Error())
	} else {
		for _, addr := range addrs {
			unicast = append(unicast, addr.String())
		}
	}

	var multicast []string
	addrs, err = iface.MulticastAddrs()
	if err != nil {
		multicast = append(multicast, "error: " + err.Error())
	} else {
		for _, addr := range addrs {
			multicast = append(multicast, addr.String())
		}
	}

	format := "%s (%s)\n" +
		"  Flags: %s\n" +
		"  Unicast addresses:\n" +
		"    %s\n" +
		"  Multicast addresses:\n" +
		"    %s\n"
	return fmt.Sprintf(format,
		iface.Name, iface.HardwareAddr.String(),
		strings.Join(flags, ", "),
		strings.Join(unicast, "\n    "),
		strings.Join(multicast, "\n    "))
}

func GetInterfaces(params ...QueryParam) ([]net.Interface, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("listing interfaces: %w", err)
	}
	return FilterInterfaces(ifs, params...)
}
