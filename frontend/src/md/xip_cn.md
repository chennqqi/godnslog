# xip

xip.io是简单好用的免费泛域名服务，它可以实现以下功能。可以通过自定义IP域名实现三级域名的泛解析，在辅助测试、调试，内网探测时可能用到的一个工具，目前xip.io服务已不可用。

``` bash
          10.0.0.1.xip.io   resolves to   10.0.0.1
      www.10.0.0.1.xip.io   resolves to   10.0.0.1
   mysite.10.0.0.1.xip.io   resolves to   10.0.0.1
  foo.bar.10.0.0.1.xip.io   resolves to   10.0.0.1
```
godnslog实现了类似于xip.io的功能

``` bash
          10.0.0.1.godnslog.com   resolves to   10.0.0.1
      www.10.0.0.1.godnslog.com  resolves to   10.0.0.1
   mysite.10.0.0.1.godnslog.com   resolves to   10.0.0.1
  foo.bar.10.0.0.1.godnslog.com   resolves to   10.0.0.1
```
TTL 默认为86400，不可配置，可以通过修改源代码`server/dnsserver.go`中的`XIP_TTL`配置

## 其他类似的服务

- https://nip.io/
- https://sslip.io/
- https://github.com/peterhellberg/xip.name
- https://github.com/basecamp/xip-pdns
