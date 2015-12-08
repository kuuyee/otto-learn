package appfile

import (
	_ "bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform/dag"
	"github.com/kuuyee/otto-learn/appfile/detect"
)

const (
	// compileVersion 是我们当前要编译的版本。This can be used in the future to change
	// the directory structure and on-disk format of compiled appfiles.CompileVersion = 1
	CompileFilename        = "Appfile.compiled"
	CompileDepsFolder      = "deps"
	CompileImportsFolder   = "deps"
	CompileVersionFilename = "version"
)

// Compiled 表示一个""编译的"Appfile.一个编译的Appfile装载所有依赖
// 的Appfile,完整的导入，验证正确性，等等
//
// Appfile compilation is a process that requires network activity and
// has to occur once. The idea is that after compilation, a fully compiled
// Appfile can then be loaded in the future without network connectivity.
// Additionally, since we can assume it is valid, we can load it very quickly.
type Compiled struct {
	// File 是原始应用
	File *File

	// Graph 是DAG(Directed Acyclic Graph) 有向无环图，包括所以依赖。
	// This is already verified to have no cycles. Each vertex is a *CompiledGraphVertex.
	Graph *dag.AcyclicGraph
}

// CompileGraphVertex is the type of the vertex within the Graph of Compiled.
type CompiledGraphVertex struct {
	// File 是原始Appfile
	File *File

	// Dir is the directory of the data root for this dependency. This
	// is only non-empty for dependencies (the root vertex does not have
	// this value).
	Dir string

	// 不要在包外使用
	NameValue string
}

// CompileOpts 是编译选项
type CompileOpts struct {
	// Dir是所有编译数据存放目录
	// 要想使用编译的Appfile，这个目录不能为空
	Dir string

	// Detect 是发现配置，用来处理默认的依赖
	Detect *detect.Config

	// Callbak 是在编译期间接收事件通知的选项。参数CompileEvent需要
	// 用Type switch决定
	Callback func(CompileEvent)
}

// CompileEvent 是Callback可能接收的事件
type CompileEvent interface{}

// Compile 编译Appfile
//
// 这里如果有外部依赖，可能需要网络连接
// 在给定目录回装载全部的依赖，Appfile也会保存在这里。
//
// LoadCompile会装载一个提前编译的Appfile
//
// 如果你不想重新装载一个编译的Appfile,你可以完全的删除目录。
// 这个功能前提是目录存在
func Compile(f *File, opts *CompileOpts) (*Compiled, error) {
	// 先清理目录 In the future, we can keep it around
	// and do incremental compilations.
	if err := os.RemoveAll(opts.Dir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return nil, err
	}

	// 写入一个版本
	if err := compileVersion(opts.Dir); err != nil {
		return nil, fmt.Errorf("写入编译的Appfile版本报错： %s", err)
	}

	// 开始构建编译的Appfile
	compiled := &Compiled{File: f, Graph: new(dag.AcyclicGraph)}

	// 检查是否有ID.如果没有那么需要写入一个
	// 前提是文件有路径
	if f.Path != "" {
		hasID, err := f.hasID()
		if err != nil {
			return nil, fmt.Errorf(
				"检查Appfile UUID错误：%s", err)
		}

		if !hasID {
			if err := f.initID(); err != nil {
				return nil, fmt.Errorf(
					"在Appfile写入UUID错误: %s", err)
			}
		}

		if err := f.loadID(); err != nil {
			return nil, fmt.Errorf(
				"读取Appfile UUID错误: %s", err)
		}
	}

	return compiled, nil
}

func compileVersion(dir string) error {
	f, err := os.Create(filepath.Join(dir, CompileVersionFilename))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%d", CompileVersionFilename)
	return err
}
