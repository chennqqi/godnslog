# Payloads

PRs are welcome!
This Document located in `$SRC/frontend/src/md/payload_cn.md`

## DNSLOG

TODO:

## HTTPLOG

TODO:

## Others

1. phpRFI PHP

```
curl http://${IP}/payload/phprfi
<?php  echo md5("GODNSLOG");
```

RETURNS:

694ef536e5d0245f203a1bcf8cbf3294
	

## Reference

收集一些不错的文章：

**<http://ceye.io/payloads>**
<https://www.cnblogs.com/rnss/p/11320305.html>
<http://foreversong.cn/archives/861>
<https://bbs.ichunqiu.com/thread-22002-1-1.html>
<http://blog.knownsec.com/2016/06/how-to-scan-and-check-vulnerabilities/>
<https://www.freebuf.com/column/184587.html>
<https://blog.51cto.com/tar0cissp/2337779>
<https://github.com/BugScanTeam/DNSLog>

推荐仔细读一读http://ceye.io/payloads