package main

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"

	"github.com/stretchr/testify/assert"
)

func RunLocalUDPServer(laddr string) (*dns.Server, string, error) {
	server, l, _, err := RunLocalUDPServerWithFinChan(laddr)

	return server, l, err
}

func RunLocalUDPServerWithFinChan(laddr string) (*dns.Server, string, chan struct{}, error) {
	pc, err := net.ListenPacket("udp", laddr)
	if err != nil {
		return nil, "", nil, err
	}
	server := &dns.Server{PacketConn: pc, ReadTimeout: time.Hour, WriteTimeout: time.Hour}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	fin := make(chan struct{}, 0)

	go func() {
		server.ActivateAndServe()
		close(fin)
		pc.Close()
	}()

	waitLock.Lock()
	return server, pc.LocalAddr().String(), fin, nil
}

func TestDNS_OK(t *testing.T) {
	h := newDefaultDNSServerHandler()
	dns.HandleFunc("test.", serve)
	defer dns.HandleRemove("test.")

	err := h.ds.ListenAndServe()
	if err != nil {
		assert.Error(t, err)
	}

	m := new(dns.Msg)
	m.SetQuestion("go.test.", dns.TypeA)

	s, addrstr, err := RunLocalUDPServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	defer s.Shutdown()

	cn, err := dns.Dial("udp", addrstr)
	if err != nil {
		t.Errorf("failed to dial %s: %v", addrstr, err)
	}
	err = cn.WriteMsg(m)
	assert.NoError(t, err)

	r, err := cn.ReadMsg()
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.NotEqual(t, r, dns.RcodeSuccess)
}
