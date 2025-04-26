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
		ifList, err = netutil.GetInterfaces(netutil.WithNames(os.Args[1:]...))
		if len(ifList) != len(os.Args)-1 {
			log.Fatalln("error: one or more requested interfaces were not found")
		}
	} else {
		ifList, err = netutil.GetInterfaces(
			netutil.IsNotLoopback(), netutil.IsUp(), netutil.HasIPv4Address(), netutil.HasNoPublicIPv4Address(),
		)
	}
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	if len(ifList) == 0 {
		log.Fatalf("No interfaces matched the specified criteria.")
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
