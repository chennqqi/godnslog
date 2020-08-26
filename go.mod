module github.com/chennqqi/godnslog

go 1.14

replace (
	docs => ./main/docs
	models => ./models
)

require (
	docs v0.0.0-00010101000000-000000000000
	github.com/BurntSushi/toml v0.3.1
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/appleboy/gin-jwt/v2 v2.6.4
	github.com/buger/jsonparser v1.0.0 // indirect
	github.com/chennqqi/crashreport v1.0.3 // indirect
	github.com/chennqqi/goutils v0.1.5
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eapache/channels v1.1.0 // indirect
	github.com/gin-contrib/cors v1.3.1 // indirect
	github.com/gin-contrib/pprof v1.3.0 // indirect
	github.com/gin-contrib/static v0.0.0-20200815103939-31fb0c56a3d1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-xorm/xorm v0.7.9 // indirect
	github.com/grt1st/dnsgo v0.0.0-20180722082459-70705c96dd1c
	github.com/hashicorp/consul/api v1.6.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.7
	github.com/kaeuferportal/stack2struct v0.0.0-20150211121835-6838fbadc918 // indirect
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/miekg/dns v1.1.31
	github.com/muesli/crunchy v0.4.0 // indirect
	github.com/nsqio/go-nsq v1.0.8 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sethvargo/go-password v0.2.0 // indirect
	github.com/shirou/gopsutil v2.20.7+incompatible // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/sluu99/uuid v0.0.0-20130306162636-1dd34a9ad6c0 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.2.0
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	gopkg.in/olivere/elastic.v5 v5.0.86 // indirect
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5 // indirect
	models v0.0.0-00010101000000-000000000000
	xorm.io/xorm v1.0.3
)
