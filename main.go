package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tekintian/googlefonts-tools/app/controller"
	"github.com/tekintian/googlefonts-tools/app/model"
	"github.com/tekintian/googlefonts-tools/app/repository"
	"github.com/tekintian/googlefonts-tools/app/service"
	"github.com/tekintian/googlefonts-tools/utils"
	"github.com/tekintian/googlefonts-tools/utils/db"
)

const (
	AppName    = "GoogleFonts Download Tools"
	AppVersion = "2.0.0"
)

func main() {
	mode := flag.String("mode", "", "运行模式: server / download (也可用 -s / -d 简写)")
	downloadUrl := flag.String("d", "", "直接下载指定URL (等价于 -mode download -url <URL>)")
	serverMode := flag.Bool("s", false, "启动Web服务 (等价于 -mode server)")
	url := flag.String("url", "", "Google Fonts URL (download模式必填)")
	urlFile := flag.String("file", "", "包含URL列表的文件，每行一个URL (download模式批量下载)")
	output := flag.String("output", "", "输出目录 (默认为 storage/ 目录下)")
	concurrency := flag.Int("concurrency", 5, "并发下载数")
	retry := flag.Int("retry", 3, "下载失败重试次数")
	port := flag.Int("port", 0, "服务端口 (0则从配置文件读取)")
	workers := flag.Int("workers", 3, "异步任务Worker数量")
	configFile := flag.String("config", "storage/config.ini", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		return
	}

	ensureConfigFile(*configFile)

	if *downloadUrl != "" {
		*mode = "download"
		if *downloadUrl == "-" {
			*url = readURLFromStdin()
		} else {
			*url = *downloadUrl
		}
	} else if *serverMode {
		*mode = "server"
	}
	if *mode == "" {
		*mode = "server"
	}

	if *port == 0 {
		*port = envOrInt("GF_SERVER_PORT", utils.IniReadInt(*configFile, "server", "port", 8000))
	}

	driverName := envOr("GF_DB_DRIVER", utils.IniReadString(*configFile, "database", "driver", "sqlite"))
	dsn := envOr("GF_DB_DSN", utils.IniReadString(*configFile, "database", "dsn", ""))

	if err := initDatabase(driverName, dsn); err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewRepo(db.DB)
	if err := repo.Init(); err != nil {
		fmt.Printf("数据库表初始化失败: %v\n", err)
		os.Exit(1)
	}

	engine := service.NewDownloadEngine()
	engine.MaxConcurrency = *concurrency
	engine.MaxRetry = *retry

	notifier := initNotifier(*configFile)

	controller.AppVer = AppVersion

	serverHost := envOr("GF_SERVER_HOST", utils.IniReadString(*configFile, "server", "host", "localhost"))

	switch *mode {
	case "download":
		runCLIMode(*url, *urlFile, *output, engine, repo, notifier)
	case "server":
		runServerMode(serverHost, *port, *workers, engine, repo, notifier)
	default:
		fmt.Printf("未知模式: %s (支持: server, download)\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

func initDatabase(driverName, dsn string) error {
	if driverName == "sqlite" && dsn == "" {
		dsn = "storage/db/googlefonts.db"
		os.MkdirAll("storage/db", 0755)
	}
	fmt.Printf("[DB] 初始化数据库: driver=%s\n", driverName)
	return db.Init(driverName, dsn)
}

func initNotifier(configFile string) *service.NotifyDispatcher {
	dispatcher := service.NewNotifyDispatcher()

	dingtalkWebhook := envOr("GF_DINGTALK_WEBHOOK", utils.IniReadString(configFile, "notify", "dingtalk_webhook", ""))
	if dingtalkWebhook != "" {
		dispatcher.Add(service.NewDingTalkNotifier(dingtalkWebhook))
		fmt.Println("[Notify] 钉钉通知已启用")
	}

	wechatWebhook := envOr("GF_WECHAT_WEBHOOK", utils.IniReadString(configFile, "notify", "wechat_webhook", ""))
	if wechatWebhook != "" {
		dispatcher.Add(service.NewWeChatNotifier(wechatWebhook))
		fmt.Println("[Notify] 微信通知已启用")
	}

	smtpHost := envOr("GF_SMTP_HOST", utils.IniReadString(configFile, "notify", "smtp_host", ""))
	smtpPort := envOrInt("GF_SMTP_PORT", utils.IniReadInt(configFile, "notify", "smtp_port", 0))
	smtpFrom := envOr("GF_SMTP_FROM", utils.IniReadString(configFile, "notify", "smtp_from", ""))
	smtpPassword := envOr("GF_SMTP_PASSWORD", utils.IniReadString(configFile, "notify", "smtp_password", ""))
	smtpTo := envOr("GF_SMTP_TO", utils.IniReadString(configFile, "notify", "smtp_to", ""))
	if smtpHost != "" && smtpFrom != "" && smtpTo != "" {
		dispatcher.Add(service.NewEmailNotifier(smtpHost, smtpPort, smtpFrom, smtpPassword, smtpTo))
		fmt.Println("[Notify] 邮件通知已启用")
	}

	return dispatcher
}

func runCLIMode(url, urlFile, output string, engine *service.DownloadEngine, repo repository.TaskRepository, notifier *service.NotifyDispatcher) {
	fmt.Printf("%s v%s - CLI模式\n", AppName, AppVersion)

	var urls []string
	if url != "" {
		urls = append(urls, url)
	}

	if len(urls) == 0 && urlFile == "" {
		fmt.Println("请输入 Google Fonts URL (支持粘贴，无需加引号):")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			urls = append(urls, input)
		}
	}
	if urlFile != "" {
		data, err := os.ReadFile(urlFile)
		if err != nil {
			fmt.Printf("读取URL文件失败: %v\n", err)
			os.Exit(1)
		}
		for _, line := range splitLines(string(data)) {
			line = trimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
	}

	if len(urls) == 0 {
		fmt.Println("错误: 请通过 -url 或 -file 参数提供Google Fonts URL")
		flag.Usage()
		os.Exit(1)
	}

	if output != "" {
		os.MkdirAll(output, 0755)
	}

	tm := service.NewTaskManager(repo, engine, notifier, 1)

	for i, u := range urls {
		fmt.Printf("\n[%d/%d] 下载: %s\n", i+1, len(urls), u)
		task, err := tm.SubmitSync(u, "")
		if err != nil {
			fmt.Printf("  ❌ 失败: %v\n", err)
			continue
		}

		if task.Status == model.StatusSuccess {
			fmt.Printf("  ✅ 完成: %s (%s)\n", task.ZipPath, formatSize(task.ZipSize))
			if output != "" {
				if err := copyToOutput(task.ZipPath, output); err == nil {
					fmt.Printf("  📁 已复制到: %s\n", output)
				}
			}
		} else {
			fmt.Printf("  ❌ 失败: %s\n", task.ErrorMsg)
		}
	}

	fmt.Printf("\n全部完成!\n")
}

func runServerMode(host string, port, workers int, engine *service.DownloadEngine, repo repository.TaskRepository, notifier *service.NotifyDispatcher) {
	fmt.Printf("%s v%s - Server模式\n", AppName, AppVersion)

	tm := service.NewTaskManager(repo, engine, notifier, workers)
	tm.Start()

	router := controller.NewRouter()
	router.Setup(engine)

	os.MkdirAll("storage/fonts", 0755)
	os.MkdirAll("storage/cache", 0755)
	os.MkdirAll("storage/zip", 0755)
	os.MkdirAll("storage/db", 0755)

	fmt.Printf("[Server] 监听端口: %d\n", port)
	fmt.Printf("[Server] 访问地址: http://%s:%d\n", host, port)
	fmt.Printf("[Server] API文档: http://%s:%d/api/v1/tasks\n", host, port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		panic(err)
	}
}

func readURLFromStdin() string {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err == nil {
			result := strings.TrimSpace(string(data))
			if result != "" {
				return result
			}
		}
	}
	fmt.Println("请输入 Google Fonts URL (支持粘贴，无需加引号):")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

func copyToOutput(src, destDir string) error {
	filename := filepath.Base(src)
	dst := filepath.Join(destDir, filename)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func ensureConfigFile(path string) {
	if _, err := os.Stat(path); err == nil {
		return
	}
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	defaultConfig := `[server]
host=localhost
port=8000

[database]
driver=sqlite
dsn=

[notify]
dingtalk_webhook=
wechat_webhook=
smtp_host=
smtp_port=25
smtp_from=
smtp_password=
smtp_to=
`
	if err := os.WriteFile(path, []byte(defaultConfig), 0644); err == nil {
		fmt.Printf("[Config] 已生成默认配置文件: %s\n", path)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
