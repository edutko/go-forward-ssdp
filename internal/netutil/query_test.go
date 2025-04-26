package netutil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterInterfaces_HasIPv4Address(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, HasIPv4Address())
	assert.Len(t, filtered, 2)
}

func TestFilterInterfaces_HasNoIPv4Address(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, HasNoIPv4Address())
	assert.Len(t, filtered, 2)
}

func TestFilterInterfaces_HasPublicIPv4Address(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, HasPublicIPv4Address())
	assert.Len(t, filtered, 0)
}

func TestFilterInterfaces_HasNoPublicIPv4Address(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, HasNoPublicIPv4Address())
	assert.Len(t, filtered, 4)
}

func TestFilterInterfaces_HasIPv6Address(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, HasIPv6Address())
	assert.Len(t, filtered, 2)
}

func TestFilterInterfaces_HasNoIPv6Address(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, HasNoIPv6Address())
	assert.Len(t, filtered, 2)
}

func TestFilterInterfaces_HasPublicIPv6Address(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, HasPublicIPv6Address())
	assert.Len(t, filtered, 1)
}

func TestFilterInterfaces_HasNoPublicIPv6Address(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, HasNoPublicIPv6Address())
	assert.Len(t, filtered, 3)
}

func TestFilterInterfaces_IsBroadcast(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsBroadcast())
	assert.Len(t, filtered, 2)
}

func TestFilterInterfaces_IsNotBroadcast(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsNotBroadcast())
	assert.Len(t, filtered, 2)
}

func TestFilterInterfaces_IsUp(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsUp())
	assert.Len(t, filtered, 3)
}

func TestFilterInterfaces_IsDown(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsDown())
	assert.Len(t, filtered, 1)
}

func TestFilterInterfaces_IsLoopback(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsLoopback())
	assert.Len(t, filtered, 1)
}

func TestFilterInterfaces_IsNotLoopback(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsNotLoopback())
	assert.Len(t, filtered, 3)
}

func TestFilterInterfaces_IsMulticast(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsMulticast())
	assert.Len(t, filtered, 3)
}

func TestFilterInterfaces_IsNotMulticast(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsNotMulticast())
	assert.Len(t, filtered, 1)
}

func TestFilterInterfaces_IsPointToPoint(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsPointToPoint())
	assert.Len(t, filtered, 1)
}

func TestFilterInterfaces_IsNotPointToPoint(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, IsNotPointToPoint())
	assert.Len(t, filtered, 3)
}

func TestFilterInterfaces_WithIP(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, WithIP("127.0.0.1"))
	assert.Len(t, filtered, 1)
}

func TestFilterInterfaces_WithIPs(t *testing.T) {
	getAddrsForInterface = mockGetAddrsForInterface
	filtered, _ := FilterInterfaces(testIfs, WithIPs("::1", "192.168.100.200"))
	assert.Len(t, filtered, 2)
}

func TestFilterInterfaces_WithMAC(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, WithMAC("00:01:02:03:04:05"))
	assert.Len(t, filtered, 1)
}

func TestFilterInterfaces_WithMACs(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, WithMACs("00:01:02:03:04:05", "00:01:02:03:04:06"))
	assert.Len(t, filtered, 2)
}

func TestFilterInterfaces_WithName(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, WithName("lo0"))
	assert.Len(t, filtered, 1)
}

func TestFilterInterfaces_WithNames(t *testing.T) {
	filtered, _ := FilterInterfaces(testIfs, WithNames("lo0", "en0"))
	assert.Len(t, filtered, 2)
}

var testIfs = []net.Interface{
	{
		Index:        0,
		Name:         "lo0",
		HardwareAddr: net.HardwareAddr{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
		Flags:        net.FlagUp | net.FlagLoopback | net.FlagMulticast,
	},
	{
		Index:        1,
		Name:         "en0",
		HardwareAddr: net.HardwareAddr{0x00, 0x01, 0x02, 0x03, 0x04, 0x06},
		Flags:        0,
	},
	{
		Index:        2,
		Name:         "en1",
		HardwareAddr: net.HardwareAddr{0x00, 0x01, 0x02, 0x03, 0x04, 0x07},
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	},
	{
		Index:        3,
		Name:         "utun0",
		HardwareAddr: net.HardwareAddr{0x00, 0x01, 0x02, 0x03, 0x04, 0x08},
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagPointToPoint | net.FlagMulticast,
	},
}

var mockGetAddrsForInterface = func(iface net.Interface) ([]net.Addr, error) {
	addrs := [][]net.Addr{
		{mockIPAddr("127.0.0.1"), mockIPAddr("::1")},
		{},
		{mockIPAddr("a64e:d060:3add:7c04:3158:6898:7538:6cdb")},
		{mockIPAddr("192.168.100.200")},
	}
	return addrs[iface.Index], nil
}

func mockIPAddr(ip string) *net.IPNet {
	return &net.IPNet{IP: net.ParseIP(ip)}
}
