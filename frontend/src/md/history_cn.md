# 更新日志

## v0.7.0

xip
自定义rebinding

## v0.5.0

新特性：
1. 可以当成一个普通DNS服务器使用，支持A,TXT,CNAME,MX四种类型的记录，不支持泛解析
2. 完善前端和后端的翻译

## v0.4.0

新特性：
1. 支持[xip](http://xip.io/)功能
    10.0.0.1.godnslog.com   resolves to   10.0.0.1
    www.10.0.0.1.godnslog.com   resolves to   10.0.0.1
    mysite.10.0.0.1.godnslog.com   resolves to   10.0.0.1
    foo.bar.10.0.0.1.godnslog.com   resolves to   10.0.0.1
2. 启动程序使用子命令
3. 添加重置用户的子命令

## v0.3.0

添加客户端sdk
使用dockerhub自动化构建

## v0.2.0

修复v0.1.0BUG，使用docker部署运行

## v0.1.0

实现基本功能