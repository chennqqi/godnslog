# 参考项目功能分析报告

> 对 doc/references.md 中列出的开源项目进行功能、能力和设计分析，评估其对 godnslog 的借鉴价值。

---

## 1. [adysec/DNSLog](https://github.com/adysec/DNSLog)

| 维度 | 内容 |
|------|------|
| 语言 | Rust |
| 技术栈 | trust-dns-server + actix-web + SQLite |
| Stars | 117 |

### 功能与架构

基于 Rust 的轻量级 DNSLog 平台，集成了 DNS 服务和 Web 仪表盘。支持自动注册用户、生成唯一子域名以及实时展示 DNS 日志。

### 对比 godnslog

- godnslog 在功能维度上已全面超越该项目（多协议支持、HA、Workflow 等）
- Rust 实现的单二进制部署理念与 godnslog 一致

### 值得借鉴

| 特性 | 说明 | 优先级 |
|------|------|--------|
| ⭐ 无外部依赖设计 | Rust 编译为单二进制，零外部依赖启动。godnslog 同样可以做到但默认未强调 | 低 |
| Web 仪表盘实时展示 | actix-web 的 SSE 方式实现实时更新 | 低 |

### 综合评价

该项目功能基础，godnslog 在架构和功能上已全面超越，参考价值有限。

---

## 2. [yumusb/DNSLog-Platform-Golang](https://github.com/yumusb/DNSLog-Platform-Golang)

| 维度 | 内容 |
|------|------|
| 语言 | Go |
| 技术栈 | TOML 配置 + Go 原生 |
| Stars | ~200+ |

### 功能与架构

极简的 Go DNSLog 平台，通过 `config.toml` 驱动。核心特性：Token 隐私机制、HTTP Basic Auth、多域名支持。

### 对比 godnslog

godnslog 已完全覆盖其功能，且在架构复杂度和功能丰富度上远超。

### 值得借鉴

| 特性 | 说明 | 优先级 |
|------|------|--------|
| ⭐ 多域名支持 | 配置中可指定多个域名，查询时可选 | 低 |
| ⭐ Basic Auth 保护 | 简单有效的管理接口保护 | 低（godnslog 已有 Token 认证） |
| Token 隐私机制 | 每次生成的子域名用随机 token 隔离，互不可见 | 已有 |

### 综合评价

基础实现，godnslog 已全面超越。

---

## 3. [ac0d3r/Hyuga](https://github.com/ac0d3r/Hyuga) ⭐

| 维度 | 内容 |
|------|------|
| 语言 | Go + Vue |
| 技术栈 | WebSocket + Caddy + 第三方通知 |
| Stars | 540 |

### 功能与架构

OOB 流量监控平台，支持 DNS、HTTP、LDAP、RMI 协议。核心特色：WebSocket 实时推送、DNS Rebinding、第三方通知（Bark/Lark/钉钉/飞书/Server酱）。

### 对比 godnslog

godnslog 已通过 listener 模块支持 LDAP，但 RMI 协议尚未明确支持。通知系统方面 godnslog 有 notification service，但渠道类型需要确认。

### 值得借鉴

| 特性 | 说明 | 优先级 |
|------|------|--------|
| ⭐⭐⭐ **WebSocket 实时推送** | Hyuga 使用 WebSocket 将日志实时推送到前端。godnslog 的 listener README 在 "Future Enhancements" 中列出此项但尚未实现 | **高** |
| ⭐⭐⭐ **第三方通知渠道** | Bark、Lark、钉钉、飞书、Server酱等多种通知渠道集成 | **高** |
| ⭐⭐ **RMI 协议监听** | 支持 RMI 协议回调捕获，godnslog 目前 listener 覆盖 LDAP/SMTP/SMB/FTP 但未见 RMI | **中** |

### 综合评价

Hyuga 的 **WebSocket 实时推送** 和 **第三方通知集成** 是 godnslog 可以直接吸收的高价值特性。

---

## 4. [SPuerBRead/Bridge](https://github.com/SPuerBRead/Bridge)

| 维度 | 内容 |
|------|------|
| 语言 | Java (Spring Boot) |
| 技术栈 | Netty + MySQL + Docker |
| Stars | 404 |

