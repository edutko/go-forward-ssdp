package ssdp

import (
	"log"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Relay struct {
	listeners             []Listener
	senders               []Sender
	throttleCheckInterval time.Duration
	throttlePacketLimit   uint64
}

func NewRelay(in []net.Interface, out []net.Interface) (Relay, error) {
	r := Relay{
		listeners:             []Listener{},
		senders:               []Sender{},
		throttleCheckInterval: 500 * time.Millisecond,
		throttlePacketLimit:   250,
	}

	var e error
	for _, ifi := range in {
		l, err := NewListener(ifi, "udp4")
		if err != nil {
			e = err
			break
		}
		r.listeners = append(r.listeners, l)

		// Go does not currently support listening for UDPv6 multicast on Windows
		if runtime.GOOS != "windows" {
			l, err = NewListener(ifi, "udp6")
			if err != nil {
				e = err
				break
			}
			r.listeners = append(r.listeners, l)
		}
	}

	for _, ifi := range out {
		s, err := NewSender(ifi, "udp4")
		if err != nil {
			e = err
			break
		}
		r.senders = append(r.senders, s)

		// We're not listening on IPv6 on Windows, so no need to send
		if runtime.GOOS != "windows" {
			s, err := NewSender(ifi, "udp6")
			if err != nil {
				e = err
				break
			}
			r.senders = append(r.senders, s)
		}
	}

	if e != nil {
		_ = r.close()
	}

	return r, e
}

func (r Relay) Serve() error {
	wg := sync.WaitGroup{}
	msgs := make(chan Message, len(r.listeners)*2)
	errs := make(chan error, len(r.listeners))

	for _, l := range r.listeners {
		go l.Listen(msgs, errs, &wg)
	}

	err := r.serve(msgs, errs)
	_ = r.close()
	wg.Wait()

	return err
}

func (r Relay) serve(messages <-chan Message, errs <-chan error) error {
	var packetCount uint64
	tick := time.Tick(r.throttleCheckInterval)
	for {
		select {
		case <-tick:
			packetCount = 0
		case m := <-messages:
			packetCount++
			if packetCount > r.throttlePacketLimit {
				log.Println("warning: too many packets per second; dropping packet")
			} else {
				r.relay(m)
			}
		case e := <-errs:
			return e
		}
	}
}

func (r Relay) relay(m Message) {
	host, ps, err := net.SplitHostPort(m.SourceIP.String())
	if err != nil {
		log.Printf("error splitting host and port: %s\n", err.Error())
	}
	port, err := strconv.Atoi(ps)
	if err != nil {
		log.Printf("error parsing port: %s\n", err.Error())
	}
	for _, s := range r.senders {
		if s.network == m.Network && s.ifi.Name != m.IfName {
			_, err := s.Send(m.Data, net.ParseIP(host), port)
			if err != nil {
				log.Printf("error relaying packet from %s: %s\n", m.SourceIP.String(), err.Error())
			}
		}
	}
}

func (r Relay) close() error {
	for _, l := range r.listeners {
		_ = l.Close()
	}
	return nil
}
