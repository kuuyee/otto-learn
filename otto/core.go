package otto

import (
	_ "fmt"
	"github.com/hashicorp/otto/appfile"
	"github.com/hashicorp/otto/directory"
	//"github.com/hashicorp/otto/infrastructure"
	"github.com/hashicorp/otto/ui"
	"github.com/kuuyee/otto-learn/app"
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