### 功能与架构

三层域名架构、DNS Rebinding、自定义 HTTP Response、数据查询 API。域名设计为 `ns.dnslog.com`（权威 DNS）/ `dns.dnslog.com`（Payload 域名）/ `dnslog.dnslog.com`（管理平台）。

### 对比 godnslog

godnslog 的架构更加现代化，Bridge 的 Java 单体架构较重。

### 值得借鉴

| 特性 | 说明 | 优先级 |
|------|------|--------|
| ⭐⭐⭐ **自定义 HTTP Response** | 可定制响应内容、状态码、Header。这对 SSRF 测试非常有用——模拟不同后端服务的响应 | **高** |
| ⭐⭐ **三层域名架构** | NS / Payload / Admin 三层域名分离，逻辑更清晰 | 中 |
| ⭐ **注册暗号机制** | 启动时传入注册暗号，限制未授权用户注册 | 低 |

### 综合评价

**自定义 HTTP Response** 是 SSRF 漏洞利用中的高频需求，godnslog 可以借鉴实现 UI 级别的响应定制能力。

---

## 5. [sa1tor/dnslog](https://github.com/sa1tor/dnslog)

| 维度 | 内容 |
|------|------|
| 语言 | Python (Tornado) |
| 技术栈 | Tornado + SQLite |
| Stars | 23 |

### 功能与架构

极轻量的 Python DNSLog，核心代码仅两个文件。

### 对比 godnslog

功能极为基础，参考价值有限。

### 综合评价

小型演示级项目，无值得吸收的特性。

---

## 6. [AlphabugX/Alphalog](https://github.com/AlphabugX/Alphalog) ⭐

| 维度 | 内容 |
|------|------|
| 语言 | Go |
| 技术栈 | Redis |
| Stars | 461 |

### 功能与架构

DNS/HTTP/RMI/LDAP/JNDI 综合日志平台。核心特色：Redis 存储、反弹 Shell 一键生成、SSRF 辅助路径、完全匿名设计。

### 值得借鉴

| 特性 | 说明 | 优先级 |
|------|------|--------|
| ⭐⭐⭐ **反弹 Shell 命令生成** | 通过 `fuzz.red/sh4ll/ip:port` 自动生成 bash/sh/nc/python/awk/telnet 多种反弹 Shell 命令 | **高** |
| ⭐⭐ **SSRF 辅助路径** | 提供 `/fuzz.red/ssrf/` 路径辅助 SSRF 测试 | 中 |
| ⭐⭐ **匿名设计哲学** | "完全匿名，不记录请求来源"，日志 1 天自动过期 | 中 |
| ⭐ ⭐**JNDI 支持** | 支持 JNDI 注入场景 | 低（godnslog 已有 LDAP） |
| ⭐ **纯 Redis 存储** | 部署更轻量，不需要关系型数据库 | 低 |

### 综合评价

**反弹 Shell 命令生成** 是 godnslog 目前缺失的实用功能，可以直接集成到 Payload Studio 中。匿名设计理念也可作为可选项引入。

---

## 7. [r00tSe7en/JNDIMonitor](https://github.com/r00tSe7en/JNDIMonitor)

| 维度 | 内容 |
|------|------|
| 语言 | Java |
| 技术栈 | 基于 JNDIExploit 精简 |
| Stars | ~100+ |

### 功能与架构

独立 LDAP 请求监听器，专门检测 JNDI 注入攻击。

### 综合评价

godnslog 的 listener 模块已有 LDAP 实现，功能覆盖。无额外借鉴价值。

---

## 8. [jiangsir404/POC-S](https://github.com/jiangsir404/POC-S)

| 维度 | 内容 |
|------|------|
| 语言 | Python |
| 技术栈 | POC-T 增强版 |
| Stars | 356 |

### 功能与架构

Web 漏洞验证框架，集成 POC 加载器、ZoomEye 搜索引擎、内置 DNSLog 服务。

### 值得借鉴

