package netutil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceToString(t *testing.T) {
	ifi := net.Interface{
		Name:         "en0",
		HardwareAddr: net.HardwareAddr{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
		Flags:        net.FlagUp | net.FlagLoopback | net.FlagMulticast,
	}
	assert.Equal(t,
		"en0 (00:01:02:03:04:05)\n" +
		"  Flags: up, loopback, multicast\n" +
		"  Unicast addresses:\n" +
		"    \n" +
		"  Multicast addresses:\n" +
		"    \n",
		InterfaceToString(ifi))
}

func TestFilterInterfaces(t *testing.T) {
	//ifis := []net.Interface{{
	//	Name:         "en0",
	//	HardwareAddr: net.HardwareAddr{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
	//	Flags:        net.FlagUp | net.FlagLoopback | net.FlagMulticast,
	//},
	//}

	ifis, _ := net.Interfaces()
	filtered, _ := FilterInterfaces(ifis, HasNoPublicIPv4Address())
	assert.Equal(t, []net.Interface{}, filtered)
}