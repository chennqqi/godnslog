# Phase 17 Spec B: 实时推送与通知扩展

> WebSocket 实时推送 + 通知渠道扩展（Bark/Server酱/Telegram/Slack）
> 对应 ROADMAP 2.4 智能增强版

## 1. 概述

两个独立的部分：

- **WebSocket Hub**：Interaction 写入后通过 WebSocket 广播到前端，替代轮询刷新。后端 Hub + 前端客户端。
- **通知渠道扩展**：在现有 notification service 中新增 4 个渠道：Bark、Server酱、Telegram、Slack。

## 2. WebSocket 实时推送

### 2.1 架构

```
Interaction Created (service.go)
       │
       ▼
┌──────────────┐
│  WebSocket   │
│  Hub         │  ← 全局单例，管理所有连接
│              │
│  ┌────────┐  │
│  │ Client1│  │
│  ├────────┤  │
│  │ Client2│  │
│  ├────────┤  │
│  │ ...    │  │
│  └────────┘  │
└──────┬───────┘
       │
       ▼
   Gin Route: GET /ws
   (需要 authHandler 认证)
```

### 2.2 包结构

```
internal/websocket/
  hub.go       # Hub 管理所有客户端连接，广播消息
  client.go    # 单个 WebSocket 连接的管理
  hub_test.go  # 测试
```

### 2.3 Hub 设计

```go
// Hub 管理所有 WebSocket 客户端连接
type Hub struct {
    clients    map[*Client]bool  // 所有注册的客户端
    broadcast  chan []byte       // 广播通道
    register   chan *Client      // 注册通道
    unregister chan *Client      // 注销通道
    mu         sync.RWMutex
}

func NewHub() *Hub
func (h *Hub) Run()                // 启动 Hub 事件循环
func (h *Hub) Broadcast(msg []byte) // 向所有客户端广播消息
```

### 2.4 Client 设计

```go
// Client 代表一个 WebSocket 连接
type Client struct {
    hub  *Hub
    conn *gorillawebsocket.Conn
    send chan []byte
}

func NewClient(hub *Hub, conn *gorillawebsocket.Conn) *Client
func (c *Client) ReadPump()   // 读取 goroutine（处理心跳/pong）
func (c *Client) WritePump()  // 写入 goroutine（从 send chan 读取并写 conn）
```

### 2.5 消息协议

```json
{
    "type": "interaction",
    "payload": { /* Interaction JSON */ }
}
```

仅广播 `type: "interaction"` 一种消息类型。前端按需过滤。

### 2.6 后端集成

**Gin 路由**（在 `webserver.go` 的 `Run()` 中注册）：

```go
// WebSocket endpoint (auth required)
ws.GET("/ws", self.authHandler, self.wsHandler)
```

**WebSocket Handler**（在 `webserver.go` 新增）：

```go
var wsHub = websocket.NewHub()

func (self *WebServer) initWS() {
    go wsHub.Run()
}

func (self *WebServer) wsHandler(c *gin.Context) {
    // Upgrade HTTP → WebSocket
    upgrader := gorillawebsocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    client := websocket.NewClient(wsHub, conn)
    wsHub.Register(client)
    go client.WritePump()
    go client.ReadPump()
}
```

**Interaction Service 集成**：在 `enhanceInteraction` 之后、`InsertOne` 之后广播：

```go
// 在 service.go 的 CreateInteraction 末尾
if s.wsHub != nil {
    if data, err := json.Marshal(interaction); err == nil {
        s.wsHub.Broadcast(data)
    }
}
```

### 2.7 配置

不需要额外配置。WebSocket 升级复用现有 HTTP 端口。默认启用。

### 2.8 依赖

```go
import "github.com/gorilla/websocket" // v1.5
```

### 2.9 前端集成

**新文件** `frontend-next/src/lib/ws-client.ts`：

```typescript
type WSCallback = (data: unknown) => void;

class WSClient {
    private ws: WebSocket | null = null;
    private listeners: Map<string, WSCallback[]> = new Map();
    private reconnectTimer: number | null = null;
    private baseURL: string;

    constructor(baseURL?: string) {
        this.baseURL = baseURL || (typeof window !== 'undefined' ? window.location.host : 'localhost:8080');
    }

    connect(): void {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const url = `${protocol}//${this.baseURL}/ws`;
        this.ws = new WebSocket(url);

        this.ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                const typeListeners = this.listeners.get(msg.type) || [];
                typeListeners.forEach(cb => cb(msg.payload));
            } catch {}
        };
        this.ws.onclose = () => {
            this.scheduleReconnect();
        };
    }

    on(type: string, callback: WSCallback): void {
        if (!this.listeners.has(type)) this.listeners.set(type, []);
        this.listeners.get(type)!.push(callback);
    }

    disconnect(): void {
        if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
        this.ws?.close();
        this.ws = null;
    }

    private scheduleReconnect(): void {
        this.reconnectTimer = window.setTimeout(() => this.connect(), 5000);
    }
}