| 特性 | 说明 | 优先级 |
|------|------|--------|
| ⭐⭐ **搜索引擎集成** | ZoomEye/Shodan 批量目标搜索，可与 godnslog 的 scannerhub 结合 | 中 |
| ⭐⭐ **POC 加载框架** | 兼容 POC-T 语法的 POC 执行引擎 | 中 |
| ⭐ **框架与 POC 分离** | `pocs` 命令实现框架与脚本解耦 | 低 |

### 综合评价

godnslog 已有 workflow、rule、scannerhub 等模块，POC-S 的搜索引掌集成思路值得借鉴。但该项目的 DNSLog 部分很基础，无参考价值。

---

## 9. [Daybr4ak/ShiroScan](https://github.com/Daybr4ak/ShiroScan)

| 维度 | 内容 |
|------|------|
| 语言 | Java |
| 类型 | Burp Suite 插件 |
| Stars | 486 |

### 功能

Apache Shiro 漏洞检测 Burp 插件，已合并到 BurpShiroPassiveScan。

### 综合评价

Burp 插件而非 DNSLog 平台，与本项目关联度低。无可吸收的特性。

---

## 10. [wuba/Antenna](https://github.com/wuba/Antenna) ⭐⭐⭐

| 维度 | 内容 |
|------|------|
| 语言 | Python (Django) + Vue |
| 技术栈 | MySQL + Docker + Supervisor |
| Stars | 720 |

### 功能与架构

58同城安全团队开发的 OAST 平台。核心特性：插件化组件体系、任务驱动、多协议支持（DNS/HTTP/HTTPS/LDAP/RMI/FTP/JNDI/XSS/JSONP）、DNS Rebinding、CallBack & OpenAPI、邮箱通知、自定义 Template。

### 对比 godnslog

godnslog 在底层架构（Go 性能、HA、MCP）上更优，但 Antenna 的**插件化设计**和**模板系统**是独特的架构优势。

### 值得借鉴

| 特性 | 说明 | 优先级 |
|------|------|--------|
| ⭐⭐⭐ **插件化 Template 系统** | 用户可编写自定义检测组件（Template），将检测逻辑与平台解耦。godnslog 的 workflow 可以实现类似能力，但 Template 更轻量易用 | **高** |
| ⭐⭐⭐ **任务驱动模型** | 以任务为单位聚合多个检测场景，便于批量管理和执行。godnslog 有 case 模块，可借鉴任务化的组织方式 | **高** |
| ⭐⭐ **OpenAPI + CallBack** | Antenna_Inside 计划主动适配第三方扫描工具，实现漏洞检测流程闭环 | 中（godnslog 已有 openapi 模块） |
| ⭐⭐ **邮箱通知** | 支持 QQ 邮箱通知告警 | 中 |
| ⭐ **多协议 XSS/JSONP** | 除了标准 OOB 协议外还支持 XSS 和 JSONP 回调 | 低 |

### 综合评价

**Template 插件系统**和**任务驱动模型**是 Antenna 最值得 godnslog 吸收的架构设计。虽然 godnslog 的 workflow 功能更强大，但 Antenna 的 User Experience 更简洁。

---

## 汇总：功能矩阵

