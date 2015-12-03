package appfile

import (
	"fmt"
	"path/filepath"

	"github.com/kuuyee/otto-learn/appfile/detect"
)

// Default 在给定目录下生成一个默认的Appfile
//
// 作为决定applicaiton的名字，路径必须是绝对路径
func Defalt(dir string, det *detect.Config) (*File, error) {
	appName := filepath.Base(dir)
	fmt.Printf("[KuuYee]====> appName:", appName)
	return &File{}, nil
}
