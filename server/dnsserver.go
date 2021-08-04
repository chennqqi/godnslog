package server

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/models"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
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

TODO:
1. 支持IPv6
2. 支持A记录通配
3. 支持SRV记录
4. 支持NS记录
*/

const (
	LOG_TTL              = 0
	NS_TTL               = 600
	DEFAULT_TTL          = 300
	XIP_TTL              = 86400 * 7 // one week
	CUSTOM_REBIND_EXPIRE = 3600 * time.Second
)

var (
	ipv6Regexp = regexp.MustCompile(`^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::(?:[0-9a-fA-F]{1,4}:){0,6}[0-9a-fA-F]{1,4}$|^[0-9a-fA-F]{1,4}::(?:[0-9a-fA-F]{1,4}:){0,5}[0-9a-fA-F]{1,4}$|^[0-9a-fA-F]{1,4}:[0-9a-fA-F]{1,4}::(?:[0-9a-fA-F]{1,4}:){0,4}[0-9a-fA-F]{1,4}$|^(?:[0-9a-fA-F]{1,4}:){0,2}[0-9a-fA-F]{1,4}::(?:[0-9a-fA-F]{1,4}:){0,3}[0-9a-fA-F]{1,4}$|^(?:[0-9a-fA-F]{1,4}:){0,3}[0-9a-fA-F]{1,4}::(?:[0-9a-fA-F]{1,4}:){0,2}[0-9a-fA-F]{1,4}$|^(?:[0-9a-fA-F]{1,4}:){0,4}[0-9a-fA-F]{1,4}::(?:[0-9a-fA-F]{1,4}:)?[0-9a-fA-F]{1,4}$|^(?:[0-9a-fA-F]{1,4}:){0,5}[0-9a-fA-F]{1,4}::[0-9a-fA-F]{1,4}$|^(?:[0-9a-fA-F]{1,4}:){0,6}[0-9a-fA-F]{1,4}::$`)
)

type Resolve struct {
	Name  string
	Type  string
	Value string
	Ttl   uint32
}
type DnsServerConfig struct {
	Domain             string
	RTimeout, WTimeout time.Duration
	V4, V6             net.IP
	Upstream           string
}

type DnsServer struct {
	DnsServerConfig
	store *cache.Cache

	tcpServer *dns.Server
	udpServer *dns.Server

	ipv4Regexp  *regexp.Regexp
	ipv4uRegexp *regexp.Regexp
	ipv4bRegexp *regexp.Regexp

	wg      sync.WaitGroup
	handler dns.Handler
	client  *dns.Client
}

