package main

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
	dhcp "github.com/krolaw/dhcp4"
)

func main() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("DEBUG"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	log.Printf("xserver is running and serving: %s, %s, %s", "DHCP", "TFTP", "DNS")
	log.Printf("DHCP on %s", DHCP_SERVER_ADDR)
	log.Printf("TFTP on %s", TFTP_SERVER_ADDR)
	log.Printf("DNS  on %s", DNS_SERVER_ADDR)

	dhcpHandler := newDHCPServer()

	// TODO: Add multi interface support with dhcp.ListenAndServeIf()
	log.Fatal(dhcp.ListenAndServe(dhcpHandler))
}
