package main

import (
	"log"
	"net"
	"os"

	"github.com/edutko/go-forward-ssdp/internal/netutil"
	"github.com/edutko/go-forward-ssdp/internal/ssdp"
)

func main() {
	var ifList []net.Interface
	var err error
	if len(os.Args) > 1 {
		ifList, err = netutil.GetInterfaces(netutil.WithNames(os.Args[1:]))
	} else {
		ifList, err = netutil.GetInterfaces(netutil.IsLoopback(false), netutil.IsUp(true))
	}
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	for _, ifi := range ifList {
		log.Printf("Listening on %s (%s)\n", ifi.Name, ifi.HardwareAddr.String())
	}

	r, err := ssdp.NewRelay(ifList, ifList)
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	err = r.Serve()
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}
}
