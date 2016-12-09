package main

import (
	"net"
	"testing"
	"time"

	dhcp "github.com/krolaw/dhcp4"
	"github.com/stretchr/testify/assert"
)

func TestNewDHCPServer(t *testing.T) {
	expected := &dhcpServerHandler{
		ip:            net.ParseIP(DHCP_SERVER_ADDR),
		start:         net.ParseIP(DHCP_SERVER_LEASE_START_ADDR),
		leaseDuration: DHCP_LEASE_DURATION * time.Minute,
		leases:        make(map[int]lease, DHCP_LEASE_COUNT),
		leaseRange:    DHCP_LEASE_RANGE,
		options: dhcp.Options{
			dhcp.OptionTFTPServerName:   []byte(TFTP_SERVER_ADDR),
			dhcp.OptionBootFileName:     []byte(PXELINUX_LOADER),
			dhcp.OptionDomainNameServer: []byte(DNS_SERVER_ADDR),
		},
	}

	actual := newDHCPServer()
	assert.Equal(t, expected, actual)
}

func TestFreeLease_OK(t *testing.T) {
	dhcpServer := newDHCPServer()
	NotExpected := -1
	actual := dhcpServer.freeLease()
	assert.NotEqual(t, NotExpected, actual)
}

func TestServeDHCP(t *testing.T) {
	dhcpServer := newDHCPServer()
	p := dhcp.NewPacket(dhcp.BootReply)

	expected := dhcp.ReplyPacket(p, dhcp.Offer, dhcpServer.ip,
		dhcp.IPAdd(dhcpServer.start, dhcpServer.freeLease()),
		dhcpServer.leaseDuration, dhcpServer.options.SelectOrderOrAll(nil))
	actual := dhcpServer.ServeDHCP(p, dhcp.Discover, nil)

	//TODO: Need a strong assertion
	assert.ObjectsAreEqual(actual, expected)
}
