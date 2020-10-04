# API

`GODNSLOG` client api

## Parameters description

`domain`: user private domain
`secret`: user private secret to access client api

![](https://s1.ax1x.com/2020/08/31/dOzKZd.png)

`q`: query parameter, domain's prefix for dns query or http path for http query 
`blur`: 0, exactly match; wildcard match
`t`: caller's current unix timestamp
`hash`: hash of all parameters sorted by dict order

hash algorithm：

```Go
var keys []string
for k, _ := range querys {
	keys = append(keys, k)
}

# sort
sort.Strings(keys)

h := md5.New()
for _, k := range keys {
	h.Write([]byte(querys.Get(k)))
}
h.Write([]byte(secret))
hash := hex.EncodeToString(h.Sum(nil))
```

## How to query dns records?

1.Request
```
curl bkkpdcy7lo84.godnslog.com/data/dns?q=${domain}&t=${timestamp}&hash={hash}&blur=0

```
2.Response code

- 200 OK
- 401 Authorized failed
- 400 Parameters error
- 502 Service error

## How to query http records?

```
curl bkkpdcy7lo84.godnslog.com/data/dns?q=${domain}&t=${timestamp}&hash={hash}&blur=0

```

2.Response code

- 200 OK
- 401 Authorized failed
- 400 Parameters error
- 502 Service error


## Sample codes

Golang Client

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

python client

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