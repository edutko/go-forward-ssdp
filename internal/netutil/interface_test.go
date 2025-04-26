package netutil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceToString(t *testing.T) {
	ifi := net.Interface{
		Index:        0,
		Name:         "lo0",
		HardwareAddr: net.HardwareAddr{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
		Flags:        net.FlagUp | net.FlagLoopback | net.FlagMulticast,
	}

	s := InterfaceToString(ifi)

	assert.Contains(t, s,
		"lo0 (00:01:02:03:04:05)\n"+
			"  Flags: up, loopback, multicast\n"+
			"  Unicast addresses:\n",
	)
}
