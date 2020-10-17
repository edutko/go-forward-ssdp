package netutil

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

type QueryParam func(q *interfaceQuery) error

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

	if len(params) == 0 {
		return ifs, nil
	}

	var q interfaceQuery
	for _, p := range params {
		if err = p(&q); err != nil {
			return nil, fmt.Errorf("listing interfaces: %w", err)
		}
	}

	matchingIfs := make([]net.Interface, 0)
	for _, iface := range ifs {
		var addrs []net.Addr
		if len(q.ips) > 0 {
			addrs, err = iface.Addrs()
			if err != nil && q.FailOnError {
				return nil, fmt.Errorf("listing interfaces: %w", err)
			}
		}
		if matches(q, iface, addrs) {
			matchingIfs = append(matchingIfs, iface)
		}
	}

	return matchingIfs, nil
}

func WithName(name string) QueryParam {
	return func(q *interfaceQuery) error {
		q.names = append(q.names, name)
		return nil
	}
}

func WithNames(names []string) QueryParam {
	return func(q *interfaceQuery) error {
		q.names = append(q.names, names...)
		return nil
	}
}

func WithIP(ip string) QueryParam {
	return func(q *interfaceQuery) error {
		q.ips = append(q.ips, ip)
		return nil
	}
}

func WithIPs(ips []string) QueryParam {
	return func(q *interfaceQuery) error {
		q.ips = append(q.ips, ips...)
		return nil
	}
}

func WithMAC(mac string) QueryParam {
	return WithMACs([]string{mac})
}

func WithMACs(macs []string) QueryParam {
	return func(q *interfaceQuery) error {
		hwAddrs, err := parseMacs(macs)
		if err != nil {
			return err
		}
		q.macs = append(q.macs, hwAddrs...)
		return nil
	}
}

func IsUp(up bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.isUp = &up
		return nil
	}
}

func IsDown() QueryParam {
	return IsUp(false)
}

func IsBroadcast(b bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.isBroadcast = &b
		return nil
	}
}

func IsLoopback(lo bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.isLoopback = &lo
		return nil
	}
}

func IsPointToPoint(ptp bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.isPointToPoint = &ptp
		return nil
	}
}

func IsMulticast(m bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.isMulticast = &m
		return nil
	}
}

type interfaceQuery struct {
	names          []string
	macs           []net.HardwareAddr
	ips            []string
	isUp           *bool
	isBroadcast    *bool
	isLoopback     *bool
	isPointToPoint *bool
	isMulticast    *bool
	FailOnError    bool
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

func matches(q interfaceQuery, iface net.Interface, addrs []net.Addr) bool {
	if !flagMatches(q.isUp, iface, net.FlagUp) {
		return false
	}
	if !flagMatches(q.isBroadcast, iface, net.FlagBroadcast) {
		return false
	}
	if !flagMatches(q.isLoopback, iface, net.FlagLoopback) {
		return false
	}
	if !flagMatches(q.isPointToPoint, iface, net.FlagPointToPoint) {
		return false
	}
	if !flagMatches(q.isMulticast, iface, net.FlagMulticast) {
		return false
	}

	if len(q.names) > 0 {
		if !stringArrayContains(q.names, iface.Name) {
			return false
		}
	}

	if len(q.macs) > 0 {
		if !bytesArrayContains(q.macs, iface.HardwareAddr) {
			return false
		}
	}

	if len(q.ips) > 0 {
		matchedIP := false
		for _, a := range addrs {
			switch t := a.(type) {
			case *net.IPNet:
				if stringArrayContains(q.ips, t.IP.String()) {
					matchedIP = true
					break
				}
			}
		}
		if !matchedIP {
			return false
		}
	}

	return true
}

func flagMatches(desired *bool, iface net.Interface, flag net.Flags) bool {
	return desired == nil || *desired == hasFlag(iface, flag)
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
