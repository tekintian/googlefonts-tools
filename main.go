package main

import (
	"fmt"
	"net/http"

	"github.com/tekintian/googlefonts-tools/app/controller"
	"github.com/tekintian/googlefonts-tools/utils"
)

// 纯 go 原生 Googlefonts字体解析下载工具, 无任何第三方框架的依赖!
// @author tekintian@gmail.com
// @url  http://dev.yunnan.ws
func main() {
	//这里的顺序,需要在前面绑定fn
	http.HandleFunc("/", controller.GoogleFontsCtl.Fetch)
	//从配置文件获取端口信息 , 默认 8080
	port := utils.IniReadInt("config.ini", "server", "port", 8080)
	//监听并启动服务
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Server %d started!", port)
	}
}