export const wsClient = new WSClient();
```

**集成点**：在 `layout.tsx` 或 `app-shell` 组件中初始化 WS 连接，监听 `interaction` 事件后触发 TanStack Query 的 invalidate。

### 2.10 边界处理

- **断线重连**：前端自动 5 秒后重连
- **认证**：复用现有 Cookie/Token 认证（WebSocket 连接在认证后建立）
- **并发**：Hub 使用 channel 做 goroutine-safe 的消息分发
- **心跳**：gorilla/websocket 内置 ping/pong 处理
- **大规模连接**：Hub 适合数百并发，更多需考虑 Redis Pub/Sub（非当前范围）

### 2.11 测试

```go
func TestHubBroadcast(t *testing.T) {
    hub := NewHub()
    go hub.Run()

    // 模拟客户端
    msg := []byte(`{"type":"interaction","payload":{}}`)
    hub.Broadcast(msg)
    // 验证所有客户端收到消息
}
```

## 3. 通知渠道扩展

### 3.1 架构

在现有 `internal/notification/service.go` 的 `SendNotification` switch 中新增 4 个 case：

```go
case "bark":
    sendErr = s.sendBark(channel.Config, message, payload)
case "serverchan":
    sendErr = s.sendServerchan(channel.Config, message, payload)
case "telegram":
    sendErr = s.sendTelegram(channel.Config, message, payload)
case "slack":
    sendErr = s.sendSlack(channel.Config, message, payload)
```

### 3.2 各渠道配置与实现

#### Bark (iOS 推送)

配置 JSON：
```json
{
    "url": "https://api.day.app/xxxx"
}
```

实现：
```go
func (s *Service) sendBark(config, message, payload string) error {
    // GET https://api.day.app/xxxx/{title}/{body}
    // 或 POST 到 /push 带 JSON body
}
```

#### Server酱 (微信推送)

配置 JSON：
```json
{
    "sendkey": "xxxx"
}
```

实现：
```go
func (s *Service) sendServerchan(config, message, payload string) error {
    // POST https://sctapi.ftqq.com/{sendkey}.send
    // title=xxx&desp=xxx
}
```

#### Telegram Bot

配置 JSON：
```json
{
    "bot_token": "123:abc",
    "chat_id": "123456"
}
```

实现：
```go
func (s *Service) sendTelegram(config, message, payload string) error {
    // POST https://api.telegram.org/bot{token}/sendMessage
    // chat_id={chat_id}&text={text}
}
```

#### Slack Webhook

配置 JSON：
```json
{
    "webhook_url": "https://hooks.slack.com/services/xxx"
}
```

实现：
```go
func (s *Service) sendSlack(config, message, payload string) error {
    // POST {webhook_url} with {"text":"..."}
}
```

### 3.3 数据模型更新

`models/v2.go` 中更新 channel type 验证：

```go
Type string `json:"type" binding:"required,oneof=webhook wechat feishu dingtalk bark serverchan telegram slack"`
```

### 3.4 通知消息格式

各渠道统一消息模板：

```
[GODNSLOG] {message}

{payload}

时间: {time}
```

- Bark、Server酱、Telegram 使用此格式
- Slack 使用消息块（blocks）格式以获得更好展示

### 3.5 测试

```go
func TestSendBark(t *testing.T) {
    svc := NewService(nil)
    // mock HTTP 请求，验证请求格式
}

func TestSendTelegram(t *testing.T) {
    // 同上
}
```

### 3.6 依赖

无外部依赖。全部使用 `net/http` 发送 HTTP POST/GET 请求，与现有通知渠道一致。

## 4. 安全性

- WebSocket 端点复用现有 `authHandler` 认证中间件
- 通知渠道的 API token/key 存储在数据库加密字段（现有 `Config` text 字段）
- 通知请求使用 HTTPS（由 URL 协议控制）

## 5. 交付物清单

**WebSocket 部分：**
- `internal/websocket/hub.go` — Hub 管理
- `internal/websocket/client.go` — 客户端连接管理
- `internal/websocket/hub_test.go` — Hub 测试
- `server/websocket.go` — Gin handler + Hub 初始化
- `server/webserver.go` — WS 路由注册 + Interaction service WS Hub 注入
- `internal/interaction/service.go` — Broadcast 调用（新增字段 `wsHub`）
- `frontend-next/src/lib/ws-client.ts` — WS 客户端

**通知渠道部分：**
- `internal/notification/service.go` — 新增 4 个发送函数 + switch cases
- `models/v2.go` — 更新 channel type 验证
- `internal/notification/service_test.go` — 新增渠道测试
