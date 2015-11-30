package otto

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/otto/appfile"
	"github.com/hashicorp/otto/directory"
	"github.com/hashicorp/otto/helper/localaddr"
	"github.com/hashicorp/otto/ui"
	"github.com/hashicorp/terraform/dag"
	"github.com/kuuyee/otto-learn/app"
	"github.com/kuuyee/otto-learn/context"
	"github.com/kuuyee/otto-learn/foundation"
	"github.com/kuuyee/otto-learn/infrastructure"
)

// Core是otto库的主结构
type Core struct {
	appfile         *appfile.File
	appfileCompiled *appfile.Compiled
	apps            map[app.Tuple]app.Factory
	dir             directory.Backend
	infras          map[string]infrastructure.Factory
	foundationMap   map[foundation.Tuple]foundation.Factory
	dataDir         string
	localDir        string
	compileDir      string
	ui              ui.Ui
}

// CoreConfig是创建NewCore的配置
type CoreConfig struct {
	// DataDir是本地数据存放的目录
	// 对于所有Otto进程是全局数据
	//
	// LocalDir是单个Appfile的本地数据目录
	// 这清理与否对编译并不重要
	//
	// CompiledDir 编译数据存放的路径，每次
	// 编译都会清理目录
	DataDir    string
	LocalDir   string
	CompileDir string

	// Appfile 是core使用的配置文件，其必须
	// 是编译过的Appfile
	Appfile *appfile.Compiled

	// Directory是Appfile相关数据存储目录
	Directory directory.Backend

	// Apps是存在的app实现映射
	Apps map[app.Tuple]app.Factory

	// Infrastructures是存在的infrastructures
	// 映射，值是创建infrastructures实现的factory
	Infrastructures map[string]infrastructure.Factory

	// Foundation是存在的foundations的实现，值是创建foundation
	// 实现的factory
	Foundations map[foundation.Tuple]foundation.Factory

	// Ui 用来于用户交互
	Ui ui.Ui
}

// NewCore创建一个core
//
// 一旦调用这个函数，since the Core may use parts of it without deep copying.
// CoreConfig不能再被使用和更改。
func NewCore(c *CoreConfig) (*Core, error) {
	return &Core{
		appfile:         c.Appfile.File,
		appfileCompiled: c.Appfile,
		apps:            c.Apps,
		dir:             c.Directory,
		infras:          c.Infrastructures,
		foundationMap:   c.Foundations,
		dataDir:         c.DataDir,
		localDir:        c.LocalDir,
		compileDir:      c.CompileDir,
		ui:              c.Ui,
	}, nil
}

// 编译任务，编译Appfile下所有的结果数据
func (c *Core) Compile() error {
	// 获取infra实现
	infra, infraCtx, err := c.infra()
	if err != nil {
		return err
	}

	// 取得所有foundation实现(和infra关联的)
	foundcations, foundationCtxs, err := c.foundations()

	// 删除之前的output目录
	log.Printf("[INFO] 删除之前编译的内容：%s", c.compileDir)
	if err := os.RemoveAll(c.compileDir); err != nil {
		return err
	}

	// 编译Infrastructure给应用
	log.Printf("[INFO] 运行infra编译...")
	c.ui.Message("编译 infra...")
	if _, err := infra.Compile(infraCtx); err != nil {
		return err
	}

	// 编译foundcation给应用(不关联任何应用)。这个foundcation编译用来
	// 给`otto infra`设置准备
	log.Printf("[INFO] 运行foundation编译...")
	for i, f := range foundcations {
		ctx := foundationCtxs[i]
		c.ui.Message(fmt.Sprintf(
			"编译foundation: %s", ctx.Tuple.Type))
		if _, err := f.Compile(ctx); err != nil {
			return err
		}
	}

	//排除所有依赖，并全部编译。我们必须编译每个依赖以备
	//dev构建
	var resultLock sync.Mutex
	results := make([]*app.CompileResult, 0, len(c.appfileCompiled.Graph.Vertices()))
	err = c.walk(func(app app.App, ctx *app.Context, root bool) error {
		if !root {
			c.ui.Header(fmt.Sprintf(
				"编译依赖 '%s'...", ctx.Appfile.Application.Name))
		} else {
			c.ui.Header(fmt.Sprintf(
				"编译主应用"))
		}

		// 如果是root，设置dev dep fragments
		if root {
			// We grab the lock just in case although if we're the
			// root this should be serialized.
			resultLock.Lock()
			ctx.DevDepFragments = make([]string, 0, len(results))
			for _, result := range results {
				if result.DevdepFragmentPath != "" {
					ctx.DevDepFragments = append(ctx.DevDepFragments, result.DevdepFragmentPath)
				}
			}
			resultLock.Unlock()
		}

		// 编译！
		result, err := app.Compile(ctx)
		if err != nil {
			return err
		}

		// 为应用编译foundation
		subdirs := []string{"app-dev", "app-dev-dep", "app-build", "app-deploy"}
		for i, f := range foundcations {
			fCtx := foundationCtxs[i]
			fCtx.Dir = ctx.FoundationDirs[i]
			if result != nil {
				fCtx.AppConfig = &result.FoundationConfig
			}

			if _, err := f.Compile(fCtx); err != nil {
				return err
			}

			// 确保子目录存在
			for _, dir := range subdirs {
				if err := os.MkdirAll(filepath.Join(fCtx.Dir, dir), 0755); err != nil {
					return err
				}
			}
		}

		// 最后保存编译结果
		resultLock.Lock()
		defer resultLock.Unlock()
		results = append(results, result)

		return nil
	})

	//fmt.Println(infra, infraCtx, foundcations, foundationCtxs)
	return err
}

