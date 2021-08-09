# Rebinding

## 1. 自定义Rebinding

从v0.7.x起，新增了自定义rebinding功能

```bash
	dig 127.0.0.1-100.100.100.100.cr.godnslog.com

	#第一次返回:
	127.0.0.1-10.10.10.10.cr.1frwakkp3j14.godnslog.com. 0 IN A 127.0.0.1

	#第二次返回:
	127.0.0.1-10.10.10.10.cr.1frwakkp3j14.godnslog.com. 0 IN A 10.10.10.10
```
	
## 2. 全局配置Rebinding

目前Rebinding的实现跟ceye.io一样是随机策略，多个配置随机出现。

`r.${shortId}.godnslog.com`及其子域名作为Rebinding域名, 每次请求随机返回一个配置的结果

```bash
	dig r.godnslog.com

	#FIRST ANSWER SECTION:
	r.1frwakkp3j14.godnslog.com. 0 IN A 127.0.0.1

	#SECOND ANSWER SECTION:
	r.1frwakkp3j14.godnslog.com. 0 IN A 10.10.10.10
```

配置
![](https://s1.ax1x.com/2020/08/31/dOO6fO.png)


## 使用

![](https://s1.ax1x.com/2020/08/31/dOOgpD.png)
