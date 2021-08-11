# GODNSLOG

![](https://z3.ax1x.com/2021/08/10/fGd4IJ.png)

A dns&amp;http log server for verify SSRF/XXE/RFI/RCE vulnerability 

[English Doc](https://github.com/chennqqi/godnslog) | 中文文档

## 功能特性

- 标准DNS解析服务
- DNSLOG
- HTTPLOG
- Rebinding/CustomRebinding
- Push (callback)
- Multi-user
- 支持Docker一键运行
- python/golang客户端代码
- 标准DNS解析功能，支持A,CNAME,MX,TXT四种类型
- xip

### DNSLOG

默认账户: `admin`
首次运行程序时会创建随机密码，可以在控制台打印获取到该密码。
也可以使用`resetpw`子命令来自定义密码

![](https://s1.ax1x.com/2020/08/31/dXPba4.png)


### HTTPLOG
![](https://s1.ax1x.com/2020/08/31/dXiiIH.png)

## 编译前端

依赖: 

`yarn`

```
cd frontend
yarn install
yarn build
```
	
## 编译后端

依赖: 

`golang >= 1.13.0`

```bash
go build
```

## docker build

```bash
docker build -t "user/godnslog" .
```

国内用户使用下面的Dockefile:

```bash
docker build -t "user/godnslog" -f DockerfileCN .
```

## docker一键运行

### I 修改域名的DNS服务器地址

以阿里云为例

1. 在域名-管理页面, 自定义DNS Host，定义域名的NS地址, ns1.xxx.xxx, ns2.xxx.xxx
如果你的服务器IP有别的域名解析，可以跳过这一步，这里要配置两个不同的IP地址，你可以将其中一个临时指向一个假地址，例如100.100.100.100![](https://s1.ax1x.com/2020/09/04/wFiaM8.png)


2. DNS修改，将域名的DNS服务器修改为上一步定义的DNS地址，这一步中限制只能填入域名。
这里有个限制最少填入两个地址，而且两个地址不能是同一个IP。 
![](https://s1.ax1x.com/2020/09/04/wFitRP.png)
![](https://s1.ax1x.com/2020/09/04/wFiJPI.png)

3. 最后回到第一步中将假地址IP也修改为真实的IP地址

4. whois验证修改是否生效
<https://lookup.icann.org/lookup>
![](https://s1.ax1x.com/2020/09/04/wFk04s.png)

如果使用腾讯云(DNSPOD), 需要使用第二个域名的NS记录(如ns1.seconddomain.com,ns2.seconddomain)指向100.100.100.100并修改域名解析服务器为第二个域名的NS地址(如ns1.seconddomain.com,ns2.seconddomain).

### II. 查看获取最新版本
https://hub.docker.com/r/sort/godnslog/tags

### III.拉取并运行

```bash
docker pull "sort/godnslog"
docker run -p80:8080 -p53:53/udp "sort/godnslog" serve -domain yourdomain.com -4 100.100.100.100
```

yourdomain.com 替换为你的域名
100.100.100.100 替换为你的公网IP


## TODO && 已知问题

- [ ] ~~增强反向代理场景~~
- [ ] 管理员查看所有用户
- [ ] 允许未登录用户访问文档
- [ ] 自定义Rebind第二阶段TTL可配置
- [ ] 前端页面口令失效后的一些重复提示
- [ ] markdown组件使用了cloudflare提供的js CDN(可以先用ReplaceCDN临时解决)，如果cf被墙会导致网页一直卡住

## 文档

详细信息可以阅读自述文档

guest/guest123

[introduce](https://www.godnslog.com/document/introduce)
[payload](https://www.godnslog.com/document/payload)
[api](https://www.godnslog.com/document/api)
[rebiding](https://www.godnslog.com/document/rebinding)
[resolve](https://www.godnslog.com/document/resolve)

## 关注我

![](https://open.weixin.qq.com/qr/code?username=gh_4a48daaf398b)