package main

import (
	"errors"
	"github.com/miekg/dns"
	"log"
	"net"
	"sync"
	"time"
)

type RequestType uint8

const (
	RequestIPv4 RequestType = iota
	RequestIPv6
	RequestOther
)

func HandlerTCP(w dns.ResponseWriter, req *dns.Msg) {
	totalRequestsTcp.Inc()
	Handler(w, req)
}

func HandlerUDP(w dns.ResponseWriter, req *dns.Msg) {
	totalRequestsUdp.Inc()
	Handler(w, req)
}

func Handler(w dns.ResponseWriter, req *dns.Msg) {
	defer w.Close()

	question := req.Question[0]

	var reqType RequestType
	switch question.Qtype {
	case dns.TypeA:
		reqType = RequestIPv4
	case dns.TypeAAAA:
		reqType = RequestIPv6
	default:
		reqType = RequestOther
	}

	cachedReq := cache.Get(question.Qtype, question.Name)
	if cachedReq != nil {
		totalCacheHits.Inc()

		response := &dns.Msg{}
		response.SetReply(req)
		response.Answer = append(response.Answer, cachedReq)

		w.WriteMsg(response)
		totalRequestsSuccess.Inc()
		return
	}

	if reqType != RequestOther && blackList.Contains(question.Name) {
		response := &dns.Msg{}
		response.SetReply(req)

		if reqType == RequestIPv4 {
			response.Answer = append(response.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(config.UpdateInterval.Seconds()),
				},
				A: net.ParseIP(config.BlockAddress4),
			})
		} else {
			response.Answer = append(response.Answer, &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    uint32(config.UpdateInterval.Seconds()),
				},
				AAAA: net.ParseIP(config.BlockAddress6),
			})
		}

		w.WriteMsg(response)
		log.Println("blocked", question.Name)
		totalRequestsBlocked.Inc()
		return
	}

	resp, err := Lookup(req)
	if err != nil {
		resp = &dns.Msg{}
		resp.SetRcode(req, dns.RcodeServerFailure)
		log.Println("fail", question.Name)
		totalRequestsFailed.Inc()
	} else {
		totalRequestsSuccess.Inc()
		if len(resp.Answer) > 0 {
			cache.Set(question.Qtype, question.Name, resp.Answer[0])
		}
	}

	w.WriteMsg(resp)
}

func Lookup(req *dns.Msg) (*dns.Msg, error) {
	c := &dns.Client{
		Net:          "tcp",
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	qName := req.Question[0].Name

	res := make(chan *dns.Msg, 1)
	var wg sync.WaitGroup
	L := func(nameserver string) {
		defer wg.Done()
		r, _, err := c.Exchange(req, nameserver)
		totalRequestsToGoogle.Inc()
		if err != nil {
			log.Printf("%s socket error on %s", qName, nameserver)
			log.Printf("error:%s", err.Error())
			return
		}
		if r != nil && r.Rcode != dns.RcodeSuccess {
			if r.Rcode == dns.RcodeServerFailure {
				return
			}
		}
		select {
		case res <- r:
		default:
		}
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Start lookup on each nameserver top-down, in every second
	for _, nameserver := range config.Nameservers {
		wg.Add(1)
		go L(nameserver)
		// but exit early, if we have an answer
		select {
		case r := <-res:
			return r, nil
		case <-ticker.C:
			continue
		}
	}

	// wait for all the namservers to finish
	wg.Wait()
	select {
	case r := <-res:
		return r, nil
	default:
		return nil, errors.New("can't resolve ip for" + qName)
	}
}
