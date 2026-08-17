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
go build -o googlefonts-tools .
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
| `-config` | | 配置文件路径（默认 `config.ini`） |
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
├── db/              # SQLite 数据库文件 (googlefonts.db)
├── cache/           # 字体 CSS 缓存 ({fontName}.css)
├── fonts/           # 下载的字体原始文件 ({fontName}/{version}/{file})
└── zip/             # 生成的 ZIP 打包文件 ({fontName}_{sign}.zip)
```

## 配置文件

`config.ini`：

```ini
[server]
port=8000

[database]
driver=sqlite
; dsn=storage/db/googlefonts.db

[notify]
dingtalk_webhook=
wechat_webhook=
smtp_host=
smtp_port=25
smtp_from=
smtp_password=
smtp_to=
```

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

## 注意事项

- 访问过于频繁可能被 Google 限制，工具内置了随机 IP 和 UA 以及请求间隔来缓解
- 如需更高频率访问，可考虑使用 [chromedp](https://github.com/chromedp/chromedp) 方案
- SQLite 模式下数据库连接池限制为 1（SQLite 单写限制），MySQL/PostgreSQL 默认 25 连接

## License

MIT