package netutil

import (
	"fmt"
	"net"
)

type QueryParam func(q *interfaceQuery) error

func FilterInterfaces(ifs []net.Interface, params ...QueryParam) ([]net.Interface, error) {
	if len(params) == 0 {
		return ifs, nil
	}

	var q interfaceQuery
	for _, p := range params {
		if err := p(&q); err != nil {
			return nil, fmt.Errorf("filtering interfaces: %w", err)
		}
	}

	matchingIfs := make([]net.Interface, 0)
	for _, iface := range ifs {
		match, err := matches(q, iface)
		if err != nil {
			return nil, err
		} else if match {
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

func IsUp() QueryParam {
	return isUp(true)
}

func IsDown() QueryParam {
	return isUp(false)
}

func isUp(up bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.isUp = &up
		return nil
	}
}

func IsBroadcast(b bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.isBroadcast = &b
		return nil
	}
}

func IsLoopback() QueryParam {
	return isLoopback(true)
}

func IsNotLoopback() QueryParam {
	return isLoopback(false)
}

func isLoopback(lo bool) QueryParam {
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

func HasIPv4Address() QueryParam {
	return hasIPv4Address(true)
}

func HasNoIPv4Address() QueryParam {
	return hasIPv4Address(false)
}

func hasIPv4Address(b bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.hasPublicIP = &b
		return nil
	}
}

func HasPublicIPv4Address() QueryParam {
	return hasPublicIPv4Address(true)
}

func HasNoPublicIPv4Address() QueryParam {
	return hasPublicIPv4Address(false)
}

func hasPublicIPv4Address(b bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.hasPublicIP = &b
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
	hasPublicIP	   *bool
	FailOnError    bool
}

func matches(q interfaceQuery, iface net.Interface) (bool, error) {
	if !flagMatches(q.isUp, iface, net.FlagUp) {
		return false, nil
	}
	if !flagMatches(q.isBroadcast, iface, net.FlagBroadcast) {
		return false, nil
	}
	if !flagMatches(q.isLoopback, iface, net.FlagLoopback) {
		return false, nil
	}
	if !flagMatches(q.isPointToPoint, iface, net.FlagPointToPoint) {
		return false, nil
	}
	if !flagMatches(q.isMulticast, iface, net.FlagMulticast) {
		return false, nil
	}

	if len(q.names) > 0 {
		if !stringArrayContains(q.names, iface.Name) {
			return false, nil
		}
	}

	if len(q.macs) > 0 {
		if !bytesArrayContains(q.macs, iface.HardwareAddr) {
			return false, nil
		}
	}

	if len(q.ips) > 0 || q.hasPublicIP != nil {
		addrs, err := iface.Addrs()
		if err != nil && q.FailOnError {
			return false, fmt.Errorf("filtering interfaces: %w", err)
		}
		matchedIP := false
		foundPublicIPv4 := false
		for _, a := range addrs {
			switch t := a.(type) {
			case *net.IPNet:
				if stringArrayContains(q.ips, t.IP.String()) {
					matchedIP = true
				}
				if t.IP.To4() != nil && !isPrivateIPv4(t.IP) {
					foundPublicIPv4 = true
				}
			}
		}
		if len(q.ips) > 0 && !matchedIP {
			return false, nil
		}
		if q.hasPublicIP != nil && foundPublicIPv4 != *q.hasPublicIP {
			return false, nil
		}
	}

	return true, nil
}

func flagMatches(desired *bool, iface net.Interface, flag net.Flags) bool {
	return desired == nil || *desired == hasFlag(iface, flag)
}
