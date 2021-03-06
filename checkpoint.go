package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/go-checkpoint"
	"github.com/kuuyee/otto-learn/command"
)

func init() {
	checkpointResult = make(chan *checkpoint.CheckResponse, 1)
}

var checkpointResult chan *checkpoint.CheckResponse

// runCheckponit 运行一个HashiCorp Checkpoint请求。
func runCheckpoint(c *Config) {
	// 如果用户完全不想checkpoint，那么return
	if c.DisableCheckpoint {
		checkpointResult <- nil
		return
	}

	configDir, err := ConfigDir()
	if err != nil {
		log.Printf("[ERR] Checkpoint 设置报错：%s", err)
		checkpointResult <- nil
		return
	}

	version := Version

	if VersionPrerelease != "" {
		version += fmt.Sprintf("-%s", VersionPrerelease)
	}

	signaturePath := filepath.Join(configDir, "checkpoint_signature")
	if c.DisableCheckpointSignature {
		log.Printf("[INFO] Checkpoint signature 不可用")
		signaturePath = ""
	}

	resp, err := checkpoint.Check(&checkpoint.CheckParams{
		Product:       "otto",
		Version:       version,
		SignatureFile: signaturePath,
		CacheFile:     filepath.Join(configDir, "checkpoint_cache"),
	})

	if err != nil {
		log.Printf("[ERR] Checkpoint error: %s", err)
		resp = nil
	}

	checkpointResult <- resp
}

// commandVersionCheck实现command.VersionCheckFunc，用来做版本检查
func commandVersionCheck() (command.VersionCheckInfo, error) {
	// 等待获取结果
	info := <-checkpointResult
	if info == nil {
		var zero command.VersionCheckInfo
		return zero, nil
	}

	// 构建我们收到的版本输出
	alerts := make([]string, len(info.Alerts))
	for i, a := range info.Alerts {
		alerts[i] = a.Message
	}
	return command.VersionCheckInfo{
		Outdated: info.Outdated,
		Lastest:  info.CurrentVersion,
		Alerts:   alerts,
	}, nil
}
