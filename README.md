# GODNSLOG

![](https://z3.ax1x.com/2021/08/10/fGd4IJ.png)

A dns&amp;http log server for verify SSRF/XXE/RFI/RCE vulnerability

English Doc | [中文文档](https://github.com/chennqqi/godnslog/blob/master/README_CN.md)

## features

- Standard Domain Resolve Service
- DNSLOG
- HTTPLOG
- Rebinding/CustomRebinding
- Push (callback)
- Multi-user
- dockerlized
- python/golang client sdk
- as a standard name resolve service with support `A,CNAME,TXT,MX`
- xip


### DNSLOG

super admin user: `admin`
password will be showed in console logs when first run.
you can change it by subcommand `resetpw`

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
docker run -p80:8080 -p53:53/udp "user/godnslog"  serve -domain yourdomain.com -4 100.100.100.100
```

or use dockerhub

```bash
docker pull "sort/godnslog"
docker run -p80:8080 -p53:53/udp -p80:8080  "sort/godnslog" serve -domain yourdomain.com -4 100.100.100.100
```

iii. access http://100.100.100.100

## Doc

guest/guest123

[introduce](https://www.godnslog.com/document/introduce)
[payload](https://www.godnslog.com/document/payload)
[api](https://www.godnslog.com/document/api)
[rebiding](https://www.godnslog.com/document/rebinding)
[resolve](https://www.godnslog.com/document/resolve)

## TODO && Known Issues

- [x]fix demo code
- [x]add docker-compose
- [x]add default www resolve
- [x]init guest user option
- [x]fix custom clean interval setting
- [x]enhance rebinding https://github.com/chennqqi/godnslog/issues/14
- [ ]~~enhance reverse proxy~~
- [ ]admin user can read all recordds
- [ ]allow Anonymous user access document page
- [ ]enable custom rebinding stage two setting
- [ ]fix logout dialog overlap by markdown 
- [ ]fix login logical problem

## Follow me

![](https://open.weixin.qq.com/qr/code?username=gh_4a48daaf398b)