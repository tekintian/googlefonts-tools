package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func GetPwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //当前程序运行目录获取
	if err != nil {
		return ""
	}
	return dir
}

// 判断所给路径是否存在 返回error 如果error为nil表示已经存在
func IsExist(path string) error {
	//如果非 / 开头，自动增加当前程序运行目录
	if len(path) >= 1 && path[0:1] != "/" {
		path = GetPwd() + "/" + path
	}
	_, err := os.Stat(path)
	return err
}

// 判断所给路径是否存在，如果不存在则自动创建对应目录
func MustExist(path string) error {
	err := IsExist(path)
	if err != nil {
		fdir := filepath.Dir(path)           //获取目录
		if err = IsExist(fdir); err != nil { //如果目录不存在，则自动创建目录
			if err := os.MkdirAll(fdir, 0777); err != nil { //os.ModePerm
				log.Println(err.Error())
				return err
			}
		}
	}
	return err
}

func YmDirStr(dt time.Time) string {
	if dt.IsZero() {
		dt = time.Now()
	}
	return fmt.Sprintf("%s/%s", dt.Format("Y"), dt.Format("m"))
}

func YmdDirStr(dt time.Time) string {
	if dt.IsZero() {
		dt = time.Now()
	}
	return fmt.Sprintf("%s/%s/%s", dt.Format("Y"), dt.Format("m"), dt.Format("d"))
}
