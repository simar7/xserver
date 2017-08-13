package main

import (
	"os"
	"sync"

	dhcp "github.com/krolaw/dhcp4"
	log "github.com/sirupsen/logrus"
)

func main() {
	// WaitGroup for multiple servers on separate co-routines
	wg := &sync.WaitGroup{}

	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	log.WithFields(log.Fields{})

	rl := NewLogger(log.New().Writer())

	dhcpHandler := newDHCPServer()
	dnsHandler := dnsServerHandler{
		ds: newDefaultDNSServer(),
	}

	go func() {
		wg.Add(1)
		rl.Fatal(dhcp.ListenAndServe(dhcpHandler))
		wg.Done()
	}()
	rl.Infof("DHCP on %s:%d", DHCP_SERVER_ADDR, DHCP_SERVER_PORT)

	go func() {
		wg.Add(1)
		rl.Fatal(dnsHandler.RouteDNS())
		wg.Done()
	}()
	rl.Infof("DNS  on %s:%d", DNS_SERVER_ADDR, DNS_SERVER_PORT)

	rl.Infof("xserver is running and serving: %s, %s", "DHCP", "DNS")
	wg.Wait()
}
