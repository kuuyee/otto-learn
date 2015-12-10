package appfile

import (
	_ "bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/go-multierror"
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

func (c *Compiled) Validate() error {
	return nil
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

	// 构建一个存储用来保存imports
	importsStorage := &getter.FolderStorage{
		StorageDir: filepath.Join(opts.Dir, CompileImportsFolder)}
	importOpts := &compileImportOpts{
		Storage:   importsStorage,
		Cache:     make(map[string]*File),
		CacheLock: &sync.Mutex{},
	}
	if err := compileImports(f, importOpts, opts); err != nil {
		return nil, err
	}

	// 早期验证root
	if err := f.Validate(); err != nil {
		return nil, err
	}

	// 在Appfile加入root定点
	vertex := &CompiledGraphVertex{File: f, NameValue: f.Application.Name}
	compiled.Graph.Add(vertex)

	// 构建存储用来保存下载的依赖，那么可以用来触发递归调用下载所有的依赖
	storage := &getter.FolderStorage{
		StorageDir: filepath.Join(opts.Dir, CompileDepsFolder)}
	if err := compileDependencies(
		storage,
		importOpts,
		compiled.Graph, opts, vertex); err != nil {
		return nil, err
	}

	// 验证编译的文件树
	if err := compiled.Validate(); err != nil {
		return nil, err
	}

	// 写入编译的Appfile数据
	if err := compileWrite(opts.Dir, compiled); err != nil {
		return nil, err
	}

	return compiled, nil
}

func compileDependencies(
	storage getter.Storage,
	importOpts *compileImportOpts,
	graph *dag.AcyclicGraph,
	opts *CompileOpts,
	root *CompiledGraphVertex) error {
	return nil
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

func compileWrite(dir string, compiled *Compiled) error {
	return nil
}

type compileImportOpts struct {
	Storage   getter.Storage
	Cache     map[string]*File
	CacheLock *sync.Mutex
}

// compileImports需要一个文件，load所有的imports,然后合并到文件中
func compileImports(root *File, importOpts *compileImportOpts, opts *CompileOpts) error {
	// 如果没有imports，直接短路(返回)
	if len(root.Imports) == 0 {
		return nil
	}

	// 把它们放入变量，以便我们可以更早的引用
	//storage := importOpts.Storage
	//cache := importOpts.Cache
	//cacheLock := importOpts.CacheLock

	// A graph is used to track for cycles
	var graphLock sync.Mutex
	graph := new(dag.AcyclicGraph)
	graph.Add("root")

	// 自从我们平行的执行导入，在同一时间会有多个error发生
	// 我们使用multierror lock和跟踪error
	var resultErr error
	var resultErrLock sync.Mutex

	// Forward declarations for some nested functions we use. The docs
	// for these functions are above each.
	var importSingle func(parent string, f *File) bool
	var downloadSingle func(string, *sync.WaitGroup, *sync.Mutex, []*File, int)

	// importSingle is responsible for kicking off the imports and merging
	// them for a single file. This will return true on success, false on
	// failure. On failure, it is expected that any errors are appended to
	// resultErr.
	importSingle = func(parent string, f *File) bool {
		var wg sync.WaitGroup

		// 构建文件列表，后面将合并
		var mergeLock sync.Mutex
		merge := make([]*File, len(f.Imports))

		// 通过导入并开始处理下载
		for idx, i := range f.Imports {
			source, err := getter.Detect(i.Source, filepath.Dir(f.Path), getter.Detectors)
			if err != nil {
				resultErrLock.Lock()
				defer resultErrLock.Unlock()
				resultErr = multierror.Append(resultErr, fmt.Errorf(
					"获取import源错误: %s", err))
				return false
			}

			// Add this to the graph and check now if there are cycles
			graphLock.Lock()
			graph.Add(source)
			graph.Connect(dag.BasicEdge(parent, source))
			cycles := graph.Cycles()
			graphLock.Unlock()
			if len(cycles) > 0 {
				for _, cycle := range cycles {
					names := make([]string, len(cycle))
					for i, v := range cycle {
						names[i] = dag.VertexName(v)
					}

					resultErrLock.Lock()
					defer resultErrLock.Unlock()
					resultErr = multierror.Append(resultErr, fmt.Errorf(
						"Cycle found: %s", strings.Join(names, ", ")))
					return false
				}
			}
			wg.Add(1)
			go downloadSingle(source, &wg, &mergeLock, merge, idx)
		}

		// Wait for completion
		wg.Wait()

		// Go through the merge list and look for any nil entries, which
		// means that download failed. In that case, return immediately.
		// We assume any errors were put into resultErr.
		for _, importF := range merge {
			if importF == nil {
				return false
			}
		}

		for _, importF := range merge {
			// We need to copy importF here so that we don't poison
			// the cache by modifying the same pointer.
			importFCopy := *importF
			importF = &importFCopy
			source := importF.ID
			importF.ID = ""
			importF.Path = ""

			// Merge it into our file!
			if err := f.Merge(importF); err != nil {
				resultErrLock.Lock()
				defer resultErrLock.Unlock()
				resultErr = multierror.Append(resultErr, fmt.Errorf(
					"合并import错误 %s : %s", source, err))
				return false
			}
		}
		return true
	}

	// downloadSingle is used to download a single import and parse the
	// Appfile. This is a separate function because it is generally run
	// in a goroutine so we can parallelize grabbing the imports.
	downloadSingle = func(source string, wg *sync.WaitGroup, l *sync.Mutex, result []*File, idx int) {}

	importSingle("root", root)
	return resultErr
}