func (c *Core) walk(f func(app.App, *app.Context, bool) error) error {
	root, err := c.appfileCompiled.Graph.Root()
	if err != nil {
		return fmt.Errorf("装载App报错: %s", err)
	}

	//Walk the appfile graph
	var stop int32 = 0
	return c.appfileCompiled.Graph.Walk(func(raw dag.Vertex) (err error) {
		// 如果stop(发生一些错误)，那么尽早stop.
		// If we're told to stop (something else had an error), then stop early.
		// Graphs walks by default will complete all disjoint parts of the
		// graph before failing, but Otto doesn't have to do that.
		if atomic.LoadInt32(&stop) != 0 {
			return nil
		}

		//如果报错退出，我们标记stop atomic
		defer func() {
			if err != nil {
				atomic.StoreInt32(&stop, 1)
			}
		}()

		// 转换至丰富的Vertex以便我们能够访问数据
		v := raw.(*appfile.CompiledGraphVertex)

		// 给appfile获取App上下文
		appCtx, err := c.appContext(v.File)
		if err != nil {
			return fmt.Errorf(
				"loading Appfile for '%s': %s 报错", dag.VertexName(raw), err)
		}

		app, err := c.app(appCtx)
		if err != nil {
			return fmt.Errorf(
				"获取App实现报错 '%s': %s", dag.VertexName(raw), err)
		}

		// 执行回调
		return f(app, appCtx, raw == root)
	})
}

func (c *Core) infra() (infrastructure.Infrastructure, *infrastructure.Context, error) {
	// 取得infrastructure配置
	config := c.appfile.ActiveInfrastructure()
	if config == nil {
		return nil, nil, fmt.Errorf(
			"infrastructure在Appfile中没找到： %s", c.appfile.Project.Infrastructure)
	}

	// 获取infrastructure工厂
	f, ok := c.infras[config.Type]
	if !ok {
		return nil, nil, fmt.Errorf(
			"infrastructure类型不支持：%s",
			config.Type)
	}

	// 开始实现infrastructure
	infra, err := f()
	if err != nil {
		return nil, nil, err
	}

	// 数据输出目录
	outputDir := filepath.Join(
		c.compileDir, fmt.Sprintf("infra-%s", c.appfile.Project.Infrastructure))

	// Build the context
	return infra, &infrastructure.Context{
		Dir:   outputDir,
		Infra: config,
		Shared: context.Shared{
			Appfile:    c.appfile,
			InstallDir: filepath.Join(c.dataDir, "binaries"),
			Directory:  c.dir,
			Ui:         c.ui,
		},
	}, nil
}

