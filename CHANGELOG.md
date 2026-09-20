# Changelog

## v2.3.0 (2026-09-20)

### 新增
- 自托管 CSS 生成：自动将 Google Fonts CSS 中的绝对 URL 替换为本地相对路径（`../{version}/xxx.woff2`），摆脱 Google Fonts CDN 依赖
- 静态目录直出：下载文件按 `storage/c/{fontName}/` 组织，Nginx 只需配置 `/c/` 即可服务全部静态资源
- HTTP 缓存：静态文件（CSS、字体、ZIP）支持 `Cache-Control: public, max-age=31536000, immutable`
- 协议相对 URL：自托管 `<link>` 引用使用 `//host/path` 格式，浏览器自动匹配 HTTP/HTTPS
- 最近下载页面（`/recent`）：展示最近 50 条下载记录，含 Google 原始 URL 方便溯源
- 下载完成页面 UI 重构：自托管引用展示 `<link>` 标签，CSS 地址可点击新窗口打开

### 优化
- 文件路径精简：移除字体名称重复目录，路径从 `/c/{fontName}/fonts/{fontName}/{version}/` 精简为 `/c/{fontName}/{version}/`
- ZIP 下载路径归入 `storage/c/d/` 子目录，避免根目录文件混乱
- 签名精简：URL 签名取 MD5 后 16 位（`ShortSign`），避免 32 位过长
- 同名字体文件共享：不同子集的 CSS 独立存储于 `{sign16}/` 目录，字体文件共享 `{version}/` 目录
- 幂等性设计：基于 URL 的 MD5 签名确保资源唯一性，同一 URL 不重复下载

### 修复
- 修复 `self-host css not found` 错误：CSS 文件读取路径与保存路径不一致
- 修复 `GetPwd()` 返回临时目录导致文件丢失：改用 `os.Getwd()` 获取当前工作目录
- 修复自托管 CSS 中字体 URL 替换后多余右括号问题
- 修复 `go vet` 报错：模板参数顺序与实际参数传递顺序不匹配

### 目录结构变更
```
# 旧结构
storage/
├── fonts/          # 字体文件
└── zip/            # ZIP 包

# 新结构
storage/
└── c/              # 静态内容目录（Nginx 配置 /c/ 即可）
    ├── d/                      # ZIP 下载包
    │   └── {fontName}_{sign16}.zip
    └── {fontName}/              # 如 inter/
        ├── {version}/           # 共享字体文件（如 v24/xxx.woff2）
        └── {sign16}/            # URL 签名后16位，不同子集 CSS 各自独立
            └── {fontName}.css   # 引用 ../{version}/xxx.woff2 相对路径
```

## v2.2.0

### 新增
- Web 服务模式：提交任务后生成永久链接，随时随地查看状态和下载结果
- 异步任务调度：Goroutine Worker 池 + Channel 并发处理
- SSE 实时进度推送
- 多通道通知：支持钉钉、微信、邮件通知
- 多数据库支持：SQLite / MySQL / PostgreSQL

## v2.1.0

### 新增
- 命令行下载模式
- 批量下载（从文件读取 URL 列表）
- 并发下载（带信号量的 Goroutine 池）

## v2.0.0

### 新增
- 纯 Go 重写，四层分层架构：Controller → Service → Repository → Model
- Docker 支持（多平台构建、GHCR 发布）
- GitHub Actions 自动发布