package main

import (
	dhcp "github.com/krolaw/dhcp4"

	"log"
	"math/rand"
	"net"
	"time"
)

type lease struct {
	nic    string    // Client's CHAddr
	expiry time.Time // When the lease expires
}

type dhcpServerHandler struct {
	ip            net.IP        // Server IP to use
	options       dhcp.Options  // Options to send to DHCP Clients
	start         net.IP        // Start of IP range to distribute
	leaseRange    int           // Number of IPs to distribute (starting from start)
	leaseDuration time.Duration // Lease period
	leases        map[int]lease // Map to keep track of leases
}

func newDHCPServer() *dhcpServerHandler {
	return &dhcpServerHandler{
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
}

func (h *dhcpServerHandler) freeLease() int {
	now := time.Now()
	b := rand.Intn(h.leaseRange) // Try random first
	for _, v := range [][]int{[]int{b, h.leaseRange}, []int{0, b}} {
		for i := v[0]; i < v[1]; i++ {
			if l, ok := h.leases[i]; !ok || l.expiry.Before(now) {
				return i
			}
		}
	}
	return -1
}

func (h *dhcpServerHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
	log.Print("message type = ", msgType)
	switch msgType {
	case dhcp.Discover:
		free, nic := -1, p.CHAddr().String()
		for i, v := range h.leases { // Find previous lease
			if v.nic == nic {
				free = i
				goto reply
			}
		}
		if free = h.freeLease(); free == -1 {
			return
		}
	reply:
		return dhcp.ReplyPacket(p, dhcp.Offer, h.ip, dhcp.IPAdd(h.start, free), h.leaseDuration,
			h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))

	case dhcp.Request:
		if server, ok := options[dhcp.OptionServerIdentifier]; ok && !net.IP(server).Equal(h.ip) {
			return nil // Message not for this dhcp server
		}
		reqIP := net.IP(options[dhcp.OptionRequestedIPAddress])
		if reqIP == nil {
			reqIP = net.IP(p.CIAddr())
		}

		if len(reqIP) == 4 && !reqIP.Equal(net.IPv4zero) {
			if leaseNum := dhcp.IPRange(h.start, reqIP) - 1; leaseNum >= 0 && leaseNum < h.leaseRange {
				if l, exists := h.leases[leaseNum]; !exists || l.nic == p.CHAddr().String() {
					h.leases[leaseNum] = lease{nic: p.CHAddr().String(), expiry: time.Now().Add(h.leaseDuration)}
					return dhcp.ReplyPacket(p, dhcp.ACK, h.ip, reqIP, h.leaseDuration,
						h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
				}
			}
		}
		return dhcp.ReplyPacket(p, dhcp.NAK, h.ip, nil, 0, nil)

	case dhcp.Release, dhcp.Decline:
		nic := p.CHAddr().String()
		for i, v := range h.leases {
			if v.nic == nic {
				delete(h.leases, i)
				break
			}
		}
	}
	return nil
}

func main() {
	log.Printf("xserver is running and serving: %s, %s, %s", "DHCP", "TFTP", "DNS")
	log.Printf("DHCP on %s", DHCP_SERVER_ADDR)
	log.Printf("TFTP on %s", TFTP_SERVER_ADDR)
	log.Printf("DNS  on %s", DNS_SERVER_ADDR)

	dhcpHandler := newDHCPServer()

	// TODO: Add multi interface support with dhcp.ListenAndServeIf()
	log.Fatal(dhcp.ListenAndServe(dhcpHandler))
}
