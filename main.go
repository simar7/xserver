package main

import (
	"os"

	dhcp "github.com/krolaw/dhcp4"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	log.WithFields(log.Fields{})

	logger := log.New()
	rl := NewLogger(logger.Writer())

	rl.Infof("xserver is running and serving: %s, %s, %s", "DHCP", "TFTP", "DNS")
	rl.Infof("DHCP on %s", DHCP_SERVER_ADDR)
	rl.Infof("TFTP on %s", TFTP_SERVER_ADDR)
	rl.Infof("DNS  on %s", DNS_SERVER_ADDR)

	dhcpHandler := newDHCPServer()

	// TODO: Add multi interface support with dhcp.ListenAndServeIf()
	rl.Fatal(dhcp.ListenAndServe(dhcpHandler))
}
