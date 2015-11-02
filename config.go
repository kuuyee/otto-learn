package main

import (
//"fmt"
//"github.com/hashicorp/hcl"
//"io/ioutil"
)

// 用一个结构来配置Otto CLI
//
// 这不是用来配置Otto本身，那在`config`包中
type Config struct {
	DisableCheckpoint          bool `hcl:"disable_checkpoint"`
	DisableCheckpointSignature bool `hcl:disable_checkpoint_signature`
}

var BuiltinConfig Config

// ConfigDir 为Otto返回配置目录
func ConfigDir() (string, error) {
	return configDir()
}

// LoadConfig 从.ottorc文件装载CLI配置