func NewDnsServer(cfg *DnsServerConfig, store *cache.Cache) (*DnsServer, error) {
	domain := cfg.Domain
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	ipv4Exp := `((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))`
	ipv4Exp = ipv4Exp + strings.Replace("."+domain, ".", `\.`, -1)
	ipv4uExp := `(?i:0x)?([0-f]?[0-f]{7})`
	ipv4uExp = ipv4uExp + strings.Replace("."+domain, ".", `\.`, -1)
	ipv4bExp := `(?:0b)?((?:0|1){25,32})`
	ipv4bExp = ipv4bExp + strings.Replace("."+domain, ".", `\.`, -1)

	addr := ""
	// debug
	// addr = ":10053"
	handler := dns.NewServeMux()
	var s = &DnsServer{
		client: &dns.Client{
			SingleInflight: true,
		},
		DnsServerConfig: *cfg,
		store:           store,
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
	s.ipv4Regexp = regexp.MustCompile(ipv4Exp)
	s.ipv4uRegexp = regexp.MustCompile(ipv4uExp)
	s.ipv4bRegexp = regexp.MustCompile(ipv4bExp)

	handler.HandleFunc(domain, s.Do)
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

	//async log
	go func() {
		defer s.wg.Done()
		store := s.store
		store.Input() <- rcd
	}()
}

// for standard dns
func (self *DnsServer) responseStandard(w dns.ResponseWriter, req *dns.Msg) {
	store := self.store
	q := req.Question[0]

	origin := parseQuestionName(q.Name, self.Domain)
	// fmt.Println("query origin:", origin)

	m := new(dns.Msg)
	m.SetReply(req)

	resolveToAnswer := func(q string, r *Resolve) dns.RR {
		switch r.Type {
		case "A":
			return &dns.A{
				Hdr: dns.RR_Header{
					Name:   q,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    r.Ttl,
				},
				A: net.ParseIP(r.Value),
			}
		case "CNAME":
			return &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   q,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    r.Ttl,
				},
				Target: r.Value + ".",
			}

		case "MX":
			return &dns.MX{
				Hdr: dns.RR_Header{
					Name:   q,
					Rrtype: dns.TypeMX,
					Class:  dns.ClassINET,
					Ttl:    r.Ttl,
				},
				Preference: 0,
				Mx:         r.Value,
			}

		case "TXT":
			return &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   q,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    r.Ttl,
				},
				Txt: []string{r.Value},
			}

		default:
			return nil
		}
		return nil
	}

	findResolves := func(origin, t string) []*Resolve {
		v, exist := store.Get(origin + t)
		if exist {
			return v.([]*Resolve)
		}
		subs := strings.Split(origin, ".")
		for i := 0; i < len(subs); i++ {
			subs[i] = "*"
			altquery := strings.Join(subs[i:], ".")
			store.Get(altquery)
			if exist {
				return v.([]*Resolve)
			}
		}
		return nil
	}

	c := self.client
	cnameToAnswer := func(cname string) []dns.RR {
		v, exist := store.Get(cname + "#A")
		if exist {
			rr := v.([]*Resolve)
			var r []dns.RR
			for i := 0; i < len(rr); i++ {
				r = append(r, resolveToAnswer(cname, rr[i]))
			}
			return r
		}

		// singleflight
		{
			m := new(dns.Msg)
			m.Id = dns.Id()
			m.RecursionDesired = true
			m.Question = []dns.Question{
				dns.Question{
					Name:   cname + ".",
					Qtype:  dns.TypeA,
					Qclass: dns.ClassINET,
				},
			}

			rm, _, err := c.Exchange(m, self.Upstream)
			if err != nil {
				logrus.Errorf("[dnsserver.go::responseStandard.cnameToAnswer] client.Exchange(%v): %v", cname, err)
				return nil
			}
			if len(rm.Answer) > 0 {
				var toStore []*Resolve
				for i := 0; i < len(rm.Answer); i++ {
					a := rm.Answer[i]
					if a.Header().Rrtype == dns.TypeA {
						at := a.(*dns.A)
						toStore = append(toStore, &Resolve{
							Name:  at.Hdr.Name,
							Value: at.A.String(),
							Type:  "A",
							Ttl:   at.Hdr.Ttl,
						})
					}
				}
				//store answer to []*Resolve
				if len(toStore) > 0 {
					store.Set(cname+"#A", toStore, time.Duration(rm.Answer[0].Header().Ttl)*time.Second)
				}
			}
			return rm.Answer
		}
	}

	switch q.Qtype {
	case dns.TypeA:
		// A记录
		rr := findResolves(origin, "#A")
		if len(rr) > 0 {
			for i := 0; i < len(rr); i++ {
				m.Answer = append(m.Answer, resolveToAnswer(q.Name, rr[i]))
			}
			w.WriteMsg(m)
			return
		}

		// 查找CNAME
		rrc := findResolves(origin, "#CNAME")
		if len(rrc) > 0 {
			// CNAME 只支持1个
			a := &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    rrc[0].Ttl,
				},
				Target: rrc[0].Value + ".",
			}
			m.Answer = append(m.Answer, a)
			ca := cnameToAnswer(rrc[0].Value)
			if len(ca) > 0 {
				m.Answer = append(m.Answer, ca...)
			}
			w.WriteMsg(m)
			return
		}

	case dns.TypeCNAME:
		rr := findResolves(origin, "#CNAME")
		if len(rr) > 0 {
			for i := 0; i < len(rr); i++ {
				m.Answer = append(m.Answer, resolveToAnswer(q.Name, rr[i]))
			}
			w.WriteMsg(m)
			return
		}

	case dns.TypeTXT:
		rr := findResolves(origin, "#TXT")
		if rr == nil {
			break
		}
		for i := 0; i < len(rr); i++ {
			m.Answer = append(m.Answer, resolveToAnswer(q.Name, rr[i]))
		}
		w.WriteMsg(m)
		return

	case dns.TypeMX:
		rr := findResolves(origin, "#MX")
		if rr == nil {
			break
		}
		for i := 0; i < len(rr); i++ {
			m.Answer = append(m.Answer, resolveToAnswer(q.Name, rr[i]))
		}
		w.WriteMsg(m)
		return

	default:
	}

	dns.HandleFailed(w, req)
}

