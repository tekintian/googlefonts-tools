package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(txt string) string {
	o := md5.New()
	o.Write([]byte(txt))
	return hex.EncodeToString(o.Sum(nil))
}

func ShortSign(sign string) string {
	if len(sign) > 16 {
		return sign[len(sign)-16:]
	}
	return sign
}
