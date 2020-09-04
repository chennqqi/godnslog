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

default account: `admin/password`

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

## RUN

i. Register your domain, eg: `example.com`
Set your DNS Server point to your host, eg: ns.example.com => 100.100.100.100
Some registrar limit set to NS host, your can set two ns host point to only one address.
Some registrar to ns host must be different ip address, you can set one to a fake addresss and then change to the same addresss


ii. self build

```bash
docker run -p80:8080 -p53:53/udp "user/godnslog"  -domain yourdomain.com -4 100.100.100.100
```

or use dockerhub

```bash
docker pull "sort/godnslog:version-0.3.0"
docker run -p80:8080 -p53:53/udp "sort/godnslog:version-0.3.0" -domain yourdomain.com -4 100.100.100.100
```

## Follow us


![](https://open.weixin.qq.com/qr/code?username=gh_d110440c4890)