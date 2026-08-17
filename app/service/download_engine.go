package service

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
	"github.com/tekintian/googlefonts-tools/utils"
)

const (
	StorageDir = "storage"
	FontsDir   = "storage/fonts"
	CacheDir   = "storage/cache"
	ZipDir     = "storage/zip"
	DbDir      = "storage/db"
	FPrefix    = "https://fonts.gstatic.com/s/"

	FNameReg = `family=(.*?):`
	FUrlReg  = `url\((.*?)\)`
	FFUrlReg = `https:\/\/fonts\.gstatic\.com\/s\/(.*?)\/(.*?)\/(.*)`
)

var (
	regFName = regexp.MustCompile(FNameReg)
	regFUrl  = regexp.MustCompile(FUrlReg)
	regFFUrl = regexp.MustCompile(FFUrlReg)
)

type FontFile struct {
	URL      string
	DirName  string
	FileName string
}

type ProgressCallback func(progress model.TaskProgress)

type DownloadEngine struct {
	MaxConcurrency int
	MaxRetry       int
	RetryDelay     time.Duration
}

func NewDownloadEngine() *DownloadEngine {
	return &DownloadEngine{
		MaxConcurrency: 5,
		MaxRetry:       3,
		RetryDelay:     2 * time.Second,
	}
}

func (e *DownloadEngine) Download(task *model.Task, onProgress ProgressCallback) error {
	os.MkdirAll(FontsDir, 0755)
	os.MkdirAll(CacheDir, 0755)
	os.MkdirAll(ZipDir, 0755)

	ip := utils.GetFakeIp()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	fontName := e.parseFontName(task.URL)
	task.FontName = fontName

	data, err := utils.HttpGet(task.URL, ip, ua)
	if err != nil {
		return fmt.Errorf("fetch css error: %w", err)
	}

	if strings.Contains(string(data), "errors/robot.png") {
		return fmt.Errorf("request blocked by Google, please try again later")
	}

	fontCss := fmt.Sprintf("%s/%s/%s.css", utils.GetPwd(), CacheDir, fontName)
	if err := e.saveFile(fontCss, strings.ReplaceAll(string(data), FPrefix, "")); err != nil {
		return fmt.Errorf("save css error: %w", err)
	}

	fontFiles := e.parseFontURLs(string(data))
	totalFiles := len(fontFiles) + 1

	if onProgress != nil {
		onProgress(model.TaskProgress{
			Sign:       task.Sign,
			Status:     model.StatusRunning,
			Progress:   10,
			DoneFiles:  1,
			TotalFiles: totalFiles,
		})
	}

	downloadedFiles, err := e.downloadConcurrently(fontFiles, ip, ua, task.Sign, totalFiles, onProgress)
	if err != nil {
		return err
	}

	allFiles := make([]string, 0, len(downloadedFiles)+1)
	allFiles = append(allFiles, fontCss)
	allFiles = append(allFiles, downloadedFiles...)

	zipSign := fmt.Sprintf("%s_%s", fontName, task.Sign)
	zipPath := fmt.Sprintf("%s/%s/%s.zip", utils.GetPwd(), ZipDir, zipSign)

	if err := e.createZip(allFiles, zipPath); err != nil {
		return fmt.Errorf("create zip error: %w", err)
	}

	var zipSize int64
	if fi, err := os.Stat(zipPath); err == nil {
		zipSize = fi.Size()
	}

	task.ZipPath = zipPath
	task.ZipSize = zipSize

	return nil
}

func (e *DownloadEngine) parseFontName(url string) string {
	if strs := regFName.FindStringSubmatch(url); len(strs) > 1 {
		return strings.ToLower(strs[1])
	}
	return "unknown"
}

func (e *DownloadEngine) parseFontURLs(cssContent string) []FontFile {
	matches := regFUrl.FindAllStringSubmatch(cssContent, -1)
	var files []FontFile
	for _, m := range matches {
		if len(m) > 1 && strings.HasPrefix(m[1], "https://fonts.gstatic.com/s/") {
			ffMatches := regFFUrl.FindAllStringSubmatch(m[1], -1)
			if len(ffMatches) > 0 && len(ffMatches[0]) >= 4 {
				dirName := fmt.Sprintf("%s/%s/%s", FontsDir, ffMatches[0][1], ffMatches[0][2])
				files = append(files, FontFile{
					URL:      m[1],
					DirName:  dirName,
					FileName: ffMatches[0][3],
				})
			}
		}
	}
	return files
}

