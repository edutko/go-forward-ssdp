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

func WithNames(names ...string) QueryParam {
	return func(q *interfaceQuery) error {
		q.names = append(q.names, names...)
		return nil
	}
}

func WithIP(ip string) QueryParam {
	return WithIPs(ip)
}

func WithIPs(ips ...string) QueryParam {
	return func(q *interfaceQuery) error {
		q.ips = append(q.ips, ips...)
		return nil
	}
}

func WithMAC(mac string) QueryParam {
	return WithMACs(mac)
}

func WithMACs(macs ...string) QueryParam {
	return func(q *interfaceQuery) error {
		hwAddrs, err := parseMacs(macs)
		if err != nil {
			return err
		}
		q.macs = append(q.macs, hwAddrs...)
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
		q.hasIPv4 = &b
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
		q.hasPublicIPv4 = &b
		return nil
	}
}

func HasIPv6Address() QueryParam {
	return hasIPv6Address(true)
}

func HasNoIPv6Address() QueryParam {
	return hasIPv6Address(false)
}

func hasIPv6Address(b bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.hasIPv6 = &b
		return nil
	}
}

func HasPublicIPv6Address() QueryParam {
	return hasPublicIPv6Address(true)
}

func HasNoPublicIPv6Address() QueryParam {
	return hasPublicIPv6Address(false)
}

func hasPublicIPv6Address(b bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.hasPublicIPv6 = &b
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

func IsBroadcast() QueryParam {
	return isBroadcast(true)
}

func IsNotBroadcast() QueryParam {
	return isBroadcast(false)
}

func isBroadcast(b bool) QueryParam {
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

func IsPointToPoint() QueryParam {
	return isPointToPoint(true)
}

func IsNotPointToPoint() QueryParam {
	return isPointToPoint(false)
}

func isPointToPoint(ptp bool) QueryParam {
	return func(q *interfaceQuery) error {
		q.isPointToPoint = &ptp
		return nil
	}
}

func IsMulticast() QueryParam {
	return isMulticast(true)
}

func IsNotMulticast() QueryParam {
	return isMulticast(false)
}

func isMulticast(m bool) QueryParam {
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
	hasIPv4        *bool
	hasPublicIPv4  *bool
	hasIPv6        *bool
	hasPublicIPv6  *bool
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

	matchedIP := false
	hasIPv4 := false
	hasPublicIPv4 := false
	hasIPv6 := false
	hasPublicIPv6 := false
	if len(q.ips) > 0 || q.hasIPv4 != nil || q.hasPublicIPv4 != nil || q.hasIPv6 != nil || q.hasPublicIPv6 != nil {
		addrs, err := getAddrsForInterface(iface)
		if err != nil && q.FailOnError {
			return false, fmt.Errorf("filtering interfaces: %w", err)
		}

		for _, a := range addrs {
			switch t := a.(type) {
			case *net.IPNet:
				if stringArrayContains(q.ips, t.IP.String()) {
					matchedIP = true
				}
				if t.IP.To4() != nil {
					hasIPv4 = true
					if !isPrivateIPv4(t.IP) {
						hasPublicIPv4 = true
					}
				} else if t.IP.To16() != nil {
					hasIPv6 = true
					if !isPrivateIPv6(t.IP) {
						hasPublicIPv6 = true
					}
				}
			}
		}
	}

	if len(q.ips) > 0 && !matchedIP {
		return false, nil
	}
	if q.hasIPv4 != nil && hasIPv4 != *q.hasIPv4 {
		return false, nil
	}
	if q.hasPublicIPv4 != nil && hasPublicIPv4 != *q.hasPublicIPv4 {
		return false, nil
	}
	if q.hasIPv6 != nil && hasIPv6 != *q.hasIPv6 {
		return false, nil
	}
	if q.hasPublicIPv6 != nil && hasPublicIPv6 != *q.hasPublicIPv6 {
		return false, nil
	}

	return true, nil
}

func flagMatches(desired *bool, iface net.Interface, flag net.Flags) bool {
	return desired == nil || *desired == hasFlag(iface, flag)
}

var getAddrsForInterface = func(iface net.Interface) ([]net.Addr, error) {
	return iface.Addrs()
}