func (h *DnsServer) Do(w dns.ResponseWriter, req *dns.Msg) {
	// only handler first
	q := req.Question[0]
	store := h.store

	if q.Qclass != dns.ClassINET {
		dns.HandleFailed(w, req)
		return
	}

	//variables
	var remoteIp net.IP
	var uid int64
	var ttl uint32
	var ip net.IP
	var prefix, shortId string

	remoteAddr, ok := w.RemoteAddr().(*net.UDPAddr)
	if ok {
		remoteIp = remoteAddr.IP
	} else {
		remoteAddr, _ := w.RemoteAddr().(*net.TCPAddr)
		remoteIp = remoteAddr.IP
	}

	doResp := func(ip net.IP, t uint16) {
		m := new(dns.Msg)
		m.SetReply(req)
		rr_header := dns.RR_Header{
			Name:   q.Name,
			Rrtype: uint16(t),
			Class:  dns.ClassINET,
			Ttl:    ttl,
		}
		a := &dns.A{rr_header, ip}
		m.Answer = append(m.Answer, a)
		w.WriteMsg(m)

		if (t == dns.TypeA || t == dns.TypeAAAA) && ttl == LOG_TTL {
			h.log(&DnsRecord{
				Uid:    uid,
				Domain: strings.TrimSuffix(q.Name, "."),
				Var:    prefix,
				Ctime:  time.Now(),
				Ip:     remoteIp.String(),
			})
		}
		return
	}

	//r.u3yszl9nidbsx8p9.example.com.
	//127.0.0.1-100.100.100.100.cr.u3yszl9nidbsx8p9.example.com.
	prefix, shortId, isRebind := parseDomain(q.Name, h.Domain)
	if prefix == "" {
		ttl = DEFAULT_TTL // improve performance
	}
	doResponseError := func() {
		doResp(net.ParseIP("255.255.255.255"), q.Qtype)
	}

	doResponseCustomRebind := func(prefix string) {
		key := prefix + shortId

		// go-cache not like redis.Incr
		last, _ := store.IncrementInt(key, 1)
		if last == 0 {
			store.Set(key, 0, CUSTOM_REBIND_EXPIRE)
		}
		ips := strings.Split(prefix, "-")
		if len(ips) > 0 {
			ip, err := parseIP(ips[last%len(ips)])
			if err != nil {
				doResponseError()
				return
			}
			doResp(ip, q.Qtype)
			return
		}
		doResponseError()
	}

	//xip return custom ip
	{
		xip, err := h.parseXip(q.Name)
		if err == nil {
			ttl = XIP_TTL
			doResp(xip, q.Qtype)
			return
		}
	}

	v, exist := store.Get(shortId + ".suser")
	var user *models.TblUser
	if !exist {
		h.responseStandard(w, req)
		return
	}

	user = v.(*models.TblUser)
	uid = user.Id
	ttl = LOG_TTL
	ip = h.V4
	if isRebind && len(user.Rebind) > 0 {
		idx := time.Now().Second() % len(user.Rebind)
		ip = net.ParseIP(user.Rebind[idx])
	} else if pprefix, mode := parsePrefix(prefix); mode == "cr" {
		// custom rebind
		doResponseCustomRebind(pprefix)
		return
	}

	switch q.Qtype {
	case dns.TypeA:
		doResp(ip, q.Qtype)
		//rebinding
		return

	case dns.TypeAAAA:
		// not ipv6 now
		dns.HandleFailed(w, req)
		return

	case dns.TypeNS:
		// TODO:
		// return V4 direct
		ttl = 600
		doResp(h.V4, q.Qtype)
		return

	// for standard dns
	case dns.TypeTXT:

	case dns.TypeCNAME:

	case dns.TypeMX:

	default:
		dns.HandleFailed(w, req)
		return
	}

	dns.HandleFailed(w, req)
}

func (h *DnsServer) parseXip(qName string) (net.IP, error) {
	// 127.0.0.1.example.com
	if h.ipv4Regexp.MatchString(qName) {
		subs := h.ipv4Regexp.FindAllStringSubmatch(qName, 1)
		if len(subs) > 0 {
			ip := subs[0][1]
			return net.ParseIP(ip), nil
		}
	}
	// binary maybe match hex, try first
	if h.ipv4bRegexp.MatchString(qName) {
		// 0b1111111000000000000000000000001.example.com
		subs := h.ipv4bRegexp.FindAllStringSubmatch(qName, 1)
		if len(subs) > 0 {
			ip := subs[0][1]
			return parseBinaryIP(ip)
		}
	}
	if h.ipv4uRegexp.MatchString(qName) {
		// 0x7f000001.example.com
		// 7f000001.example.com
		fmt.Println(h.ipv4uRegexp.String(), "match", qName)
		subs := h.ipv4uRegexp.FindAllStringSubmatch(qName, 1)
		if len(subs) > 0 {
			ip := subs[0][1]
			return parseHexIP(ip)
		}
	}
	return nil, fmt.Errorf("not xip")
}
