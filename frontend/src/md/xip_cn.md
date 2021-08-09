# xip

xip.io是简单好用的免费泛域名服务，它可以实现以下功能。可以通过自定义IP域名实现三级域名的泛解析，在辅助测试、调试，内网探测时可能用到的一个工具，目前xip.io服务已不可用。

``` bash
          10.0.0.1.xip.io   resolves to   10.0.0.1
      www.10.0.0.1.xip.io   resolves to   10.0.0.1
   mysite.10.0.0.1.xip.io   resolves to   10.0.0.1
  foo.bar.10.0.0.1.xip.io   resolves to   10.0.0.1
```
godnslog实现了类似于xip.io的功能, 还支持十六进制和二进制格式

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
TTL 默认为86400，不可配置，可以通过修改源代码`server/dnsserver.go`中的`XIP_TTL`配置

## 其他类似的服务

- https://nip.io/
- https://sslip.io/
- https://github.com/peterhellberg/xip.name
- https://github.com/basecamp/xip-pdns
