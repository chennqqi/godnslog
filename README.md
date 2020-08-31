# GODNSLOG

![](https://s1.ax1x.com/2020/08/31/dXFLg1.png)

A dns&amp;http log server for verify SSRF/XXE/RFI/RCE vulnerability

English Doc | [中文文档](https://github.com/chennqqi/godnslog/blob/master/README_CN.md)

## features

- DNSLOG
- HTTPLGO
- Rebinding
- Push (callback)
- Multi-user
- dockerlized
- python/golang client sdk


### DNSLOG
![](https://s1.ax1x.com/2020/08/31/dXPba4.png)


### HTTPLOG
![](https://s1.ax1x.com/2020/08/31/dXiiIH.png)


## build frontend

requirements: 

`yarn`

```
cd frontend
yarn install
yarn build
```
	
## build backend

requirements: 

`golang >= 1.13.0`

```bash
go build
```

## docker build

```bash
docker build -t "user/godnslog" .
```

For Chinese user:

```bash
docker build -t "user/godnslog" -f DockerfileCN .
```

## Follow us

wechat:
![](https://open.weixin.qq.com/qr/code?username=gh_d110440c4890)