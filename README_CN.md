# GODNSLOG

![](https://s1.ax1x.com/2020/08/31/dXFLg1.png)

A dns&amp;http log server for verify SSRF/XXE/RFI/RCE vulnerability 

[Engglish]()|中文

## 功能特性

- DNSLOG
- HTTPLGO
- Rebinding
- Push (callback)
- Multi-user
- 支持Docker一键运行
- python/golang客户端代码

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

## 关注我们

![](https://open.weixin.qq.com/qr/code?username=gh_d110440c4890)