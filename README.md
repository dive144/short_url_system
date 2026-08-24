
# ShortURL - 短链接生成服务

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://golang.org)
[![Gin](https://img.shields.io/badge/Gin-v1.12-009688)](https://github.com/gin-gonic/gin)
[![Swagger](https://img.shields.io/badge/Swagger-UI-85EA2D)](https://swagger.io)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-blue)](LICENSE)

> 一个基于 Go + Gin + GORM 构建的短链接生成与访问服务。这是我的第一个 Go 练手项目，支持短链接生成、自定义短码、过期时间、访问计数，以及 Docker 一键部署。

---

## 目录

- [项目概述](#项目概述)
- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
  - [环境要求](#环境要求)
  - [本地运行](#本地运行)
  - [Docker Compose 部署](#docker-compose-部署)
- [API 文档](#api-文档)
  - [接口概览](#接口概览)
  - [创建短链接](#创建短链接)
  - [访问短链接](#访问短链接)
- [功能流程](#功能流程)
  - [创建短链接流程](#创建短链接流程)
  - [访问短链接流程](#访问短链接流程)
  - [短码生成算法](#短码生成算法)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [配置说明](#配置说明)
- [版本路线图](#版本路线图)

---

## 项目概述

用户可以将长链接转换为短链接，并通过访问短链接快速跳转到原始链接。v1.0 版本实现了最核心的两个功能：**生成短链接** + **302 重定向跳转**，并完成了 Docker 容器化部署。

### 核心业务流程

```
用户提交长链接 → 生成短码 → 保存到 MySQL → 返回短链接

用户访问短链接 → 查询数据库 → 302 重定向到原始 URL
```

---

## 功能特性

### ✅ 已实现

- **短链接生成** — 提交原始 URL，自动生成唯一短码（Base62 编码）
- **自定义短码** — 支持用户自定义短码（字母数字，1-8 位）
- **302 重定向** — 访问短链接自动跳转到原始 URL
- **过期时间控制** — 支持设置链接过期时间（相对时长或绝对时间），过期返回 410
- **访问计数** — 原子性统计每个短链接的点击次数（并发安全）
- **MySQL 存储** — 使用 GORM ORM + AutoMigrate 自动建表
- **Swagger 文档** — 内置 API 文档页面，支持在线调试
- **Docker Compose 一键部署** — MySQL + 应用服务容器化编排

### 📋 待实现

- **Redis 缓存** — 缓存热门短链接，减少 MySQL 查询压力
- **访问统计详情** — 记录 IP、User-Agent、Referer 等访问日志
- **CRUD 补全** — 查询 / 更新 / 删除短链接接口
- **限流保护** — 基于 IP 的接口限流
- **链路追踪与监控** — Prometheus + Grafana
- **Nginx 反向代理 + HTTPS** — 生产环境部署
- **前端管理页面** — 浏览器可视化创建和管理短链接

---

## 技术栈

| 技术 | 用途 |
|------|------|
| [Go](https://golang.org) + [Gin](https://github.com/gin-gonic/gin) | Web 框架 |
| [GORM](https://gorm.io) | ORM（MySQL 驱动） |
| [MySQL 8.0](https://www.mysql.com) | 数据库 |
| [Viper](https://github.com/spf13/viper) | 配置管理 |
| [swaggo/swag](https://github.com/swaggo/swag) | Swagger 文档自动生成 |
| [Docker Compose](https://docs.docker.com/compose/) | 容器编排 |

---

## 快速开始

### 环境要求

- Go 1.26+
- MySQL 8.0+
- （可选）Docker + Docker Compose

### 本地运行

```bash
# 1. 克隆仓库
git clone https://github.com/<your-username>/short_url.git
cd short_url

# 2. 创建数据库
mysql -u root -p
mysql> CREATE DATABASE short_url_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
mysql> exit

# 3. 修改数据库连接信息
#    编辑 config/config.docker.yaml，将 database.dsn 改为你的配置

# 4. 启动服务
go run cmd/main.go
```

### Docker Compose 部署

```bash
# 构建并启动所有服务（MySQL + 应用）
make run

# 查看运行状态
docker compose ps

# 查看应用日志
docker compose logs -f short_url

# 停止服务（保留数据库数据）
make down

# 停止并删除数据卷（⚠️ 会清空数据库）
make clean
```

或者直接使用 Docker Compose 命令：

```bash
# 构建并启动
docker compose up -d --build

# 停止
docker compose down

# 停止并删除数据卷
docker compose down -v --remove-orphans
```

服务启动后访问：
```
http://localhost:9090
```

> **注意**：首次启动 MySQL 会自动执行 `init.sql` 初始化脚本，包括设置 `mysql_native_password` 认证插件，解决 Navicat 等客户端连接 MySQL 8.0 时的 1251 认证错误。

---

## API 文档

启动服务后访问 Swagger UI：
```
http://localhost:9090/swagger/index.html
```

### 接口概览

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/shorten` | 创建短链接 |
| `GET` | `/s/{shorturl}` | 访问短链接（302 重定向） |
| `GET` | `/swagger/*any` | Swagger 文档页面 |
| `GET/POST` | `/test` | 连通性检查 |

---

### 创建短链接

```bash
curl -X POST http://localhost:9090/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://example.com/very-long-url-here"
  }'
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `original_url` | string | ✅ | 原始长链接 |
| `custom_code` | string | ❌ | 自定义短码（字母数字，1-8 位） |
| `expire_time` | string | ❌ | 相对过期时间，如 `"2h"`、`"7d"` |
| `expire_at` | string | ❌ | 绝对过期时间（ISO 8601，如 `"2026-12-31T23:59:59Z"`） |

**响应示例：**

```json
{
  "code": 201,
  "message": "短链接创建成功",
  "data": {
    "short_code": "abc123",
    "short_url": "http://localhost:9090/s/abc123",
    "original_url": "https://example.com/very-long-url-here",
    "create_at": "2026-06-04T12:00:00+08:00",
    "expire_at": null,
    "click_count": 0
  }
}
```

**错误响应：**

| 状态码 | 说明 |
|--------|------|
| `400` | 参数错误（如缺少 original_url） |
| `409` | 自定义短码已存在 |
| `500` | 服务器内部错误 |

---

### 访问短链接

```bash
# curl 跟随 302 重定向
curl -L http://localhost:9090/s/abc123

# 浏览器直接访问
# http://localhost:9090/s/abc123
```

| 状态码 | 说明 |
|--------|------|
| `302` | 重定向到原始 URL |
| `404` | 短链接不存在 |
| `410` | 短链接已过期 |
| `500` | 服务器内部错误 |

---

## 功能流程

### 创建短链接流程

```
用户提交长链接（POST /api/v1/shorten）
    ↓
解析请求体（ShouldBindJSON）
    ↓
┌─ 是否提供 custom_code？ ──┐
│         ↓                  │
│       是                   │ 否
│         ↓                  │
│  检查短码是否已存在         │
│    ↓       ↓               │
│  已存在   不存在            │
│    ↓       ↓               │
│ 返回409  继续              │
│                          ↓
│         ↓                  │
└─────────┴─────────────────┘
    ↓
先插入记录获取自增 ID（GORM Create）
    ↓
┌─ 是否自定义短码？ ──┐
│ 是 → 用自定义短码    │ 否 → Base62 编码(ID + 100000)
└─────────────────────┘
    ↓
返回 201 + 短链接信息
```

### 访问短链接流程

```
用户访问短链接（GET /s/:shorturl）
    ↓
从路径参数获取 short_code
    ↓
查询数据库
    ↓
┌─ 记录是否存在？ ──┐
│  ↓         ↓       │
│ 存在      不存在    │
│  ↓         ↓       │
│继续      返回 404   │
└─────────────────────┘
    ↓
检查是否过期
    ↓
┌─ 已过期？ ──┐
│  ↓     ↓     │
│ 是    否     │
│  ↓     ↓     │
│返回410 继续   │
└───────────────┘
    ↓
异步更新点击次数（atomic）
    ↓
302 重定向到原始 URL
```

### 短码生成算法

短码使用 **Base62 编码**（字符集：`0-9a-zA-Z`）生成：

1. 先插入记录，GORM 自动获取自增 ID
2. 对 `ID + 100000`（偏移量，保证短码至少 4 位）进行 Base62 编码
3. 将编码结果写入 `short_code` 字段

```
示例：ID=1  →  1 + 100000 = 100001  →  Base62(100001) = "q2T"
```

---

## 项目结构

```
short_url/
├── cmd/
│   ├── main.go              # 程序入口 + Swagger 全局注解
│   └── docs/                # Swagger 自动生成的文档文件
│       ├── docs.go
│       ├── swagger.json
│       └── swagger.yaml
├── config/
│   ├── config.go            # Viper 加载 YAML 配置
│   ├── config.yaml          # 应用配置（端口、数据库连接）
│   └── mysql.go             # MySQL 连接 + GORM AutoMigrate
├── global/
│   └── global.go            # 全局 *gorm.DB 实例
├── handler/
│   └── controller.go        # HTTP 处理器（业务逻辑）
├── model/
│   └── shorturl.go          # GORM 模型 + 请求 DTO
├── routers/
│   └── router.go            # Gin 路由注册 + Swagger 挂载
├── utils/
│   └── utils.go             # Base62 编码工具函数
├── Dockerfile               # 多阶段构建（golang:1.26-alpine）
├── compose.yml              # Docker Compose 编排（MySQL + App）
├── init.sql                 # MySQL 初始化脚本（认证 + 建库）
├── makefile                 # 常用命令
├── go.mod / go.sum          # Go 模块依赖
└── README.md
```

---

## 开发指南

### 更新 Swagger 文档

修改完 API 注解后，在项目根目录执行：

```bash
swag init --parseDependency --parseInternal -g cmd/main.go -o cmd/docs
```

### 本地编译

```bash
go build -o bin/short_url cmd/main.go
```

### Docker 构建

```bash
docker build -t short_url:v1.0 .
```

### Make 命令速查

| 命令 | 说明 |
|------|------|
| `make build` | 构建 Docker 镜像 |
| `make run` | 构建镜像 + Docker Compose 启动 |
| `make down` | 停止容器（保留数据卷） |
| `make clean` | 停止容器 + 删除数据卷 |

---

## 配置说明

编辑 `config/config.yaml`：

```yaml
app:
  name: short_url
  port: :9090           # 服务监听端口

database:
  dsn: root:123456@tcp(127.0.0.1:3306)/short_url_system?charset=utf8mb4&parseTime=True&loc=Local
```

> **注意**：默认密码 `123456` 仅适用于本地开发，生产环境请务必修改。

---

## 版本路线图

### v1.0（当前版本）
- ✅ 基础短链接生成 + 跳转
- ✅ Docker Compose 一键部署

### v1.1（计划）
- [ ] 补全 CRUD 接口（查询 / 更新 / 删除）
- [ ] 完善错误处理和日志
- [ ] 添加 Redis 缓存加速
- [ ] 实现过期数据自动清理

### v1.2（计划）
- [ ] 访问统计详情（IP、User-Agent、Referer）
- [ ] 短链接密码保护
- [ ] 批量创建短链接
- [ ] 短链接禁用 / 启用

### v2.0（远期规划）
- [ ] 用户系统（注册 / 登录 / 我的短链接）
- [ ] Nginx 反向代理 + HTTPS
- [ ] 自定义域名绑定
- [ ] 前端管理页面
- [ ] 数据分析和报表
- [ ] Prometheus + Grafana 监控

---

## 致谢

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM](https://gorm.io)
- [swaggo/swag](https://github.com/swaggo/swag)
- [Viper](https://github.com/spf13/viper)
>>>>>>> 96693ac (Modified the config configuration file)
