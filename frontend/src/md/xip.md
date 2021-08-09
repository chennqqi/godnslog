# xip

xip.io is a magic domain name that provides wildcard DNS
for any IP address. Say your LAN IP address is 10.0.0.1.
Using xip.io,

``` bash
          10.0.0.1.xip.io   resolves to   10.0.0.1
      www.10.0.0.1.xip.io   resolves to   10.0.0.1
   mysite.10.0.0.1.xip.io   resolves to   10.0.0.1
  foo.bar.10.0.0.1.xip.io   resolves to   10.0.0.1
```

...and so on. You can use these domains to access virtual
hosts on your development web server from devices on your
local network, like iPads, iPhones, and other computers.
No configuration required!

Xip.io is now unavailable! godnslog implement xip.io's feature.


``` bash
          10.0.0.1.godnslog.com   resolves to   10.0.0.1
      www.10.0.0.1.godnslog.com   resolves to   10.0.0.1
   mysite.10.0.0.1.godnslog.com   resolves to   10.0.0.1
  foo.bar.10.0.0.1.godnslog.com   resolves to   10.0.0.1

          7f000001.godnslog.com   resolves to   127.0.0.1
      www.7f000001.godnslog.com   resolves to   127.0.0.1
        0x7f000001.godnslog.com   resolves to   127.0.0.1
foo.bar.0x7f000001.godnslog.com   resolves to   127.0.0.1

	    01111111000000000000000000000001.godnslog.com resolves to 127.0.0.1
	  0b01111111000000000000000000000001.godnslog.com resolves to 127.0.0.1
	www.01111111000000000000000000000001.godnslog.com resolves to 127.0.0.1
foo.bar.01111111000000000000000000000001.godnslog.com resolves to 127.0.0.1
```
By default TTL is 86400s，Modify `XIP_TTL` in `server/dnsserver.go` to change this value.

## Alternative service 

- https://nip.io/
- https://sslip.io/
- https://github.com/peterhellberg/xip.name
- https://github.com/basecamp/xip-pdns