| 功能特性 | godnslog | Hyuga | Bridge | Alphalog | Antenna | 建议优先级 |
|----------|----------|-------|--------|----------|---------|-----------|
| DNS 日志 | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| HTTP 日志 | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| DNS Rebinding | ✅ | ✅ | ✅ | ❌ | ✅ | - |
| LDAP 监听 | ✅ | ❌ | ❌ | ✅ | ✅ | - |
| RMI 监听 | ❌ | ✅ | ❌ | ✅ | ✅ | 中 |
| SMTP 监听 | ✅ | ❌ | ❌ | ❌ | ❌ | - |
| SMB 监听 | ✅ | ❌ | ❌ | ❌ | ❌ | - |
| FTP 监听 | ✅ | ❌ | ❌ | ❌ | ✅ | - |
| JNDI 支持 | 部分 | ❌ | ❌ | ✅ | ✅ | 低 |
| WebSocket 实时推送 | ❌ | ✅ | ❌ | ❌ | ❌ | **高** |
| 第三方通知渠道 | 基础 | ✅ | ❌ | ❌ | ❌ | **高** |
| 自定义 HTTP Response | ❌ | ❌ | ✅ | ❌ | ❌ | **高** |
| 反弹 Shell 生成 | ❌ | ❌ | ❌ | ✅ | ❌ | **高** |
| 插件化组件(Template) | 部分 | ❌ | ❌ | ❌ | ✅ | **高** |
| 任务驱动模型 | 部分 | ❌ | ❌ | ❌ | ✅ | **高** |
| 搜索引擎集成 | ❌ | ❌ | ❌ | ❌ | ❌ | 中 |
| 匿名模式 | ❌ | ❌ | ❌ | ✅ | ❌ | 中 |
| 邮箱通知 | ❌ | ❌ | ❌ | ❌ | ✅ | 中 |
| SSRF 辅助路径 | ❌ | ❌ | ❌ | ✅ | ❌ | 中 |
| 多域名支持 | ❌ | ❌ | ❌ | ❌ | ❌ | 低 |
| POC 执行框架 | 部分 | ❌ | ❌ | ❌ | ❌ | 低 |

---

## 最终建议

### 短期可吸收（高优先级）

1. **WebSocket 实时推送** — Hyuga 的核心体验优势，godnslog listener README 已列为 Future Enhancement，建议优先实现
2. **反弹 Shell 命令生成** — Alphalog 特色功能，可集成到 Payload Studio，代码量小、用户感知强
3. **自定义 HTTP Response** — Bridge 的特色，增强 SSRF 测试场景的仿真能力

### 中期可吸收（中优先级）

4. **第三方通知渠道** — 借鉴 Hyuga 的 Bark/Lark/钉钉/飞书/Server酱 集成，扩展 godnslog 的 notification service
5. **插件化 Template 系统** — 借鉴 Antenna 的组件设计，提供用户自定义检测逻辑的轻量方式
6. **邮箱通知** — 基础但实用的告警通道
7. **RMI 协议监听** — 补齐协议覆盖

### 长期可吸收（低优先级）

8. **搜索引擎集成** — 与 scannerhub 结合实现 ZoomEye/Shodan 批量目标搜索
9. **匿名模式** — 提供完全匿名的使用选项，日志自动过期
10. **多域名支持** — 配置化多域名管理

---

## 11. 微信公众号文章（商业 dnslog 平台）

> 来源: https://mp.weixin.qq.com/s/8YovEBZq2VKNx4lRGfCccA
> 使用 Playwright 无头浏览器获取内容

### 平台概述

这是一套商业化的自建 dnslog 检测系统（非开源），集成 DNS/HTTP/LDAP/RMI/MySQL/FTP 多协议外带捕获，并在原始记录之上做了**利用类型识别、外带数据解码、IP 归属富化、攻击链关联、实时通知**等增强能力。

### 值得借鉴的功能 ⭐⭐⭐

| 特性 | 说明 | 优先级 |
|------|------|--------|
| ⭐⭐⭐ **攻击链时间线** | 按子域名 token 将 DNS/HTTP/LDAP 等多通道命中聚合成一条时间线，时间轴展示完整攻击过程（DNS解析→HTTP回连→LDAP/JNDI回连）。godnslog 有 interaction 模块，但缺少这种可视化关联聚合 | **高** |
| ⭐⭐⭐ **外带数据自动解码** | DNS 标签中的 base32/hex 自动还原明文，HTTP body 中的 base64 自动解码并展示。盲注外带数据的核心体验 | **高** |
| ⭐⭐⭐ **利用类型自动标注** | 根据 Payload 特征自动识别 Log4Shell/JNDI、Fastjson、SSRF、XXE、SQLi 盲注等利用类型 | **高** |
| ⭐⭐ **来源指纹归属** | 自动判别请求来自扫描器、云厂商还是真实目标，辅助判断漏洞真实性 | **中** |
| ⭐⭐ **Burp 风格轮询 API** | `/api/v1/poll` 游标增量拉取，类似 Burp Collaborator 的设计，工具链对接体验好 | **中** |
| ⭐⭐ **请求回显** | HTTP 响应体中回显收到的完整请求（请求行+全部请求头+body），方便调试排查 | **中** |
| ⭐⭐ **Payload 速查表** | 总览看板内置常用 Payload 速查表，复制即用 | **中** |
| ⭐ **GeoIP + ASN + rDNS 富化** | IP 归属地、ASN 信息、反向 DNS 等上下文富化展示 | 低 |

