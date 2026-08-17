package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
)

type Notifier interface {
	Send(task *model.Task) error
	Name() string
}

type NotifyDispatcher struct {
	notifiers []Notifier
}

func NewNotifyDispatcher() *NotifyDispatcher {
	return &NotifyDispatcher{}
}

func (d *NotifyDispatcher) Add(notifier Notifier) {
	d.notifiers = append(d.notifiers, notifier)
}

func (d *NotifyDispatcher) Dispatch(task *model.Task) {
	for _, n := range d.notifiers {
		go func(notifier Notifier) {
			if err := notifier.Send(task); err != nil {
				fmt.Printf("[Notify] %s send error: %v\n", notifier.Name(), err)
			} else {
				fmt.Printf("[Notify] %s sent successfully for task: %s\n", notifier.Name(), task.Sign)
			}
		}(n)
	}

	if task.NotifyConfig != "" {
		go d.dispatchCustom(task)
	}
}

func (d *NotifyDispatcher) dispatchCustom(task *model.Task) {
	webhooks := strings.Split(task.NotifyConfig, ",")
	for _, wh := range webhooks {
		wh = strings.TrimSpace(wh)
		if strings.HasPrefix(wh, "http") {
			d.sendWebhook(wh, task)
		}
	}
}

func (d *NotifyDispatcher) sendWebhook(url string, task *model.Task) {
	payload := map[string]interface{}{
		"sign":       task.Sign,
		"font_name":  task.FontName,
		"status":     task.Status,
		"progress":   task.Progress,
		"zip_size":   task.ZipSize,
		"created_at": task.CreatedAt,
		"zip_path":   task.ZipPath,
	}
	data, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("[Notify] webhook error: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

type DingTalkNotifier struct {
	WebhookURL string
}

func NewDingTalkNotifier(webhookURL string) *DingTalkNotifier {
	return &DingTalkNotifier{WebhookURL: webhookURL}
}

func (n *DingTalkNotifier) Name() string { return "dingtalk" }

func (n *DingTalkNotifier) Send(task *model.Task) error {
	statusEmoji := "✅"
	if task.Status == model.StatusFailed {
		statusEmoji = "❌"
	}

	text := fmt.Sprintf("## %s 字体下载%s\n\n", task.FontName, statusEmoji)
	text += fmt.Sprintf("- **状态**: %s\n", task.Status)
	text += fmt.Sprintf("- **签名**: %s\n", task.Sign)

	if task.Status == model.StatusSuccess {
		text += fmt.Sprintf("- **文件大小**: %s\n", formatSize(task.ZipSize))
		if task.CompletedAt != nil {
			text += fmt.Sprintf("- **耗时**: %s\n", task.CompletedAt.Sub(task.CreatedAt).Round(time.Second))
		}
	}
	if task.Status == model.StatusFailed {
		text += fmt.Sprintf("- **错误**: %s\n", task.ErrorMsg)
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": fmt.Sprintf("字体下载%s", statusEmoji),
			"text":  text,
		},
	}
	return n.postJSON(payload)
}

func (n *DingTalkNotifier) postJSON(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(n.WebhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type WeChatNotifier struct {
	WebhookURL string
}

func NewWeChatNotifier(webhookURL string) *WeChatNotifier {
	return &WeChatNotifier{WebhookURL: webhookURL}
}

func (n *WeChatNotifier) Name() string { return "wechat" }

func (n *WeChatNotifier) Send(task *model.Task) error {
	statusEmoji := "✅"
	if task.Status == model.StatusFailed {
		statusEmoji = "❌"
	}

	content := fmt.Sprintf("**字体下载通知** %s\n", statusEmoji)
	content += fmt.Sprintf("字体: %s\n", task.FontName)
	content += fmt.Sprintf("状态: %s\n", task.Status)
	content += fmt.Sprintf("签名: %s\n", task.Sign)
	if task.Status == model.StatusSuccess {
		content += fmt.Sprintf("文件大小: %s\n", formatSize(task.ZipSize))
	}
	if task.Status == model.StatusFailed {
		content += fmt.Sprintf("错误: %s\n", task.ErrorMsg)
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": content},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(n.WebhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type EmailNotifier struct {
	Host     string
	Port     int
	From     string
	Password string
	To       string
}

func NewEmailNotifier(host string, port int, from, password, to string) *EmailNotifier {
	return &EmailNotifier{Host: host, Port: port, From: from, Password: password, To: to}
}

func (n *EmailNotifier) Name() string { return "email" }

func (n *EmailNotifier) Send(task *model.Task) error {
	subject := fmt.Sprintf("字体下载通知: %s - %s", task.FontName, task.Status)

	body := fmt.Sprintf("字体: %s\n状态: %s\n签名: %s\n", task.FontName, task.Status, task.Sign)
	if task.Status == model.StatusSuccess {
		body += fmt.Sprintf("文件大小: %s\n", formatSize(task.ZipSize))
	}
	if task.Status == model.StatusFailed {
		body += fmt.Sprintf("错误: %s\n", task.ErrorMsg)
	}

	fromAddr := mail.Address{Name: "GoogleFonts Tools", Address: n.From}
	toAddr := mail.Address{Name: "", Address: n.To}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		fromAddr.String(), toAddr.String(), subject, body)

	auth := smtp.PlainAuth("", n.From, n.Password, n.Host)
	addr := fmt.Sprintf("%s:%d", n.Host, n.Port)

	return smtp.SendMail(addr, auth, n.From, []string{n.To}, []byte(msg))
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
