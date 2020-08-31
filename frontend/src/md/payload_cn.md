# Payloads

详细的payloads欢迎大家提PR补充。
文档所在目录 `$SRC/frontend/src/md/payload_cn.md`

## DNSLOG

在公网部署时要求目标本地DNS支持将解析递归到公网上，在内网部署时需要本地DNS配置一个扫描专用的私有域名。相对于HTTPLOG，DNSLOG依赖少，payload也短一些。某些漏洞只能通过DNSLOG来验证，例如一些命令执行的场景。

## HTTPLOG

在一些网络环境下，如果本地DNS限制了公网解析，那么公网DNSLOG就无法进行验证。此时有两种方案，一是使用内网DNSLOG，需要本地DNS修改配置，指定一个私有域名；另一种方案是使用HTTPLOG。
HTTPLOG配合callback代码会简单协议，缺点是不能发现只能通过DNS验证的漏洞。

## 其他

1. phpRFI PHP文件远程包含

```
curl http://${IP}/payload/phprfi
<?php  echo md5("GODNSLOG");
```

如果漏洞存在，在漏洞所在页面会输出694ef536e5d0245f203a1bcf8cbf3294
	

## 参考资料

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