# video-parser

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue3-3.5.13-42b883?style=for-the-badge&logo=vue.js)](https://vuejs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=for-the-badge&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](https://opensource.org/licenses/MIT)
[![Stars](https://img.shields.io/github/stars/2759069519/video-parser?style=for-the-badge)](https://github.com/2759069519/video-parser/stargazers)
[![Forks](https://img.shields.io/github/forks/2759069519/video-parser?style=for-the-badge)](https://github.com/2759069519/video-parser/network)

### 简洁优雅的视频解析聚合平台

**多平台支持** | **快速响应** | **现代架构**

[在线预览](https://www.gjlove.cn) | [功能介绍](#-功能特性) | [快速开始](#-快速开始) | [API 文档](#-api-接口)

---

## 📷 Screenshot

```bash
+--------------------------------------------------+
|                                                  |
|              video-parser                        |
|                                                  |
|  +--------------------------------------------+ |
|  |                                            | |
|  |    [Video Parser Platform]                | |
|  |                                            | |
|  |    Input URL: [_______________] [Parse]   | |
|  |                                            | |
|  |    Platform: Douyin | Kuaishou | Red     | |
|  |                                            | |
|  +--------------------------------------------+ |
|                                                  |
+--------------------------------------------------+
```

---

## 📋 目录

- [功能特性](#-功能特性)
- [技术架构](#-技术架构)
- [快速开始](#-快速开始)
- [API 接口](#-api-接口)
- [项目结构](#-项目结构)
- [配置说明](#-配置说明)
- [部署建议](#-部署建议)
- [常见问题](#-常见问题)
- [贡献指南](#-贡献指南)

---

## ✨ 功能特性

| 平台 | 状态 | 说明 |
|------|------|------|
| 抖音 | ✅ | 解析抖音短视频链接 |
| 快手 | ✅ | 解析快手视频链接 |
| 小红书 | ✅ | 解析小红书笔记/视频 |
| 米游社 | ✅ | 解析米游社帖子/视频 |

| 功能 | 状态 | 说明 |
|------|------|------|
| 响应式设计 | ✅ | 完美适配移动端和桌面端 |
| 短链分享 | ✅ | 生成的分享链接简洁易记 |
| RESTful API | ✅ | 清晰的 API 设计 |
| 健康检查 | ✅ | 服务状态实时监控 |

---

## 🏗 技术架构

```
                    +-----------------+
                    |   Vue3 Frontend |
                    |    (Vite)       |
                    +--------+--------+
                             |
                    +--------v---------+
                    |   Reverse Proxy   |
                    |   (Nginx/Caddy)  |
                    +--------+--------+
                             |
                    +--------v---------+
                    |    Go Backend    |
                    |    (Gin)        |
                    +--------+--------+
                             |
              +--------------+--------------+
              |                              |
     +--------v--------+         +----------v---------+
     |  PostgreSQL DB  |         |   File Storage    |
     +-----------------+         +------------------+
```

### 后端技术

| 技术 | 说明 |
|------|------|
| Go 1.20+ | 高性能后端语言 |
| Gin | 轻量级 Web 框架 |
| GORM | Go ORM 库（含连接池配置） |
| PostgreSQL 15 | 关系型数据库 |

### 前端技术

| 技术 | 说明 |
|------|------|
| Vue 3 | 渐进式 JavaScript 框架 |
| Vite | 下一代前端构建工具 |
| TypeScript | JavaScript 超集 |

---

## 🚀 快速开始

### 环境要求

| 依赖 | 最低版本 |
|------|----------|
| Go | 1.20 |
| Node.js | 18 |
| PostgreSQL | 15 |

### 安装步骤

#### 1. 克隆项目

```bash
git clone https://github.com/2759069519/video-parser.git
cd video-parser
```

#### 2. 配置数据库

```bash
# 登录 PostgreSQL
psql -U postgres

# 创建数据库
CREATE DATABASE video_parser;

# 创建用户（可选）
CREATE USER your_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE video_parser TO your_user;
```

#### 3. 配置环境变量

```bash
# 后端配置
export DB_HOST=localhost         # 数据库主机
export DB_PORT=5432              # 数据库端口
export DB_USER=postgres          # 数据库用户
export DB_PASSWORD=postgres      # 数据库密码
export DB_NAME=video_parser     # 数据库名称
export SERVER_PORT=8080         # 服务端口
```

#### 4. 启动服务

```bash
# 启动后端
go run cmd/server/main.go

# 启动前端（开发模式）
cd frontend
npm install
npm run dev
```

#### 5. 访问应用

```
前端地址: http://localhost:3000
后端地址: http://localhost:8080
健康检查: http://localhost:8080/health
```

---

## 📡 API 接口

### 基础信息

- 基础 URL: `http://localhost:8080`
- 认证方式: 无
- 响应格式: JSON

### 接口列表

#### 1. 健康检查

```http
GET /health
```

**响应示例:**
```json
{
  "status": "ok"
}
```

#### 2. 解析视频

```http
POST /api/parse
Content-Type: application/json

{
  "url": "https://v.douyin.com/xxx"
}
```

#### 3. 获取视频地址

```http
POST /api/fetch-video-url
Content-Type: application/json

{
  "photo_id": "xxx",
  "platform": "kuaishou"
}
```

#### 4. 获取图集图片

```http
POST /api/fetch-atlas-images
Content-Type: application/json

{
  "photo_id": "xxx",
  "platform": "kuaishou"
}
```

#### 5. 下载视频

```http
GET /api/download?url=xxx
```

#### 6. 图片代理

```http
GET /api/proxy-image?url=xxx
```

#### 7. 获取视频信息

```http
GET /api/video/:id
```

#### 8. 获取记录列表

```http
GET /api/records
```

### 错误响应

```json
{
  "error": "错误信息"
}
```

| HTTP 状态码 | 说明 |
|------------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 📁 项目结构

```
video-parser/
├── api/
│   └── routes.go              # API 路由定义
├── cmd/
│   └── server/
│       └── main.go           # 服务入口
├── frontend/
│   ├── src/
│   │   ├── App.vue          # 根组件
│   │   └── main.js          # 入口文件
│   ├── vite.config.js       # Vite 配置
│   └── package.json         # 依赖配置
├── internal/
│   ├── config/              # 配置管理
│   ├── handler/             # 请求处理器
│   ├── model/               # 数据模型
│   ├── parser/              # 视频解析器
│   ├── repository/          # 数据访问层
│   ├── service/             # 业务逻辑层
│   └── utils/               # 工具函数
├── go.mod                   # Go 依赖
├── go.sum                   # 依赖校验
└── README.md                # 项目文档
```

---

## ⚙️ 配置说明

### 后端配置项

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| DB_HOST | localhost | 数据库地址 |
| DB_PORT | 5432 | 数据库端口 |
| DB_USER | postgres | 数据库用户 |
| DB_PASSWORD | postgres | 数据库密码 |
| DB_NAME | video_parser | 数据库名称 |
| DB_SSLMODE | disable | SSL 连接模式 |
| SERVER_PORT | 8080 | 服务端口 |
| APP_ENV | development | 运行环境 |

### 前端配置项

修改 `frontend/vite.config.js`:

```javascript
export default defineConfig({
  server: {
    port: 3000,           // 端口
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
```

---

## 🐳 部署建议

### 生产环境（systemd + Nginx）

#### 1. 构建项目

```bash
# 构建后端
go build -o video-parser-server ./cmd/server/main.go

# 构建前端
cd frontend && npm install && npm run build
```

#### 2. 创建 systemd 服务

```bash
# /etc/systemd/system/video-parser.service
[Unit]
Description=Video Parser API Server
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/video-parser
Environment=DB_HOST=localhost
Environment=DB_PORT=5432
Environment=DB_USER=postgres
Environment=DB_PASSWORD=your_password
Environment=DB_NAME=video_parser
Environment=SERVER_PORT=8080
ExecStart=/opt/video-parser/video-parser-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
# 启用并启动服务
systemctl daemon-reload
systemctl enable video-parser
systemctl start video-parser
```

#### 3. 配置 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### 4. 配置 HTTPS（Let's Encrypt）

```bash
# 安装 certbot
apt install certbot python3-certbot-nginx

# 申请证书并自动配置 Nginx
certbot --nginx -d your-domain.com --non-interactive --agree-tos --email your@email.com --redirect
```

### Docker 部署

```dockerfile
# 多阶段构建
FROM golang:1.20-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM node:18-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

FROM alpine:latest
WORKDIR /app
COPY --from=backend /app/server .
COPY --from=frontend /app/dist ./frontend/dist
EXPOSE 8080
CMD ["./server"]
```

---

## ❓ 常见问题

**Q: 启动失败，提示数据库连接失败？**

A: 请确认 PostgreSQL 已启动，且环境变量配置正确。

**Q: 前端无法访问后端 API？**

A: 检查 CORS 配置和代理设置。

**Q: 视频解析失败？**

A: 部分平台需要更新解析规则，可能是页面结构变化导致。

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 `git checkout -b feature/xxx`
3. 提交更改 `git commit -m 'Add xxx'`
4. 推送分支 `git push origin feature/xxx`
5. 打开 Pull Request

---

## 📄 许可证

本项目基于 [MIT](LICENSE) 许可证开源。

---

## 📝 更新日志

### 2026-06-14 安全修复

- **[Critical] 修复 SSRF 漏洞** — Download 接口添加 URL 域名校验错误检查
- **[Critical] 修复 DoS 漏洞** — 所有代理/下载接口添加 `io.LimitReader` 限制响应大小
- **[Critical] 数据库连接池** — 配置 MaxOpenConns=25, MaxIdleConns=10, ConnMaxLifetime=5min
- **[Critical] 移除硬编码 Cookie** — 小红书和米游社解析器不再使用伪造的 Cookie
- **[Critical] 错误日志记录** — `recordParse` 失败时记录日志而非静默忽略

### 2026-06-14 Bug 修复

- **[Bug] 修复抖音播放量字段映射** — 使用 `play_count` 替代错误的 `share_count`
- **[Bug] 修复 FetchVideoURL/FetchAtlasImages 平台限制** — 添加 `platform` 参数支持多平台
- **[Fix] 修正 go.mod 版本号** — `go 1.25.0` → `go 1.21`
- **[Fix] 移除 vite.config.js 内部域名泄露** — 删除 `allowedHosts` 配置

---

<div align="center">

**感谢使用**

[返回顶部](#video-parser)

</div>