func (e *DownloadEngine) downloadConcurrently(fontFiles []FontFile, ip, ua, sign string, totalFiles int, onProgress ProgressCallback) ([]string, error) {
	var mu sync.Mutex
	var downloadedFiles []string
	var doneCount int
	var firstErr error

	sem := make(chan struct{}, e.MaxConcurrency)
	var wg sync.WaitGroup

	for i, ff := range fontFiles {
		wg.Add(1)
		go func(idx int, file FontFile) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			filename, err := e.downloadWithRetry(file, ip, ua)
			if err != nil {
				fmt.Printf("[WARN] download font file failed: %s, error: %v\n", file.URL, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			downloadedFiles = append(downloadedFiles, filename)
			doneCount++
			progress := 10 + (90 * (doneCount + 1) / totalFiles)
			if progress > 99 {
				progress = 99
			}
			if onProgress != nil {
				onProgress(model.TaskProgress{
					Sign:       sign,
					Status:     model.StatusRunning,
					Progress:   progress,
					DoneFiles:  doneCount + 1,
					TotalFiles: totalFiles,
				})
			}
			mu.Unlock()
		}(i, ff)
	}

	wg.Wait()
	return downloadedFiles, firstErr
}

func (e *DownloadEngine) downloadWithRetry(file FontFile, ip, ua string) (string, error) {
	fullPath := fmt.Sprintf("%s/%s/%s", utils.GetPwd(), file.DirName, file.FileName)
	utils.MustExist(fullPath)

	var data []byte
	var err error

	for attempt := 0; attempt <= e.MaxRetry; attempt++ {
		if attempt > 0 {
			time.Sleep(e.RetryDelay * time.Duration(attempt))
			fmt.Printf("[RETRY] attempt %d for %s\n", attempt, file.URL)
		}

		time.Sleep(100 * time.Millisecond)
		data, err = utils.HttpGet(file.URL, ip, ua)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", err
	}

	if err := e.saveFile(fullPath, string(data)); err != nil {
		return "", err
	}

	return fullPath, nil
}

func (e *DownloadEngine) saveFile(filename string, content string) error {
	utils.MustExist(filename)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	bw := bufio.NewWriter(file)
	bw.WriteString(content)
	return bw.Flush()
}

func (e *DownloadEngine) saveFileBytes(filename string, data []byte) error {
	utils.MustExist(filename)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	bw := bufio.NewWriter(file)
	for _, v := range data {
		bw.WriteByte(v)
	}
	return bw.Flush()
}

func (e *DownloadEngine) createZip(srcFiles []string, dstFilename string) error {
	utils.MustExist(dstFilename)

	zipfile, err := os.Create(dstFilename)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	pwd := utils.GetPwd()
	fontsPrefix := fmt.Sprintf("%s/%s/", pwd, FontsDir)
	cachePrefix := fmt.Sprintf("%s/%s/", pwd, CacheDir)

	for _, v := range srcFiles {
		header := zip.FileHeader{}
		relPath := v
		if strings.HasPrefix(v, fontsPrefix) {
			relPath = strings.TrimPrefix(v, fontsPrefix)
		} else if strings.HasPrefix(v, cachePrefix) {
			relPath = strings.TrimPrefix(v, cachePrefix)
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(&header)
		if err != nil {
			return err
		}

		file, err := os.Open(v)
		if err != nil {
			return err
		}
		io.Copy(writer, file)
		file.Close()
	}

	return nil
}

func (e *DownloadEngine) ServeZipFile(w http.ResponseWriter, zipPath, fontName string) error {
	f, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	zipName := fmt.Sprintf("%s.zip", fontName)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, zipName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))
	http.ServeContent(w, &http.Request{Method: "GET"}, zipName, fi.ModTime(), f)

	return nil
}
