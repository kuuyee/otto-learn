package detect

import (
	_ "fmt"
	"path/filepath"
)

// Config 配置文件的格式
type Config struct {
	Detectors []*Detector
}

// Merge merges another config into this one. This will modify this
// Config object. Detectors in c2 are tried after detectors in this
// Config. Conflicts are ignored as lower priority detectors, meaning that
// if two detectors are for type "go", both will be tried.
func (c *Config) Merge(c2 *Config) error {
	c.Detectors = append(c.Detectors, c2.Detectors...)
	return nil
}

// Detector is something that detects a single type.
type Detector struct {
	Type string
	File []string
}

// Detect 如果detector匹配给定的目录返回true
func (d *Detector) Detect(dir string) (bool, error) {
	for _, pattern := range d.File {
		//func Glob(pattern string) (matches []string, err error)
		//filepath.Glob函数返回所有匹配模式匹配字符串pattern的文件或者nil（如果没有匹配的文件）
		//pattern的语法和Match函数相同。pattern可以描述多层的名字，如/usr/*/bin/ed（假设路径分隔符是'/'）。
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return false, err
		}
		if len(matches) > 0 {
			return true, nil
		}
	}
	return false, nil
}