### 攻击链时间线设计详析

这是该平台最有亮点的设计：

```
┌─────────────────────────────────────────────┐
│  Attack Chain Timeline                      │
│  ┌──────────────────────────────────────┐   │
│  │  token: abc123.dnslog.wlog.fun      │   │
│  │  命中: DNS(3) HTTP(1) LDAP(1)       │   │
│  ├──────────────────────────────────────┤   │
│  │  ⏱ 10:00:01  DNS  query            │   │
│  │    ├─ 类型: JNDI                    │   │
│  │    ├─ 解码: [base32] jndi_payload   │   │
│  │    └─ 来源: ASNXXXX 云厂商          │   │
│  │  ⏱ 10:00:02  HTTP GET /            │   │
│  │    ├─ 类型: Log4Shell              │   │
│  │    ├─ 解码: [base64] admin:pass     │   │
│  │    └─ 来源: X.X.X.X 某省联通        │   │
│  │  ⏱ 10:00:03  LDAP connect          │   │
│  │    └─ 来源: X.X.X.X                │   │
│  └──────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

该设计对 godnslog 的 interaction + evidence 模块整合有直接参考价值。

### 综合评价

虽然该平台是闭源商业产品，但产品设计思路对 godnslog 有很高的参考价值。**攻击链时间线**、**外带数据自动解码**和**利用类型标注**是三个最值得吸收的产品特性，可以直接在 godnslog 现有的 interaction/evidence 模块基础上构建。

---

## 功能矩阵（完整版）

| 功能特性 | godnslog | 商业平台 | Hyuga | Bridge | Alphalog | Antenna | 优先级 |
|----------|----------|----------|-------|--------|----------|---------|--------|
| DNS 日志 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| HTTP 日志 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| DNS Rebinding | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | - |
| LDAP 监听 | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | - |
| RMI 监听 | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | 中 |
| SMTP 监听 | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | - |
| SMB 监听 | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | - |
| FTP 监听 | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | - |
| MySQL 协议监听 | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 低 |
| JNDI 支持 | 部分 | ✅ | ❌ | ❌ | ✅ | ✅ | 低 |
| WebSocket 实时推送 | ❌ | - | ✅ | ❌ | ❌ | ❌ | **高** |
| 第三方通知渠道 | 基础 | ✅ | ✅ | ❌ | ❌ | ❌ | **高** |
| 自定义 HTTP Response | ❌ | - | ❌ | ✅ | ❌ | ❌ | **高** |
| 反弹 Shell 生成 | ❌ | - | ❌ | ❌ | ✅ | ❌ | **高** |
| 插件化组件(Template) | 部分 | - | ❌ | ❌ | ❌ | ✅ | **高** |
| 任务驱动模型 | 部分 | - | ❌ | ❌ | ❌ | ✅ | **高** |
| 攻击链时间线 | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | **高** |
| 外带数据自动解码 | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | **高** |
| 利用类型自动标注 | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | **高** |
| 搜索引擎集成 | ❌ | - | ❌ | ❌ | ❌ | ❌ | 中 |
| 匿名模式 | ❌ | - | ❌ | ❌ | ✅ | ❌ | 中 |
| 邮箱通知 | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | 中 |
| SSRF 辅助路径 | ❌ | - | ❌ | ❌ | ✅ | ❌ | 中 |
| 来源指纹归属 | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 中 |
| Burp 风格轮询 API | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 中 |
| 请求回显 | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 中 |
| Payload 速查表 | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 中 |
| GeoIP/ASN 富化 | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 低 |
| 多域名支持 | ❌ | - | ❌ | ❌ | ❌ | ❌ | 低 |

---

*报告生成日期: 2026-07-11*
*数据来源: GitHub 仓库 README、公开文档、微信公众号文章（Playwright 获取）*
