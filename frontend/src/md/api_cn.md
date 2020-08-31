# API

`GODNSLOG` 支持
源代码目录

源代码目录的`client`目录提供了SDK供客户端调用，`examples`目录中提供了样例代码。
API将会支持精准查询和模糊匹配两种模式。

## 参数说明

`domain`: 用户唯一域名
`secret`: 用户私钥令牌，在配置->系统->安全设置中可以获取，唯一，请勿泄露

![](https://s1.ax1x.com/2020/08/31/dOzKZd.png)

`q`: 查询变量，对于DNS来说，自定义的DNS前缀，对于HTTP来说，查询的是自定义的URL后缀
`blur`: 模糊查询选项，0表示精确查询，1表示模糊查询
`t`: 当前Unix时间戳，t为客户端当前时间，当t时间与服务端接收到请求时t1超过1分钟，返回认证失败。请确保客户端时间和标准时间是同步的。
`hash`: query参数中所有参数按照字典顺序进行排序，依次取值，最后和token一起计算md5

下面是使用golang演示计算query的hash过程：

```Go
var keys []string
for k, _ := range querys {
	keys = append(keys, k)
}

# 排序
sort.Strings(keys)

h := md5.New()
for _, k := range keys {
	h.Write([]byte(querys.Get(k)))
}
h.Write([]byte(secret))
hash := hex.EncodeToString(h.Sum(nil))
```

## 查询DNS记录

1.请求
```
curl bkkpdcy7lo84.godnslog.com/data/dns?q=${domain}&t=${timestamp}&hash={hash}&blur=0

```
2.返回值

- 200 成功
- 401 认证失败
- 400 参数错误
- 502 服务器错误

## 查询HTTP记录

```
curl bkkpdcy7lo84.godnslog.com/data/dns?q=${domain}&t=${timestamp}&hash={hash}&blur=0

```

2.返回值

- 200 成功
- 401 认证失败
- 400 参数错误
- 502 服务器错误


## 示例代码

Golang 客户端

```Go 
package main

import (
	"fmt"
	"log"

	"github.com/chennqqi/godnslog/client"
)

func main() {
	var (
		domain = "bkkpdcy7lo84.godnslog.com"
		secret = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	)
	c, err := client.NewClient(domain, secret, false)
	if err != nil {
		log.Println("NewClient: %v", err)
		return
	}
	{
		r, err := c.QueryDns(query, blur)
		if err != nil {
			log.Fatal("query DNS", err)
		}
	}
	{
		r, err := c.QueryHttp(query, blur)
		if err != nil {
			log.Fatal("query HTTP", err)
		}
		fmt.Printf("Result: %#v\n", r)
	}

```

python 客户端

```python
class Client(object):
    def __init__(self, domain, secret, ssl):
        self.domain = domain
        self.secret = secret
        self.host = 'http://' + domain
        if ssl:
            self.host = 'https://' + domain

    @staticmethod
    def hash(query):
        m = hashlib.new('md5')
        keys = list(query.keys())
        keys.sort()
        for key in keys:
            m.update(query[key].encode(encoding='UTF-8'))
        m.update(secret.encode(encoding='UTF-8'))
        return m.hexdigest()

    def query_dns(self, q, blur=False):
        payload = {}
        payload['t'] = str(int(time.time()))
        print('t=', int(time.time()), payload['t'])
        payload['q'] = q
        if blur:
            payload['blur'] = '1'
        else:
            payload['blur'] = '0'
        payload['hash'] = Client.hash(payload)

        r = requests.get(self.host + '/data/dns', payload)
        print('resp:', r.text)
        j = json.loads(r.text)
        if r.status_code == 200:
            return j['result'], j['message']
        else:
            return None, j['message']

    def query_http(self, q, blur):
        payload = {}
        payload['t'] = ''.format('{}', int(time.time()))
        payload['q'] = q
        if blur:
            payload['blur'] = '1'
        else:
            payload['blur'] = '0'
        payload['hash'] = Client.hash(payload)

        r = requests.get(self.host + '/data/http', payload)
        print('resp:', r.text)
        j = json.loads(r.text)
        if r.status_code == 200:
            return j['result'], j['message']
        else:
            return None, j['message']
```