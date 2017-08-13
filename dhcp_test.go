package main

import (
	"log"
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

func TestServeDHCPDiscover_OK(t *testing.T) {
	dhcpServer := newDHCPServer()

	expected := dhcp.ReplyPacket(dhcp.NewPacket(dhcp.BootReply), dhcp.Offer, dhcpServer.ip,
		dhcp.IPAdd(dhcpServer.start, dhcpServer.freeLease()),
		dhcpServer.leaseDuration, dhcpServer.options.SelectOrderOrAll(nil))
	actual := dhcpServer.ServeDHCP(dhcp.NewPacket(dhcp.BootReply), dhcp.Discover, nil)

	// TODO: Need a strong assertion
	// See Issue: https://github.com/simar7/xserver/issues/8
	assert.ObjectsAreEqual(expected, actual)
}

func TestServeDHCPInvalidRequest_OK(t *testing.T) {
	dhcpServer := newDHCPServer()

	expected := dhcp.ReplyPacket(dhcp.NewPacket(dhcp.BootReply), dhcp.NAK, dhcpServer.ip, nil, 0, nil)
	actual := dhcpServer.ServeDHCP(dhcp.NewPacket(dhcp.BootReply), dhcp.Request, dhcpServer.options)

	assert.Equal(t, expected, actual)
}

func TestServeDHCPValidRequest_OK(t *testing.T) {
	dhcpServer := newDHCPServer()
	p := dhcp.NewPacket(dhcp.BootReply)
	// Request the IP from the DHCP Server
	p.SetCIAddr(net.ParseIP("127.0.0.2"))

	reqIP := net.IP(dhcpServer.options[dhcp.OptionRequestedIPAddress])
	if reqIP == nil {
		reqIP = net.IP(p.CIAddr())
		log.Print("[DEBUG] ", "Request IP: ", reqIP)
	}

	expected := dhcp.ReplyPacket(p, dhcp.ACK, dhcpServer.ip, reqIP,
		dhcpServer.leaseDuration, dhcpServer.options.SelectOrder(dhcpServer.options[dhcp.OptionParameterRequestList]))
	actual := dhcpServer.ServeDHCP(p, dhcp.Request, dhcp.Options{
		dhcp.OptionTFTPServerName:   []byte(TFTP_SERVER_ADDR),
		dhcp.OptionBootFileName:     []byte(PXELINUX_LOADER),
		dhcp.OptionDomainNameServer: []byte(DNS_SERVER_ADDR),
	})

	assert.Equal(t, expected, actual)
}

func TestServeDHCPRelease_OK(t *testing.T) {
	dhcpServer := newDHCPServer()
	p := dhcp.NewPacket(dhcp.BootReply)
	// Request the IP from the DHCP Server
	p.SetCIAddr(net.ParseIP("172.10.0.2"))

	expected := dhcp.Packet(nil)
	actual := dhcpServer.ServeDHCP(p, dhcp.Release, nil)

	assert.Equal(t, expected, actual)
}