func (c *Core) foundations() ([]foundation.Foundation, []*foundation.Context, error) {
	// 取得infrastructure配置
	config := c.appfile.ActiveInfrastructure()
	if config == nil {
		return nil, nil, fmt.Errorf(
			"infrastructure在appfile中没找到：%s", c.appfile.Project.Infrastructure)
	}

	// 如果没有foundation，返回nil
	if len(config.Foundations) == 0 {
		return nil, nil, nil
	}

	// 给列表创建一个数组
	fs := make([]foundation.Foundation, 0, len(config.Foundations))
	ctxs := make([]*foundation.Context, 0, cap(fs))
	for _, f := range config.Foundations {
		// The tuple we're looking for is the foundation type, the
		// infrastructure type, and the infrastructure flavor. Build that
		// tuple.
		tuple := foundation.Tuple{
			Type:        f.Name,
			Infra:       config.Type,
			InfraFlavor: config.Flavor,
		}

		// 查找匹配的foundation
		fun := foundation.TupleMap(c.foundationMap).Lookup(tuple)
		if fun == nil {
			return nil, nil, fmt.Errorf(
				"tuple的foundation实现没找到: %s", tuple)
		}

		// 实例化实现
		impl, err := fun()
		if err != nil {
			return nil, nil, err
		}

		// 数据输出目录
		outputDir := filepath.Join(
			c.compileDir, fmt.Sprintf("foundation-%s", f.Name))

		// build the context
		ctx := &foundation.Context{
			Config: f.Config,
			Dir:    outputDir,
			Tuple:  tuple,
			Shared: context.Shared{
				Appfile:    c.appfile,
				InstallDir: filepath.Join(c.dataDir, "binaries."),
				Directory:  c.dir,
				Ui:         c.ui,
			},
		}

		// 加入到结果中
		fs = append(fs, impl)
		ctxs = append(ctxs, ctx)
	}

	return fs, ctxs, nil
}

func (c *Core) appContext(f *appfile.File) (*app.Context, error) {
	// 我们需要配置可用的Infrastructure,以便后面额可以build tuple
	config := f.ActiveInfrastructure()
	if config == nil {
		return nil, fmt.Errorf(
			"没有在appfile中找到infrastructure ：%s", f.Project.Infrastructure)
	}

	// The tuple we're looking for is the application type, the
	// infrastructure type, and the infrastructure flavor. Build that
	// tuple.
	tuple := app.Tuple{
		App:         f.Application.Type,
		Infra:       config.Type,
		InfraFlavor: config.Flavor,
	}

	// outputDir数据输出目录，This is either the main app so
	// it goes directly into "app" or it is a dependency and goes into
	// a dep folder.
	outputDir := filepath.Join(c.compileDir, "app")
	if id := f.ID; id != c.appfile.ID {
		outputDir = filepath.Join(c.compileDir, fmt.Sprintf("dep-%s", id))
	}

	//app的缓存目录
	cacheDir := filepath.Join(c.dataDir, "cache", f.ID)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf(
			"创建缓存目录报错 '%s': %s", cacheDir, err)
	}

	//为foundation构建context.We use this
	// to also compile the list of foundation dirs.
	foundationDirs := make([]string, len(config.Foundations))
	for i, f := range config.Foundations {
		foundationDirs[i] = filepath.Join(outputDir, fmt.Sprintf("foundcation-%s", f.Name))
	}

	//  获取dev IP地址
	ipDB := &localaddr.CachedDB{
		DB:        &localaddr.DB{Path: filepath.Join(c.dataDir, "ip.db")},
		CachePath: filepath.Join(c.localDir, "dev_ip"),
	}
	ip, err := ipDB.IP()
	if err != nil {
		return nil, fmt.Errorf(
			"接受dev IP地址报错： %s", err)
	}
	return &app.Context{
		Dir:          outputDir,
		CacheDir:     cacheDir,
		LocalDir:     c.localDir,
		Tuple:        tuple,
		Application:  f.Application,
		DevIPAddress: ip.String(),
		Shared: context.Shared{
			Appfile:        f,
			FoundationDirs: foundationDirs,
			InstallDir:     filepath.Join(c.dataDir, "binaries"),
			Directory:      c.dir,
			Ui:             c.ui,
		},
	}, nil
}

func (c *Core) app(ctx *app.Context) (app.App, error) {
	log.Printf("[INFO]为Tuple装载App实现： %s", ctx.Tuple)

	// 查找app实现，factory
	f := app.TupleMap(c.apps).Lookup(ctx.Tuple)
	if f == nil {
		return nil, fmt.Errorf(
			"tuple的app实现没有找到： %s", ctx.Tuple)
	}

	// Start the impl.
	result, err := f()
	if err != nil {
		return nil, fmt.Errorf(
			"app failed to start properly: %s", err)
	}
	return result, nil
}
