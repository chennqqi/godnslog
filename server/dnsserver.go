package server

import (
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/models"
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

TODO:
1. 支持IPv6

*/

const (
	LOG_TTL     = 0
	NS_TTL      = 600
	DEFAULT_TTL = 300
	XIP_TTL     = 86400
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
	Fixed              []Resolve
}

type DnsServer struct {
	DnsServerConfig
	store *cache.Cache

	tcpServer  *dns.Server
	udpServer  *dns.Server
	ipv4Regexp *regexp.Regexp

	wg      sync.WaitGroup
	handler dns.Handler

	fixed map[string][]Resolve
}

func NewDnsServer(cfg *DnsServerConfig, store *cache.Cache) (*DnsServer, error) {
	domain := cfg.Domain
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	ipv4Exp := `((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))`
	ipv4Exp = ipv4Exp + strings.Replace("."+domain, ".", `\.`, -1)

	addr := ""
	fixed := make(map[string][]Resolve)
	for i := 0; i < len(cfg.Fixed); i++ {
		r := cfg.Fixed[i]
		v, exist := fixed[r.Name]
		if exist {
			v = append(v, cfg.Fixed[i])
			fixed[r.Name] = v
		} else {
			fixed[r.Name] = []Resolve{r}
		}
	}

	handler := dns.NewServeMux()
	var s = &DnsServer{
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
		fixed: fixed,
	}
	s.ipv4Regexp = regexp.MustCompile(ipv4Exp)
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

func (h *DnsServer) Do(w dns.ResponseWriter, req *dns.Msg) {
	// only handler first
	q := req.Question[0]
	store := h.store
	fixed := h.fixed

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
	prefix, shortId, isRebind := parseDomain(q.Name, h.Domain)
	if prefix == "" {
		ttl = DEFAULT_TTL // improve performance
	}

	//xip return custom ip
	{
		subs := h.ipv4Regexp.FindAllStringSubmatch(q.Name, 1)
		if len(subs) > 0 {
			ip := subs[0][1]
			ttl = XIP_TTL
			doResp(net.ParseIP(ip), q.Qtype)
			return
		}
	}

	v, exist := store.Get(shortId + ".suser")
	var user *models.TblUser
	if exist {
		user = v.(*models.TblUser)
		uid = user.Id
		ttl = LOG_TTL
		ip = h.V4
		if isRebind && len(user.Rebind) > 0 {
			idx := time.Now().Second() % len(user.Rebind)
			ip = net.ParseIP(user.Rebind[idx])
		}
	} else {
		rrs, exist := fixed[shortId]
		if exist {
			idx := time.Now().Second() % len(rrs)
			r := &rrs[idx]
			ip = net.ParseIP(r.Value)
			ttl = r.Ttl
		} else {
			ip = h.V4
			ttl = DEFAULT_TTL
		}
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

	default:
		dns.HandleFailed(w, req)
		return
	}

	dns.HandleFailed(w, req)
}

func (self *DnsServer) Update(rr []Resolve) {
	fixed := make(map[string][]Resolve)
	for i := 0; i < len(rr); i++ {
		r := rr[i]
		v, exist := fixed[r.Name]
		if exist {
			v = append(v, rr[i])
			fixed[r.Name] = v
		} else {
			fixed[r.Name] = []Resolve{r}
		}
	}
	self.fixed = fixed
}
