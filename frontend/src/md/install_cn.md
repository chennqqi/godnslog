# 安装说明

## docker安装部署

### I.必要条件

公网环境部署需求：

- 一个公网域名
- 一个用用独立IPv4地址的VPS
- docker

内网环境部署：

- 内网DNS自定义的内网域名
- 内网主机
- docker 

本文只介绍公网环境部署，内网环境可以参照配置。 

### II.安装软件环境
```
yum install docker -y
```

### II.修改域名的DNS服务器地址

使用独立主机IP作为域名的DNS服务器，以阿里云为例
	
1. 在域名-管理页面, 自定义DNS Host，定义域名的NS地址, ns1.xxx.xxx, ns2.xxx.xxx
如果你的服务器IP有别的域名解析，可以跳过这一步，这里要配置两个不同的IP地址，你可以将其中一个临时指向一个假地址，例如100.100.100.100![](https://s1.ax1x.com/2020/09/04/wFiaM8.png)
2. DNS修改，将域名的DNS服务器修改为上一步定义的DNS地址，这一步中限制只能填入域名。
这里有个限制最少填入两个地址，而且两个地址不能是同一个IP。 
![](https://s1.ax1x.com/2020/09/04/wFitRP.png)
![](https://s1.ax1x.com/2020/09/04/wFiJPI.png)
	
3 最后回到第一步中将假地址IP也修改为真实的IP地址
	
4 whois验证修改是否生效
<https://lookup.icann.org/lookup>
![](https://s1.ax1x.com/2020/09/04/wFk04s.png)


### III.查看并获取最新版docker镜像并运行

查看新版本: <https://hub.docker.com/r/sort/godnslog/tags>

拉取新版本: `docker pull sort/godnslog:v0.4.0`


```bash
docker pull "sort/godnslog:version-0.3.0"
docker run -p80:8080 -p53:53/udp "sort/godnslog:version-0.4.0" serve -domain yourdomain.com -4 100.100.100.100
```

version-0.3.0 替换为最新版本号
yourdomain.com 替换为你的域名
100.100.100.100 替换为你的公网IP

如果想执行编译镜像，请参考根目录下的`Dockerfile`

首次运行后会在控制台打印超级用户密码，使用`docker logs  ${containterId}`获取


## 源代码直接编译

### I.必要条件

- yarn
- go(https://golang.google.cn/dl/)
- gcc
- 

### II. 编译代码 

编译前端
```
cd frontend
yarn run build
```

编译后端
```
go build
```
