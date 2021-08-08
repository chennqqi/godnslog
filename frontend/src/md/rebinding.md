# Rebinding

## 1. Custom Rebinding

Since 0.7.0, GODNSLOG enable custom rebinding by request

```bash
	dig 127.0.0.1-100.100.100.100.cr.godnslog.com

	#第一次返回:
	127.0.0.1-10.10.10.10.cr.1frwakkp3j14.godnslog.com. 0 IN A 127.0.0.1

	#第二次返回:
	127.0.0.1-10.10.10.10.cr.1frwakkp3j14.godnslog.com. 0 IN A 10.10.10.10
```


## 2. Global Rebinding

Like <http://ceye.io>, random policy。 

`r.${shortId}.godnslog.com` and it's subdomain as rebind domain. Each request return a random response which you have set.

### Configure
![](https://s1.ax1x.com/2020/08/31/dOO6fO.png)

### Usage

![](https://s1.ax1x.com/2020/08/31/dOOgpD.png)

```bash
	dig r.godnslog.com

	#FIRST ANSWER SECTION:
	r.1frwakkp3j14.godnslog.com. 0 IN A 127.0.0.1

	#SECOND ANSWER SECTION:
	r.1frwakkp3j14.godnslog.com. 0 IN A 10.10.10.10
```
