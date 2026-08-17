# Google Fonts Download Tools

纯 Go 实现的 Google Fonts 字体解析下载工具，支持命令行下载和 Web 服务模式，异步任务调度、实时进度推送、多通道通知。

> Author: [Tekin](https://ai.tekin.cn/) | Github: [tekintian/googlefonts-tools](https://github.com/tekintian/googlefonts-tools) | 在线体验: [gf.tekin.cn](https://gf.tekin.cn)

## 功能特性

- **命令行下载** — 一条命令直接下载字体 ZIP 包
- **Web 服务** — 提交任务后生成永久链接，随时随地查看状态和下载结果
- **异步任务调度** — Goroutine Worker 池 + Channel 并发处理
- **幂等性设计** — 基于 URL 的 MD5 签名，同一 URL 不重复下载，已有结果直接返回
- **SSE 实时进度** — Server-Sent Events 推送下载进度，前端实时更新
- **多通道通知** — 支持钉钉、微信、邮件、Webhook 通知
- **多数据库支持** — SQLite（默认）/ MySQL / PostgreSQL，Repository 模式抽象
- **并发下载** — 带信号量的 Goroutine 池，可配置并发数和重试次数

## 快速开始

### 构建

```bash
go mod tidy

# 默认构建（仅 SQLite，约 11MB）
go build -ldflags="-s -w" -o googlefonts-tools .

# 启用 MySQL 支持（约 11MB）
go build -tags mysql -ldflags="-s -w" -o googlefonts-tools .

# 启用 PostgreSQL 支持（约 15MB）
go build -tags postgres -ldflags="-s -w" -o googlefonts-tools .

# 全部启用（约 15MB）
go build -tags "mysql,postgres" -ldflags="-s -w" -o googlefonts-tools .
```

### 命令行下载

```bash
# 最简写法：-d 直接下载指定 URL
./googlefonts-tools -d "https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap"

# 批量下载（从文件读取 URL 列表）
./googlefonts-tools -d -file urls.txt

# 指定输出目录
./googlefonts-tools -d "https://fonts.googleapis.com/css?family=Roboto:wght@400;700" -output ./my-fonts

# 完整写法
./googlefonts-tools -mode download -url "https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap"
```

> ⚠️ **URL 中的 `;` 和 `&` 是 Shell 特殊字符！** 有三种方式避免 Shell 解析：
>
> 1. **引号包裹**（推荐）：`./gf -d "https://fonts...wght@300;400&display=swap"`
> 2. **交互式输入**：直接运行 `./gf -d`，然后粘贴 URL（无需引号）
> 3. **管道输入**：`echo "https://fonts...wght@300;400&display=swap" | ./gf -d -`

### Web 服务模式

```bash
# 最简写法：-s 启动服务
./googlefonts-tools -s

# 指定端口
./googlefonts-tools -s -port 9000

# 完整写法
./googlefonts-tools -mode server -port 8000 -workers 3
```

启动后访问 `http://localhost:8000`，在页面输入 Google Fonts URL 提交下载任务。

## 命令行参数

| 参数 | 简写 | 说明 |
|------|------|------|
| `-mode` | | 运行模式：`server` / `download` |
| | `-d` | 直接下载指定 URL（等价于 `-mode download -url <URL>`） |
| | `-s` | 启动 Web 服务（等价于 `-mode server`） |
| `-url` | | Google Fonts URL（download 模式必填） |
| `-file` | | URL 列表文件，每行一个 URL（批量下载） |
| `-output` | | 输出目录（默认 `storage/` 下） |
| `-concurrency` | | 并发下载数（默认 5） |
| `-retry` | | 下载失败重试次数（默认 3） |
| `-port` | | 服务端口（默认从配置文件读取，fallback 8000） |
| `-workers` | | 异步任务 Worker 数量（默认 3） |
| `-config` | | 配置文件路径（默认 `storage/config.ini`） |
| `-version` | | 显示版本信息 |

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/tasks` | 创建下载任务（`url=xxx` form 提交） |
| GET | `/api/v1/tasks` | 列出所有任务 |
| GET | `/api/v1/tasks/{sign}` | 查询任务状态 |
| GET | `/api/v1/tasks/{sign}/progress` | SSE 实时进度推送 |
| GET | `/d/{sign}` | 永久链接 — 任务状态页 |
| GET | `/d/{sign}/download` | 永久链接 — 下载 ZIP 文件 |
| GET | `/d/{sign}/progress` | 永久链接 — SSE 进度推送 |

### 创建任务示例

```bash
curl -X POST http://localhost:8000/api/v1/tasks \
  -d "url=https://fonts.googleapis.com/css2?family=Inter:wght@400;700&display=swap"
```

返回：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "sign": "a1b2c3d4e5f6...",
    "font_name": "Inter",
    "status": "pending",
    "permalink": "/d/a1b2c3d4e5f6...",
    "download_url": "/d/a1b2c3d4e5f6.../download"
  }
}
```

## 目录结构

```
storage/
├── config.ini      # 配置文件（首次运行自动生成）
├── db/             # SQLite 数据库文件 (googlefonts.db)
├── cache/          # 字体 CSS 缓存 ({fontName}.css)
├── fonts/          # 下载的字体原始文件 ({fontName}/{version}/{file})
└── zip/            # 生成的 ZIP 打包文件 ({fontName}_{sign}.zip)
```

## 配置文件

`storage/config.ini`（首次运行自动生成，也可手动创建）：

```ini
[server]
host=${GF_SERVER_HOST:-localhost}
port=${GF_SERVER_PORT:-8000}

[database]
driver=${GF_DB_DRIVER:-sqlite}
dsn=${GF_DB_DSN:-}

[notify]
dingtalk_webhook=${GF_DINGTALK_WEBHOOK:-}
wechat_webhook=${GF_WECHAT_WEBHOOK:-}
smtp_host=${GF_SMTP_HOST:-}
smtp_port=${GF_SMTP_PORT:-25}
smtp_from=${GF_SMTP_FROM:-}
smtp_password=${GF_SMTP_PASSWORD:-}
smtp_to=${GF_SMTP_TO:-}
```

> 配置文件支持 `${VAR:-default}` 语法，Docker 容器启动时通过 `envsubst` 自动替换为环境变量值。Go 程序也直接读取 `GF_*` 环境变量，优先级：**环境变量 > 配置文件 > 默认值**。

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `GF_SERVER_HOST` | 服务主机名（用于显示访问地址） | `localhost` |
| `GF_SERVER_PORT` | 服务端口 | `8000` |
| `GF_DB_DRIVER` | 数据库驱动 | `sqlite` |
| `GF_DB_DSN` | 数据库连接串 | 空 |
| `GF_DINGTALK_WEBHOOK` | 钉钉 Webhook | 空 |
| `GF_WECHAT_WEBHOOK` | 微信 Webhook | 空 |
| `GF_SMTP_HOST` | SMTP 服务器 | 空 |
| `GF_SMTP_PORT` | SMTP 端口 | `25` |
| `GF_SMTP_FROM` | 发件人 | 空 |
| `GF_SMTP_PASSWORD` | 发件人密码 | 空 |
| `GF_SMTP_TO` | 收件人 | 空 |

### 数据库配置

- **SQLite**（默认）：`driver=sqlite`，无需额外配置
- **MySQL**：`driver=mysql`，`dsn=user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4`
- **PostgreSQL**：`driver=postgres`，`dsn=postgres://user:password@127.0.0.1:5432/dbname`

## 架构

```
┌─────────────┐     ┌──────────────┐     ┌───────────────┐     ┌────────────┐
│  Controller  │ ──▶ │   Service    │ ──▶ │  Repository   │ ──▶ │   Model    │
│  (HTTP/API)  │     │ (TaskMgr/   │     │  (SQLite/     │     │  (Task/    │
│              │     │  Download/   │     │   MySQL/PG)   │     │  Progress) │
│              │     │  Notify)     │     │               │     │            │
└─────────────┘     └──────────────┘     └───────────────┘     └────────────┘
```

四层分层架构：Controller → Service → Repository → Model

## Docker

```bash
# 快速启动（数据持久化到宿主机）
docker run -d -p 8000:8000 -v ./gf-data:/app/storage ghcr.io/tekintian/googlefonts-tools:latest

# 通过环境变量配置（推荐，无需编辑配置文件）
docker run -d -p 8000:8000 -v ./gf-data:/app/storage \
  -e GF_SERVER_HOST=gf.tekin.cn \
  -e GF_SERVER_PORT=8000 \
  -e GF_DB_DRIVER=sqlite \
  -e GF_DINGTALK_WEBHOOK=https://oapi.dingtalk.com/robot/send?access_token=xxx \
  ghcr.io/tekintian/googlefonts-tools:latest

# 使用 MySQL + 邮件通知
docker run -d -p 8000:8000 -v ./gf-data:/app/storage \
  -e GF_DB_DRIVER=mysql \
  -e GF_DB_DSN="user:pass@tcp(127.0.0.1:3306)/fonts?charset=utf8mb4" \
  -e GF_SMTP_HOST=smtp.example.com \
  -e GF_SMTP_PORT=465 \
  -e GF_SMTP_FROM=noreply@example.com \
  -e GF_SMTP_PASSWORD=secret \
  -e GF_SMTP_TO=admin@example.com \
  ghcr.io/tekintian/googlefonts-tools:latest

# 挂载自定义 config.ini（支持 ${VAR:-default} 模板语法）
mkdir -p gf-data
cat > gf-data/config.ini << 'EOF'
[server]
port=${GF_SERVER_PORT:-8000}
[database]
driver=sqlite
[notify]
dingtalk_webhook=${GF_DINGTALK_WEBHOOK:-}
EOF
docker run -d -p 8000:8000 -v ./gf-data:/app/storage ghcr.io/tekintian/googlefonts-tools:latest
```

> 容器内 `/app/storage` 是数据目录，包含 `config.ini`、数据库、缓存、字体文件和 ZIP 包。挂载此目录即可持久化所有数据。镜像内置 [envsubst](https://github.com/tekintian/envsubst)（从源码编译），启动时自动将配置文件中的 `${VAR:-default}` 替换为环境变量值。

## 注意事项

- 访问过于频繁可能被 Google 限制，工具内置了随机 IP 和 UA 以及请求间隔来缓解
- 如需更高频率访问，可考虑使用 [chromedp](https://github.com/chromedp/chromedp) 方案
- SQLite 模式下数据库连接池限制为 1（SQLite 单写限制），MySQL/PostgreSQL 默认 25 连接

## License

MIT