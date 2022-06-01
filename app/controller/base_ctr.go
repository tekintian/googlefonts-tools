package controller

import (
	"fmt"
	"net/http"
)

//基础控制器封装
var BaseCtr = &baseCtr{}

type baseCtr struct {
}

//打印信息
func (c *baseCtr) Println(msg ...interface{}) {
	fmt.Println(msg...)
}
func (c *baseCtr) RespMsg(resp http.ResponseWriter, msg ...interface{}) {
	fmt.Fprintln(resp, msg...)
	return
}
