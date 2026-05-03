# GODNSLOG 2.0 Frontend

Next.js 14 + TypeScript + Tailwind CSS

## 安装依赖

```bash
npm install
```

## 环境变量

创建 `.env.local` 文件：

```
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v2
```

## 开发

```bash
npm run dev
```

访问 http://localhost:3000

## 构建

```bash
npm run build
```

## 启动生产服务器

```bash
npm start
```

## 项目结构

```
src/
├── app/
│   ├── dashboard/
│   │   ├── cases/          # Case管理
│   │   ├── payloads/       # Payload管理
│   │   ├── interactions/   # Interaction管理
│   │   └── settings/       # 系统设置
│   ├── login/              # 登录页面
│   ├── layout.tsx          # 根布局
│   ├── page.tsx            # 首页
│   └── globals.css         # 全局样式
├── components/             # 组件
├── lib/
│   ├── api.ts             # API基础客户端
│   ├── api-client.ts      # API客户端封装
│   └── utils.ts           # 工具函数
└── types/
    └── index.ts           # TypeScript类型定义
```

## 功能特性

- 用户认证（JWT）
- Case管理（创建、查看、列表）
- Payload管理（列表、搜索）
- Interaction管理（列表、筛选、详情、导出）
- 系统设置（通用、域名、监听、通知、Token管理）

## 注意事项

- 当前使用localStorage存储token，生产环境应使用更安全的方案
- 需要后端API支持才能正常运行
- lint错误是因为尚未安装依赖，安装后会消失
