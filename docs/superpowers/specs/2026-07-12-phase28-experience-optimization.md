# Phase 2.8: Experience Optimization

## 概述

本阶段聚焦产品体验优化，涵盖通知系统增强、GeoIP 自动更新、前端可视化组件补全三大方向。

## 1. 通知系统增强

### 1.1 通知发送 HTTP 超时配置

**现状：** `internal/notification/service.go` 中全部 8 个发送方法（webhook/wechat/feishu/dingtalk/bark/serverchan/telegram/slack）均使用裸 `http.Post()` / `http.PostForm()`，无超时控制。外部服务无响应时 goroutine 永久阻塞。

**改动：**

1. `Service` 结构体新增 `httpClient *http.Client` 字段
2. `NewService` 签名从 `NewService(engine *xorm.Engine)` 改为 `NewService(engine *xorm.Engine, opts ...Option)`，支持以下 Option：
   - `WithHTTPTimeout(d time.Duration)` — 默认 30 秒
3. 内部新增 `sendJSON(url string, body interface{}) error` 和 `sendForm(url string, data url.Values) error` 辅助方法，替代所有重复的 `http.Post()` 调用
4. 所有 8 个 `sendXxx` 方法改为使用 `s.httpClient.Post()` / `s.httpClient.PostForm()`
5. 服务层创建处（`server/v2_api.go` 中的 `v2CreateNotificationChannel` 和 其他初始化路径）更新 NewService 调用

**影响范围：** `internal/notification/service.go`（核心修改）+ `server/v2_api.go`（更新调用）

### 1.2 Telegram Markdown 转义处理

**现状：** `sendTelegram()` 使用 `parse_mode: Markdown`，但消息中的特殊字符（`_*[]()~>#+-=|{}.!`）未经转义，导致 Telegram API 返回 400 错误。

**改动：**

1. 在 `internal/notification/service.go` 中新增 `escapeTelegramMarkdown(s string) string` 函数
2. 对以下字符加反斜杠前缀：`_ * [ ] ( ) ~ ` > # + - = | { } . !`
3. 在 format 消息时先对 `message` 和 `payload` 做转义处理
4. 行首的 `>` 和 `#` 也需特别处理

**参考转义规则：**
```go
var telegramMarkdownEscaper = strings.NewReplacer(
    "_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]",
    "(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`",
    ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-",
    "=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}",
    ".", "\\.", "!", "\\!",
)
```

### 1.3 WebSocket Hub 优雅关闭

**现状：** `internal/websocket/hub.go` 的 `Run()` 方法无限循环，没有关闭机制。服务器关闭时活跃连接可能丢失消息，write pump 未正常退出。

**改动：**

1. `Hub` 结构体新增：
   - `shutdown chan struct{}` — 关闭信号
   - `wg sync.WaitGroup` — 追踪活跃 write pump
   - `closed atomic.Bool` — 标记是否已关闭

2. `Hub` 新增 `Shutdown(ctx context.Context) error` 方法：
   - 检查 `closed` 避免重复关闭
   - 设置 `closed = true` 阻止新注册
   - 给所有已注册 client 发送关闭帧（`CloseNormalClosure`）
   - 关闭广播 channel 让 broadcast case 退出
   - `wg.Wait()` + ctx.Done() 超时保护（最多等待 5s）

3. `Hub.Run()` 的 select 新增 `<-h.shutdown` case

4. `Client.WritePump()` ：
   - 使用 `wg.Add(1)` / `wg.Done()` 跟踪
   - 处理 `send` channel 关闭后的正常退出

5. `servecmd.go` 的 shutdown 序列中补充 `web.Shutdown()` → `wsHub.Shutdown(ctx)`

## 2. GeoIP 自动更新

### 2.1 配置与自动下载

**现状：** GeoIP 后端逻辑（`internal/interaction/fingerprint/geoip.go`）已实现但从未在生产路径中启用。所有 `interaction.NewService()` 调用都传 `nil` fingerprinter。

**改动：**

1. `servecmd.go` `servePwCmd` 结构体新增：
   - `mmdbPath string` — GeoLite2-ASN.mmdb 文件路径，默认 `./data/GeoLite2-ASN.mmdb`
   - `mmdbLicenseKey string` — MaxMind License Key，可选，用于自动下载

2. `SetFlags` 新增：
   - `-geoip-mmdb` — mmdb 路径
   - `-geoip-license-key` — MaxMind License Key，可选，用于自动下载
   - 环境变量 `MMDB_PATH` / `MMDB_LICENSE_KEY`
   - flag 描述中提示免费 Key 申请链接：`https://www.maxmind.com/en/geolite2/signup`

