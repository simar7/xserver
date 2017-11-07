package main

import (
	"fmt"

	"github.com/miekg/dns"
	"math/rand"
	log "github.com/sirupsen/logrus"
	ns1 "gopkg.in/ns1/ns1-go.v2/rest"
	"net/http"
	"time"
	"os"
	"strings"
)

var ns1Client *ns1.Client

type dnsServerHandler struct {
	ds *dns.Server
}

func getRandomAnswer(max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max - 0) + 0
}

func parseQuery(m *dns.Msg) {
	rl := NewLogger(log.New().Writer())

	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			zone := q.Name[strings.IndexAny(q.Name, ".")+1:len(q.Name)-1]
			domain := q.Name[:strings.LastIndex(q.Name, ".")]

			rl.Info("Query for > ", "domain: ", domain, " zone: ", zone, " type: ", q.Qtype)
			record, _, err := ns1Client.Records.Get(zone, domain, "A")
			if err != nil {
				rl.Error("Record: ", q.Name, err)
			} else {
				rl.Info(record.Answers)
				var answers []string
				for _, answer := range record.Answers {
					answers = append(answers, answer.String())
				}
				// TODO: Implement RR answer retrieval
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, answers[getRandomAnswer(len(answers))]))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				} else {
					rl.Error(err)
				}
			}

		// TODO: Add IPv6 support
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
func init() {
	k := os.Getenv("NS1_APIKEY")
	if k == "" {
		fmt.Println("NS1_APIKEY environment variable is not set, stopping DNS server")
	}
	httpClient := &http.Client{Timeout: time.Second * 10}
	ns1Client = ns1.NewClient(httpClient, ns1.SetAPIKey(k))
}

func (h *dnsServerHandler) RouteDNS() error {
	rl := NewLogger(log.New().Writer())

	dns.HandleFunc(".", serve)
	err := h.ds.ListenAndServe()
	if err != nil {
		rl.Error("Failed to start DNS server", err)
		return err
	}
	return nil
}
