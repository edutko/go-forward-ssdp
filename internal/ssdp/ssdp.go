package ssdp

import (
	"fmt"
	"net"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var (
	ipv4UDPAddr = &net.UDPAddr{IP: net.ParseIP("239.255.255.250"), Port: 1900}
	ipv6LinkLocalUDPAddr = &net.UDPAddr{IP: net.ParseIP("ff02::c"), Port: 1900}
)

type Message struct {
	Network  string
	IfName   string
	SourceIP net.Addr
	Data     []byte
}

type Listener struct {
	network string
	conn    *net.UDPConn
	ifi     net.Interface
	buf     []byte
}

func NewListener(ifi net.Interface, network string) (Listener, error) {
	addr, err := getMulticastUDPAddr(network)
	if err != nil {
		return Listener{}, fmt.Errorf("resolving UDP address: %w", err)
	}

	conn, err := net.ListenMulticastUDP(network, &ifi, addr)
	if err != nil {
		return Listener{}, err
	}

	return Listener{network, conn, ifi, make([]byte, 65535)}, nil
}

func (l Listener) Listen(messages chan<- Message, errs chan<- error, wg *sync.WaitGroup) {
	wg.Add(1)
	for {
		n, addr, err := l.conn.ReadFrom(l.buf)
		if err != nil {
			errs <- err
			wg.Done()
			return
		}
		msg := make([]byte, n)
		copy(msg, l.buf)
		messages <- Message{
			Network:  l.network,
			IfName:   l.ifi.Name,
			SourceIP: addr,
			Data:     msg,
		}
	}
}

func (l Listener) Close() error {
	return l.conn.Close()
}

type Sender struct {
	network string
	ifi     net.Interface
}

func NewSender(ifi net.Interface, network string) (Sender, error) {
	return Sender{network: network, ifi: ifi}, nil
}

func (s Sender) Send(data []byte, srcIP net.IP, srcPort int) (int, error) {
	switch s.network {
	case "udp4":
		return s.sendIPv4(data, srcIP, srcPort)

	case "udp6":
		return s.sendIPv6(data, srcIP, srcPort)

	default:
		return 0, fmt.Errorf("unsupported network: %s", s.network)
	}
}

func getMulticastUDPAddr(network string) (*net.UDPAddr, error) {
	switch network {
	case "udp4":
		return ipv4UDPAddr, nil
	case "udp6":
		return ipv6LinkLocalUDPAddr, nil
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
}

func (s Sender) sendIPv4(data []byte, srcIP net.IP, srcPort int) (int, error) {
	iph, payload, err := buildIPv4Packet(srcIP, srcPort, data)
	if err != nil {
		return 0, fmt.Errorf("building packet: %w", err)
	}

	conn, err := net.ListenIP("ip4:udp", nil)
	if err != nil {
		return 0, fmt.Errorf("connecting: %w", err)
	}
	defer conn.Close()

	rConn, err := ipv4.NewRawConn(conn)
	if err != nil {
		return 0, fmt.Errorf("creating raw connection: %w", err)
	}
	defer rConn.Close()

	err = rConn.SetMulticastInterface(&s.ifi)
	if err != nil {
		return 0, fmt.Errorf("setting multicast interface: %w", err)
	}
	err = rConn.SetMulticastLoopback(false)
	if err != nil {
		return 0, fmt.Errorf("disabling multicast loopback: %w", err)
	}
	err = rConn.SetMulticastTTL(1)
	if err != nil {
		return 0, fmt.Errorf("setting multicast TTL: %w", err)
	}

	return len(data), rConn.WriteTo(iph, payload, nil)
}

func (s Sender) sendIPv6(data []byte, srcIP net.IP, srcPort int) (int, error) {
	packet, cm, err := buildIPv6Packet(srcIP, srcPort, data)
	if err != nil {
		return 0, fmt.Errorf("building packet: %w", err)
	}

	conn, err := net.ListenIP("ip6:udp", nil)
	if err != nil {
		return 0, fmt.Errorf("connecting: %w", err)
	}
	defer conn.Close()

	pConn := ipv6.NewPacketConn(conn)
	defer pConn.Close()

	err = pConn.SetMulticastInterface(&s.ifi)
	if err != nil {
		return 0, fmt.Errorf("setting multicast interface: %w", err)
	}
	err = pConn.SetMulticastLoopback(false)
	if err != nil {
		return 0, fmt.Errorf("disabling multicast loopback: %w", err)
	}
	err = pConn.SetMulticastHopLimit(1)
	if err != nil {
		return 0, fmt.Errorf("setting multicast hop limit: %w", err)
	}

	return pConn.WriteTo(packet, cm, ipv6LinkLocalUDPAddr)
}

func buildIPv4Packet(srcIP net.IP, srcPort int, data []byte) (*ipv4.Header, []byte, error) {
	ip4 := &layers.IPv4{
		Version:  4,
		TTL:      1,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    srcIP,
		DstIP:    ipv4UDPAddr.IP,
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(srcPort),
		DstPort: layers.UDPPort(1900),
	}
	err := udp.SetNetworkLayerForChecksum(ip4)
	if err != nil {
		return nil, nil, fmt.Errorf("setting layer for checksum: %w", err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	err = gopacket.SerializeLayers(buf, opts, udp, gopacket.Payload(data))
	if err != nil {
		return nil, nil, err
	}
	payload := buf.Bytes()

	iph := &ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TotalLen: ipv4.HeaderLen + len(payload),
		TTL:      1,
		Protocol: 17,
		Src:      srcIP.To4(),
		Dst:      ipv4UDPAddr.IP.To4(),
	}

	return iph, payload, nil
}

func buildIPv6Packet(srcIP net.IP, srcPort int, data []byte) ([]byte, *ipv6.ControlMessage, error) {
	ip6 := &layers.IPv6{
		SrcIP:      srcIP,
		DstIP:      ipv6LinkLocalUDPAddr.IP,
		NextHeader: layers.IPProtocolUDP,
		Version:    6,
		HopLimit:   1,
	}
	ip6.LayerType()

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(srcPort),
		DstPort: layers.UDPPort(1900),
	}
	err := udp.SetNetworkLayerForChecksum(ip6)
	if err != nil {
		return nil, nil, fmt.Errorf("setting layer for checksum: %w", err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	err = gopacket.SerializeLayers(buf, opts, udp, gopacket.Payload(data))
	if err != nil {
		return nil, nil, err
	}
	packet := buf.Bytes()

	cm := &ipv6.ControlMessage{
		HopLimit: 1,
		Src: srcIP,
	}

	return packet, cm, fmt.Errorf("IPv6 ")
}
