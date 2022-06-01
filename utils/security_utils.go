package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// 普通MD5加密
func Md5(txt string) string {
	o := md5.New()
	o.Write([]byte(txt))
	return hex.EncodeToString(o.Sum(nil))
}
