# API

源代码目录的`client`目录提供了SDK供客户端调用，`examples`目录中提供了样例代码。
API将会支持精准查询和模糊匹配两种模式。

## 参数说明

`domain`: 用户唯一域名
`secret`: 用户私钥令牌，在后台安全设置中可以获取，唯一，请勿泄露


`blur`: 模糊查询选项，0表示精确查询，1表示模糊查询
`q`: 查询变量，对于DNS来说，自定义的DNS前缀，对于HTTP来说，查询的是自定义的URL后缀
`t`: 当前Unix时间戳，t为客户端当前时间，当t时间与服务端接收到请求时t1超过1分钟，返回认证失败。请确保客户端时间和标准时间是同步的。
`hash`: query参数中所有参数按照字典顺序进行排序，依次取值，最后和token一起计算md5

下面是使用golang演示计算query的hash过程：

```golang
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
h.Write([]byte(Token))
hash := hex.EncodeToString(h.Sum(nil))
```

## 查询DNS记录

1.请求
```
curl bkkpdcy7lo84.godnslog.com/data/dns?q={domain}&t={timestamp}&hash={hash}

```
2.返回值



## 查询HTTP记录

```
curl bkkpdcy7lo84.godnslog.com/data/dns?q={domain}&t={timestamp}&hash={hash}

```
