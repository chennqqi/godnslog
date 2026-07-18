# Captcha 验证码方案开发资料

## 1. 资料来源

- Confluence 页面：[captcha](https://wiki.threatbook-inc.cn/display/~chenqi/captcha)
- 页面 ID：`261015141`
- 页面版本：`6`
- 页面更新时间：`2025-11-05 12:15:52 +08:00`
- 采集日期：`2026-07-18`
- 子页面：无
- 附件：无

本文档将 Confluence 页面中的候选方案整理为项目开发资料。第 2 节是对页面内容的转录和结构化整理；后续章节为结合 OneSandbox 开发场景补充的接入评估项，不代表 Confluence 页面已经作出的技术选型。

## 2. Confluence 候选方案

| 仓库 | 页面描述 | 验证码类型 | 页面强度评价 |
| --- | --- | --- | --- |
| [mojocn/base64Captcha](https://github.com/mojocn/base64Captcha) | captcha of base64 image string | 图片数字文字及干扰 | 中；强度可配置 |
| [dchest/captcha](https://github.com/dchest/captcha) | Go package captcha implements generation and verification of image and audio CAPTCHAs | 数字文字、音频及干扰 | 已被攻破 |
| [wenlng/go-captcha](https://github.com/wenlng/go-captcha) | GoCaptcha is a behavior captcha, which implements click mode, slider mode, drag-drop mode and rotation mode | 滑动等行为验证码 | 强 |
| [samwafgo/cap_go_server](https://github.com/samwafgo/cap_go_server) | Native Go server implementation | cap.js 行为验证码 | 强 |
| [ackcoder/go-cap](https://github.com/ackcoder/go-cap) | Golang implementation of `@cap.js/server` | cap.js 行为验证码 | 强 |

说明：Confluence 页面使用动态 GitHub star 徽章，没有保存固定 star 数量，因此本文档不记录可能快速过期的 star 数。

## 3. 初步分类

### 3.1 图片或音频挑战

- `mojocn/base64Captcha`：支持图片数字、文字和干扰元素，页面评价为中等强度且可配置；
- `dchest/captcha`：支持数字图片和音频，但页面明确标记为“已被攻破”，不应作为默认候选。

### 3.2 行为挑战

- `wenlng/go-captcha`：支持点击、滑块、拖拽和旋转等多种交互模式；
- `samwafgo/cap_go_server`：cap.js 的 Go 服务端实现；
- `ackcoder/go-cap`：`@cap.js/server` 的 Golang 实现。

Confluence 页面将三个行为验证码候选均评价为“强”，但没有给出测试方法、攻击模型或选型结论。正式选型前仍需独立验证。

## 4. OneSandbox 接入评估项

以下内容是项目开发需要补充验证的事项，不是 Confluence 原文结论。

### 4.1 功能适配

- 是否支持完全离线部署；
- 是否依赖公网、第三方 SaaS、CDN 或远程字体和图片；
- Go 版本及前端框架兼容性；
- 是否能嵌入现有登录流程；
- 是否支持中文界面和自定义主题；
- 是否支持无障碍替代方案；
- 是否适配国产化浏览器和部署环境；
- 是否能在单机和集群模式下保持一致行为。

### 4.2 安全性

- challenge 是否由密码学安全随机源生成；
- challenge 是否一次性使用并具有短有效期；
- 答案或可逆推导信息是否会返回客户端；
- 服务端是否只保存必要的摘要或状态；
- 是否防止重放、暴力枚举和跨会话复用；
- 是否绑定登录会话、客户端上下文和用途；
- 是否具备请求频率限制及失败退避；
- 行为验证码的验证逻辑是否真正位于服务端；
- 集群节点间是否共享 challenge 状态，或使用可验证的无状态签名方案；
- 日志、审计和错误响应是否避免泄露答案、token 和内部评分。

### 4.3 维护性与供应链

- 开源许可证是否允许商业产品集成和再分发；
- 仓库维护活跃度、最近版本和未解决安全问题；
- 直接及间接依赖数量；
- 是否存在 CGO、浏览器原生扩展或额外运行时依赖；
- 是否便于固定版本、生成 SBOM 和执行漏洞扫描；
- 上游停止维护后是否具备内部接管成本；
- 前后端组件及协议版本是否必须严格配套。

### 4.4 用户体验与性能

- 首屏资源大小和生成耗时；
- 弱网或高延迟环境中的可用性；
- 移动端和不同分辨率适配；
- 对键盘、读屏和色觉障碍用户的支持；
- 挑战失败后的刷新与重试体验；
- 行为挑战误拒绝率和自动化攻击拦截率；
- 高并发登录时的 CPU、内存和状态存储成本。

## 5. 建议验证顺序

1. 排除 Confluence 已标记“已被攻破”的 `dchest/captcha`；
2. 对 `mojocn/base64Captcha` 做低成本图片验证码 PoC，验证现有登录接口改造量；
3. 从 `wenlng/go-captcha` 和两个 cap.js Go 服务端实现中各选一个做行为验证码 PoC；
4. 在无公网环境完成安装、构建和运行验证；
5. 对 challenge 生命周期、重放、暴力枚举、集群一致性和日志泄露做安全测试；
6. 对许可证、维护状态、依赖漏洞和升级成本做供应链评审；
7. 根据安全性、可用性、接入成本和维护成本形成正式选型记录。

## 6. 推荐的登录流程边界

验证码应作为登录风控的一部分，不能替代密码认证、账号锁定或二因素认证。建议流程：

```text
请求登录挑战
→ 服务端生成 captcha challenge
→ 客户端完成挑战
→ 服务端一次性验证 challenge
→ 校验用户名和密码
→ 应用失败计数、限流和账号锁定策略
→ 登录成功或返回统一错误
```

建议约束：

- challenge 由服务端生成，默认有效期不超过 2 分钟；
- 验证成功或失败后立即作废，禁止重复使用；
- 同一客户端连续失败时提高挑战强度或触发冷却；
- 登录错误响应避免区分“用户不存在”“密码错误”和“验证码错误”造成的账号枚举风险；
- challenge 存储使用现有 cache 抽象，集群部署必须使用共享后端或可靠的无状态验证方案；
- 验证码开关、类型和强度应通过统一后端配置管理，不写死在前端；
- 所有关键事件进入安全审计，但不记录挑战答案和敏感 token。

## 7. 待确认事项

- 验证码是所有登录强制启用，还是仅在失败、异常 IP 或高风险场景触发；
- 一期采用图片验证码还是行为验证码；
- 是否要求音频或其他无障碍替代方式；
- 是否需要和统一二次验证、双因素认证规划同时设计；
- challenge 的默认有效期、失败阈值和冷却时间；
- 单机与集群是否统一使用共享缓存；
- 是否允许前端引入 cap.js 等额外依赖；
- 最终选型所需的许可证与安全评审负责人。
