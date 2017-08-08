package main

import (
	"fmt"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type dnsServerHandler struct {
	ds *dns.Server
}

func serve(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	// TODO: Enable over the wire compression
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		// TODO: Return a valid reply for a query
		return
	}
	w.WriteMsg(m)
}

func newDNSServer(addr string, port int, conntype string) *dns.Server {
	return &dns.Server{
		Addr: fmt.Sprintf("%s:%d", addr, port),
		Net:  conntype,
	}
}

func newDefaultDNSServer() *dns.Server {
	return &dns.Server{
		Addr: fmt.Sprintf("%s:%d", DNS_SERVER_ADDR, DNS_SERVER_PORT),
		Net:  "udp",
	}
}

func (h *dnsServerHandler) RouteDNS() {
	logger := log.New()
	rl := NewLogger(logger.Writer())

	dns.HandleFunc("service.", serve)
	err := h.ds.ListenAndServe()
	if err != nil {
		rl.Error("Failed to start DNS server", err)
		// FIXME: This needs to be blocking in nature
		//return err
	}
	defer h.ds.Shutdown()
}
