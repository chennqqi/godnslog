# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

GODNSLOG 是一个 DNS/HTTP 日志服务器，用于安全测试中验证 SSRF/XXE/RFI/RCE 漏洞。该项目采用前后端分离架构，后端提供 DNS 和 HTTP 服务，前端提供 Web 管理界面。

## 架构组织

### 核心组件

- **server/dnsserver.go**: DNS 服务器实现，使用 miekg/dns 库处理 DNS 查询
- **server/webserver.go**: Web 服务器和 API 端点，基于 Gin 框架
- **server/webapi.go**: Web API 处理逻辑，包括域名解析和用户认证
- **server/webui.go**: Web UI 界面渲染
- **cache/cache.go**: 内存缓存系统，用于存储会话和临时数据
- **models/table.go**: 数据库模型定义
- **models/api.go**: 数据库操作 API

### 技术栈

- **后端**: Go 1.14+，使用 Gin Web 框架，XORM ORM，支持 SQLite3/MySQL
- **前端**: Vue.js 2.x + Ant Design Vue
- **DNS**: miekg/dns 库
- **命令行**: google/subcommands 框架

### 工作流程

1. 用户注册后获得唯一域名: `userXXXX.example.com`
2. DNS 查询该域名生成 DNSLOG 记录
3. HTTP 请求该域名生成 HTTPLOG 记录
4. 支持 DNS rebinding 攻击测试

## 构建和运行命令

### 前端开发

```bash
cd frontend
yarn install        # 安装依赖
yarn serve          # 开发服务器
yarn build          # 生产构建
yarn lint           # 代码检查
```

### 后端开发

```bash
go build            # 构建可执行文件
go test             # 运行测试
```

### Docker 构建

```bash
docker build -t "user/godnslog" .              # 标准构建
docker build -t "user/godnslog" -f DockerfileCN .  # 中国用户构建
```

### 运行服务

```bash
# 直接运行
./godnslog serve -domain yourdomain.com -4 your.public.ip

# Docker 运行
docker run -p 80:8080 -p 53:53/udp "user/godnslog" serve -domain yourdomain.com -4 your.public.ip
```

## 开发注意事项

### 命令行结构

项目使用 `google/subcommands` 框架，主要子命令:
- `serve`: 启动 DNS 和 HTTP 服务
- `resetpw`: 重置管理员密码

### 数据库配置

- 默认使用 SQLite3: `file:godnslog.db?cache=shared&mode=rwc`
- 支持 MySQL: 通过 `-driver mysql -dsn "user:pass@tcp(host:port)/dbname"`
- 数据库表结构定义在 `models/table.go`

### DNS 配置要求

1. 注册域名并设置 NS 记录指向服务器 IP
2. 服务器需要开放 UDP 53 端口和 TCP 80/8080 端口
3. 支持上游 DNS 代理 (默认 8.8.8.8:53)

### 缓存系统

- 使用内存缓存存储用户会话和 DNS 记录
- 缓存过期时间默认 24 小时
- 清理间隔默认 10 分钟

### 多语言支持

- 支持英文 (en-US) 和中文 (zh-CN)
- 通过 `-lang` 参数设置默认语言
- 翻译文件在 `server/translate.go`

### API 认证

- 使用 Token 认证机制
- 管理员用户: `admin` (首次运行密码在控制台显示)
- 支持 JWT token (github.com/dgrijalva/jwt-go)

### 测试

- Go 测试文件: `server/utils_test.go`
- 前端测试: `yarn test:unit`

### 关键配置常量

- `AuthExpire`: 24小时认证过期
- `DefaultCleanInterval`: 7200秒清理间隔  
- `DefaultQueryApiMaxItem`: API 查询最大 20 条记录
- `DefaultMaxCallbackErrorCount`: 最大回调错误次数 5 次