3. `Execute` 启动序列中新增：
   - 检查 mmdbPath 文件是否存在
   - 如不存在且 licenseKey 不为空 -> 自动下载
   - 下载 URL: `https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key={key}&suffix=tar.gz`
   - 下载后解压 tar.gz，提取 mmdb 文件到目标路径
   - 如不存在且 licenseKey 为空 -> log 警告：未配置 license key，GeoIP 自动更新已禁用，可在 https://www.maxmind.com/en/geolite2/signup 免费申请
   - 下载失败 -> 仅 log 警告，不阻止启动

4. `fingerprint.NewFingerprint(mmdbPath)` 传入有效路径创建 fingerprinter，而非传空

5. `WebServerConfig` 新增 `MMDBPath string` 字段

6. `NewWebServer` 中创建 interaction service 时使用有效的 fingerprinter

**下载实现要点：**
```go
func downloadMMDB(path, licenseKey string) error {
    url := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=%s&suffix=tar.gz", licenseKey)
    // 下载到临时文件
    // 解压 tar.gz
    // 从 GeoLite2-ASN_*/GeoLite2-ASN.mmdb 复制到目标路径
    // 清理临时文件
    return nil
}
```

### 2.2 持久化 ASN/Org

**现状：** `enhanceInteraction()` 调用 `fp.Lookup()` 获取到 `Fingerprint{ASN, Org, Country}` 后，只存了 `SourceType` 和 `SourceName`，ASN/Org/Country 被丢弃。

**改动：**

1. `internal/models/interaction.go` `Interaction` 结构体新增三个字段：
   ```go
   ASN     *uint   `json:"asn,omitempty" xorm:"'asn' int"`
   Org     *string `json:"org,omitempty" xorm:"'org' varchar(255)"`
   Country *string `json:"country,omitempty" xorm:"'country' varchar(64)"`
   ```

2. `enhanceInteraction()` 中补充 ASN/Org/Country 的存储
3. 前端 TypeScript 类型 `Interaction` 同步新增字段
4. 前端类型文件：`frontend-next/src/types/index.ts`

## 3. 前端 UI 组件补全

### 3.1 图表组件

**库选择：** `recharts` — React 原生声明式 API，TypeScript 友好，与现有 Tailwind/Radix 无冲突。

**安装：** `npm install recharts`

#### DonutChart 环形图

**用途：** 仪表盘协议分布、统计数据占比展示

**组件接口：**
```tsx
interface DonutChartProps {
  data: { name: string; value: number; color?: string }[]
  size?: number  // 默认 200
  innerRadius?: number  // 默认 60
  outerRadius?: number  // 默认 80
  showLabel?: boolean
  className?: string
}
```

**位置：** `frontend-next/src/components/charts/donut-chart.tsx`

**协议颜色映射（一致配色方案）：**
- dns: `#3b82f6` (blue-500)
- http: `#10b981` (emerald-500)
- smtp: `#f59e0b` (amber-500)
- ldap: `#8b5cf6` (violet-500)
- smb: `#ec4899` (pink-500)
- ftp: `#f97316` (orange-500)
- other: `#6b7280` (gray-500)

#### LineChart 折线图

**用途：** 交互趋势展示（按日/按小时统计）

**组件接口：**
```tsx
interface LineChartProps {
  data: { date: string; count: number; label?: string }[]
  height?: number  // 默认 300
  showGrid?: boolean
  showTooltip?: boolean
  color?: string
  className?: string
}
```

**位置：** `frontend-next/src/components/charts/line-chart.tsx`

#### 后端接口（新增）

折线图需要按日统计 API：

**`GET /api/v2/interactions/stats/daily`**

参数：`case_id?`、`payload_id?`、`days`（默认 7，最大 90）

返回：
```json
{
  "code": 0,
  "data": [
    {"date": "2026-07-06", "count": 12},
    {"date": "2026-07-07", "count": 8},
    {"date": "2026-07-08", "count": 0},
    ...
  ]
}
```

实现：`server/v2_api.go` 新增 `v2InteractionDailyStats` handler，按 UTC 日期分组计数。

#### Dashboard 集成

- 用 DonutChart 替换现有的 CSS 协议分布条
- 新增"近期趋势"折线图区域，显示近 7 天交互趋势
- 保持 stat cards 和 live stream 不变

### 3.2 Kanban Board

