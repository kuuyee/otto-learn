package otto

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/otto/appfile"
	"github.com/hashicorp/otto/directory"
	"github.com/hashicorp/otto/ui"
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

	fmt.Println(infra, infraCtx, foundcations, foundationCtxs)

	return nil
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
