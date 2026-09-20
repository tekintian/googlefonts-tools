package templates

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed *.html
var templateFS embed.FS

var AppVer = "dev"

var tmpl *template.Template

type RecentItem struct {
	FontName  string
	OriginalURL string
	Sign      string
	CreatedAt string
}

type IndexData struct {
	AppVer       string
	RecentItems  []RecentItem
}

type ProgressData struct {
	AppVer    string
	FontName  string
	Sign      string
	Progress  int
	DoneFiles int
	TotalFiles int
}

type ResultData struct {
	AppVer        string
	FontName      string
	Sign          string
	URL           string
	Size          string
	Duration      string
	DownloadCount int
	CreatedAt     string
	ZipURL        string
	CSSLinkHref   string
	CSSFullURL    string
}

type ErrorData struct {
	AppVer   string
	FontName string
	Sign     string
	ErrorMsg string
}

type NotFoundData struct {
	AppVer string
	Sign   string
}

type RecentData struct {
	AppVer string
	Rows   template.HTML
}

func InitTemplates() {
	var err error
	tmpl, err = template.New("").ParseFS(templateFS, "*.html")
	if err != nil {
		log.Fatalf("模板解析失败: %v", err)
	}
}

func Render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("模板渲染失败 %s: %v", name, err)
	}
}