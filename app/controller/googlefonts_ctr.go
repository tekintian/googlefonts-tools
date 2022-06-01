package controller

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tekintian/googlefonts-tools/utils"
)

var GoogleFontsCtl = &googleFontsCtl{}

type googleFontsCtl struct {
	baseCtr
}

const (
	FontsDIr = "fonts"
	ZipDir   = "zip"
	FPrefix  = "https://fonts.gstatic.com/s/"
	FNameReg = `family=(.*?):`                                        //字体名称匹配正则
	FUrlReg  = `url\((.*?)\)`                                         //字体文件匹配正则
	FFUrlReg = `https:\/\/fonts\.gstatic\.com\/s\/(.*?)\/(.*?)\/(.*)` //文件目录,版本,文件名匹配
)

// zip文件下载
func (c *googleFontsCtl) ZipDownload(resp http.ResponseWriter, fileName, fontName string) {

	if fileName == "" {
		c.RespMsg(resp, "fileName不能为空,请提供你要下载的文件地址")
		return
	}

	zipName := fmt.Sprintf("%s.zip", fontName)
	resp.Header().Set("Content-Type", "application/zip")
	resp.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipName))
	f, err := os.Open(fileName)
	if err != nil {
		resp.Write([]byte(err.Error()))
		os.Exit(1)
	}
	defer f.Close()
	io.Copy(resp, f)
	//reFilename := strings.TrimPrefix(fileName,fmt.Sprintf("%s",utils.GetPwd()))
}

// 获取字体css
func (c *googleFontsCtl) Fetch(resp http.ResponseWriter, r *http.Request) {
	reqUrl := r.URL.Query().Get("url")
	if reqUrl == "" {
		c.RespMsg(resp, `
		<html>
		<head>
			<title>Googlefonts Related Resource Download Tools - By Tekin</title>
		</head>
		<body>
		<p>
		Googlefonts url不能为空! 请在地址栏添加url参数: ?url=谷歌字体调用地址 <br>
		示例:<a href="/?url=https://fonts.googleapis.com/css?family=Open+Sans:400,400i,600,600i,700,700i&display=swap&subset=cyrillic-ext">
		/?url=https://fonts.googleapis.com/css?family=Open+Sans:400,400i,600,600i,700,700i&display=swap&subset=cyrillic-ext</a>
		</p>
		</body>
		</html>
`)
		return
	}
	regFname := regexp.MustCompile(FNameReg)
	fontName := ""
	if strs := regFname.FindStringSubmatch(reqUrl); len(strs) > 1 {
		fontName = strings.ToLower(strs[1])
	}
	zipFilename := c.getZipFilename(reqUrl, fontName)
	if err := utils.IsExist(zipFilename); err == nil {
		c.ZipDownload(resp, zipFilename, fontName)
	}

	ua := r.Header.Get("User-Agent")
	ip := utils.GetClientIp(r)
	if ip == "::1" {
		ip = utils.GetFakeIp()
	}

	data, err := utils.HttpGet(reqUrl, ip, ua)
	if err != nil {
		c.RespMsg(resp, err.Error())
		return
	}
	if blocked := strings.Contains(string(data), "errors/robot.png"); blocked {
		c.RespMsg(resp, "当前请求被Google阻止!,请更换浏览器/后缀稍后重试!")
		return
	}

	fontCss := fmt.Sprintf("%s/%s/%s.css", utils.GetPwd(), "fonts", fontName)
	//创建并打开文件
	file, err := os.OpenFile(fontCss, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		c.RespMsg(resp, err.Error())
		return
	}
	defer file.Close()
	dataStr := string(data)
	//替换URL
	txt := strings.ReplaceAll(dataStr, FPrefix, "")

	//写入文件时，使用带缓存的 *Writer
	bw := bufio.NewWriter(file) //创建带缓存的writer
	bw.WriteString(txt)         //写入文件
	// bw.Write(data) //写入文件
	bw.Flush() // 情况缓存

	fuReg := regexp.MustCompile(FUrlReg)
	mfurls := fuReg.FindAllStringSubmatch(dataStr, -1)

	filesArr := make([]string, 0)
	filesArr = append(filesArr, fontCss)
	fdirName := ""
	for _, v := range mfurls {
		filename, fdir, err := c.fetchFont(v[1], ip, ua, fontName)
		if err != nil {
			c.Println(err.Error())
			continue
		}
		fdirName = fdir
		filesArr = append(filesArr, filename)
	}

	fmt.Println(fdirName)
	zipFilename = fmt.Sprintf("%s/%s/%s.zip", utils.GetPwd(), ZipDir, zipFilename)
	c.gzipFile(filesArr, zipFilename)

	c.ZipDownload(resp, zipFilename, fontName)
	//fmt.Fprintln(resp, "Googlefonts fetch ok! \n "+zipFilename)
}

// 获取zip压缩文件名
func (c *googleFontsCtl) getZipFilename(url, fontName string) string {
	return fmt.Sprintf("%s_%s", fontName, utils.Md5(url))
}

// 获取字体文件,返回网站文件名
func (c *googleFontsCtl) fetchFont(furl, ip, ua, fontName string) (filename string, fidr string, err error) {
	fdir := c.getFileDir(furl)
	time.Sleep(100 * time.Microsecond) //休眠100毫秒,防止被K
	data, err := utils.HttpGet(furl, ip, ua)
	if err != nil {
		c.Println(err.Error())
		return "", fdir, err
	}
	filename = fmt.Sprintf("%s/%s", utils.GetPwd(), fdir)
	c.saveFile(filename, data)

	return filename, fdir, err
}

func (c *googleFontsCtl) saveFile(filename string, data []byte) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		panic(err.Error()) //抛异常并退出
	}
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	bw := bufio.NewWriter(file) //创建带缓存的writer
	for _, v := range data {
		bw.WriteByte(v) //写入文件
	}
	// bw.Write(data) //写入文件
	bw.Flush() // 情况缓存
}

// 获取文件路径
func (c *googleFontsCtl) getFileDir(furl string) string {
	regFF := regexp.MustCompile(FFUrlReg)
	matches := regFF.FindAllStringSubmatch(furl, -1)
	if matches == nil {
		c.Println("获取文件存储目录失败,FURL正则匹配失败!")
		return ""
	}
	//获取存储文件 相对路径
	filename := fmt.Sprintf("%s/%s/%s/%s", FontsDIr, matches[0][1], matches[0][2], matches[0][3])

	utils.MustExist(filename)
	return filename

}

// 创建zip压缩文件
func (c *googleFontsCtl) gzipFile(srcfiles []string, dstFiilename string) error {
	// 创建：zip文件
	zipfile, _ := os.Create(dstFiilename)
	defer zipfile.Close()

	// 打开：zip文件
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	//循环添加文件到压缩包
	for _, v := range srcfiles {
		//创建头信息,注意这里每个文件都需要创建一次
		header := zip.FileHeader{}
		// 将文件添加到头信息中  这里的 v是文件的绝对地址, TrimPrefix是去除前面的前缀地址
		header.Name = strings.TrimPrefix(v, fmt.Sprintf("%s/%s", utils.GetPwd(), FontsDIr))
		// 设置：zip的文件压缩算法
		header.Method = zip.Deflate

		// 创建：压缩包头部信息
		writer, _ := archive.CreateHeader(&header)
		//打开要添加的文件 v
		file, _ := os.Open(v)
		defer file.Close()
		io.Copy(writer, file)
	}
	return nil
}
