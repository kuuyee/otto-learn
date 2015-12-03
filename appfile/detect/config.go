package detect

import (
	_ "fmt"
	_ "path/filepath"
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
