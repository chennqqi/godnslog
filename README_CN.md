# GODNSLOG

![](https://s1.ax1x.com/2020/08/31/dXFLg1.png)

A dns&amp;http log server for verify SSRF/XXE/RFI/RCE vulnerability 

[English Doc](https://github.com/chennqqi/godnslog) | 中文文档

## 功能特性

- DNSLOG
- HTTPLOG
- Rebinding
- Push (callback)
- Multi-user
- 支持Docker一键运行
- python/golang客户端代码


### DNSLOG
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

查看最新版本
https://hub.docker.com/r/sort/godnslog/tags

```bash
docker pull "sort/godnslog:version-0.3.0"
docker run -p80:8080 -p53:53/udp "sort/godnslog:version-0.3.0" -domain yourdomain.com -4 100.100.100.100
```

version-0.3.0 替换为最新版本号
yourdomain.com 替换为你的域名
100.100.100.100 替换为你的公网IP


## 已知问题

- introduce/文档的mavon-editor会遮挡下拉菜单
- 修改语言不能保存到后端
- 一些没必要的调试打印问题
- 口令失效后的一些重复提示

## 关注我们

![](https://open.weixin.qq.com/qr/code?username=gh_d110440c4890)