package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DevDep 持有上游依赖的信息,这些信息被用在Dev函数，
// 目的是构建一个完整的开发环境.
type DevDep struct {

	// Files是被依赖创建和使用的文件列表.
	// 如果这些文件已经存在，那么DevDep将
	// 不会再调用，而是使用缓存的文件
	//
	// 所有的文件都保存在Context.CacheDir目录
	// 中. 相对路径将关联CacheDir目录. 如果文件
	// 不在CacheDir目录, 没有缓存发生, 日志将会
	// 记载
	Files []string `json:"files"`
}

// RelFiles把所有文件值生成为相对路径
func (d DevDep) RelFiles(dir string) error {
	for i, f := range d.Files {
		//如果路径已经是相对目录,直接忽略
		if !filepath.IsAbs(f) {
			continue
		}

		// 生成相对路径
		f, err := filepath.Rel(dir, f)
		if err != nil {
			return fmt.Errorf("不能改成相对路径: %s\n\n%s", d.Files[i], err)
		}

		d.Files[i] = f
	}
	return nil
}

// ReadDevDep 从磁盘读取一个DevDep
func ReadDevDep(path string) (*DevDep, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var result DevDep
	dec := json.NewDecoder(f)
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// WriteDevDep 向磁盘写入一个DevDep
func WriteDevDep(path string, dep *DevDep) error {
	// 格式化打印JSON数据，以便容易检查
	data, err := json.MarshalIndent(dep, "", "    ")
	if err != nil {
		return err
	}

	// 写入
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewReader(data))
	return err

}
