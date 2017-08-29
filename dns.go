package main

import (
	"fmt"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type dnsServerHandler struct {
	ds *dns.Server
}

// TODO: Move in-memory mapping toa real datastore
var records = map[string]string{
	"foo.com.": "192.168.0.1",
}

func parseQuery(m *dns.Msg) {
	rl := NewLogger(log.New().Writer())

	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			rl.Info("Query for: ", q.Name)
			ip := records[q.Name]
			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else {
				rl.Error("Record: ", q.Name, " not found")
			}
		default:
			rl.Info("Query type: ", q.Qtype, " is currently not supported")
		}
	}
}

func serve(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	// TODO: Enable over the wire compression
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}
	w.WriteMsg(m)
}

func newDNSServerHandler(addr string, port int, conntype string) *dns.Server {
	return &dns.Server{
		Addr: fmt.Sprintf("%s:%d", addr, port),
		Net:  conntype,
	}
}

func newDefaultDNSServerHandler() *dnsServerHandler {
	return &dnsServerHandler{
		ds: newDNSServerHandler(DNS_SERVER_ADDR, DNS_SERVER_PORT, "udp"),
	}
}

func (h *dnsServerHandler) RouteDNS() error {
	rl := NewLogger(log.New().Writer())

	// TODO: Add more domains to the serving list
	dns.HandleFunc("com.", serve)
	err := h.ds.ListenAndServe()
	if err != nil {
		rl.Error("Failed to start DNS server", err)
		return err
	}
	return nil
}