**库选择：** `@dnd-kit/core` + `@dnd-kit/sortable` — 轻量、无障碍、React 优先的拖拽库。

**安装：** `npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities`

**组件架构：**

| 组件 | 路径 | 用途 |
|------|------|------|
| `KanbanBoard` | `frontend-next/src/components/kanban/kanban-board.tsx` | 主容器，管理 DnD context 和列状态 |
| `KanbanColumn` | `frontend-next/src/components/kanban/kanban-column.tsx` | 单列（droppable），显示列标题 + 卡片列表 |
| `KanbanCard` | `frontend-next/src/components/kanban/kanban-card.tsx` | 单张卡片（draggable），显示标题/描述/标签 |

**状态列映射：**
| 列 | 对应 Case Status | 颜色 |
|----|-----------------|------|
| Active | `active` | `#3b82f6` (blue-500) |
| Completed | `completed` | `#10b981` (emerald-500) |
| Archived | `archived` | `#6b7280` (gray-500) |

**Cases 页面集成：**

- 在现有 Cases 列表页（`frontend-next/src/app/dashboard/cases/page.tsx`）添加 Table / Board 视图切换按钮
- Board 视图：三列横向排列，卡片按 status 分组
- 拖拽操作：将卡片拖到其他列 → 调用 `PUT /api/v2/cases/:id` 更新 status
- 乐观更新（先更新本地状态，API 失败时回滚）
- 卡片点击：维持现有导航到 `/dashboard/cases/[id]`
- 空状态：列内无卡片时显示 "No cases" 占位文字
- 加载状态：`<LoadingSkeleton />`
- 错误状态：`<ErrorBoundary />`

## 4. 状态修正

ROADMAP 中的 `❌ app-shell / sidebar / top-bar` 三个布局组件已在 `frontend-next/src/components/app-shell/` 中完整实现，应改为 `✅`。

## 5. 文件变更清单

### Backend
| 文件 | 变更 |
|------|------|
| `internal/notification/service.go` | 添加 HTTP 超时、Telegram 转义、重构为 httpClient |
| `internal/websocket/hub.go` | 添加 Shutdown 方法 + 优雅关闭逻辑 |
| `internal/websocket/client.go` | 添加 WaitGroup 跟踪 |
| `internal/models/interaction.go` | 添加 ASN/Org/Country 字段 |
| `internal/interaction/service.go` | enhanceInteraction 存储 ASN/Org |
| `server/v2_api.go` | 新增 `v2InteractionDailyStats` handler |
| `server/webserver.go` | WebServerConfig 添加 MMDBPath；路由注册 |
| `servecmd.go` | 添加 CLI flag、mmdb 下载、fingerprinter 串联 |
| `config/config.go` | 添加 MMDB 配置项 |

### Frontend
| 文件 | 变更 |
|------|------|
| `package.json` | 添加 recharts, @dnd-kit/core, @dnd-kit/sortable |
| `src/components/charts/donut-chart.tsx` | 新建环形图组件 |
| `src/components/charts/line-chart.tsx` | 新建折线图组件 |
| `src/components/charts/index.ts` | barrel export |
| `src/components/kanban/kanban-board.tsx` | 新建看板主组件 |
| `src/components/kanban/kanban-column.tsx` | 新建看板列组件 |
| `src/components/kanban/kanban-card.tsx` | 新建看板卡片组件 |
| `src/components/kanban/index.ts` | barrel export |
| `src/app/dashboard/page.tsx` | 集成 DonutChart + LineChart |
| `src/app/dashboard/cases/page.tsx` | 添加 Table/Board 视图切换 + Kanban 集成 |
| `src/types/index.ts` | Interaction 添加 ASN/Org/Country 字段 |
| `src/lib/api-client.ts` | 新增 daily stats API 调用 |

## 6. 拒绝的特性

- **GeoIP City/Country mmdb 下载**：当前使用 GeoLite2-ASN.mmdb，不含城市/国家数据。如需完整地理位置需额外下载 GeoLite2-City.mmdb，但这超出了本阶段范围。
- **图表动画/交互动效**：recharts 内置默认动画，不额外定制复杂动画。
- **Kanban 多列重排（列级别拖拽）**：仅支持卡片在列间拖动，不支持列本身的重排序。
- **WebSocket room/channel**：优雅关闭不引入话题/频道功能，仅保证进程退出时安全关闭。
- **通知超时 per-channel 配置**：使用全局超时时间，不引入 per-channel 配置复杂度。
