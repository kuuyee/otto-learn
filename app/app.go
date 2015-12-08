// app包中包括了接口和结构，实现了Otto应用类型。
// 应用是一些一些组件用来dev,build和deploy一些
// 像Rails，PHP等类型的应用。
//
// 所有的app实现都指定包含了三个元数据:
// (app type, infra type, infora flavor),例如：
// (rails,aws,vpc-public-private). app实现只需要
// 满足这个三个元数据。
//
// When building app plugins, it is possible for that plugin to support
// multiple matrix elements, but each implementation of the interface
// is expeced to only implement one.

package app

import (
	"github.com/kuuyee/otto-learn/appfile"
	//"github.com/hashicorp/otto/context"
	"github.com/hashicorp/otto/ui"
	"github.com/kuuyee/otto-learn/context"
	"github.com/kuuyee/otto-learn/foundation"
)

// App接口，必须实现(app type,infra type,infra flavor)这三个元数据
type App interface {
	// Compile被调用编译APP文件
	Compile(*Context) (*CompileResult, error)

	Build(*Context) error

	Deploy(*Context) error

	// Dev用来管理开发环境，在本地调用。
	Dev(*Context) error

	// DevDep是当这个应用是其它应用的上层依赖是被调用。
	// 开发环境自己build和配置
	//
	// DevDep给定两个上下文。第一个是目标APP(开发的APP)
	// 第二个是源APP(上游APP)
	//
	// DevDep如果是nil表示DevDep结构不需要做什么.
	DevDep(dst *Context, src *Context) (*DevDep, error)
}

// Context是操作应用的上下文.这里有些字段只是用来操作某一种操作
type Context struct {
	context.Shared

	CompileResult *CompileResult

	// Action是一个子动作
	//
	// ActionArgs是动作的参数列表
	//
	// 只能被当前调用的Infra的设置
	Action     string
	ActionArgs []string

	// Dir是编译时可用来做持久化存储的目录。当编译完成
	// 时，就清理这个目录
	Dir string

	// 缓存目录
	CacheDir string

	// 单一Appfile本地数据存储
	// 编译后不清楚
	LocalDir string

	// Tuple是foundation用的元数据
	Tuple Tuple

	// appfile中应用配置本身
	Application *appfile.Application

	// DevDepFragment用来填充开发依赖，Vagrantfile片段。
	// 只在编译期调用。
	DevDepFragments []string

	// DevIPAddress 是本地IP地址，用在开发环境
	//
	// 只在app是root应用时可用
	DevIPAddress string
}

// RouteName实现了router.Context接口，所以我们能用Router
func (c *Context) RouteName() string {
	return c.Action
}

// RouteName实现了router.Context接口，所以我们能用Router
func (c *Context) RouteArgs() []string {
	return c.ActionArgs
}

// RouteName实现了router.Context接口，所以我们能用Router
func (c *Context) UI() ui.Ui {
	return c.UI()
}

type CompileResult struct {
	// Version是编译结构的版本。纯元数据
	// app本身应该直接使用某些特性to run
	Version uint32 `json:"version"`

	// FoundationConfig 配置otto各种foundation元素
	FoundationConfig foundation.Config `json:"foundation_config"`

	// DevDepFragmentPath 是Vagrantfile碎片路径，
	// 增加依赖是加入Vagrantfile文件中
	DevdepFragmentPath string `json:"dev_dep_fragment_path"`

	// FoundationResults 是foundation的编译结果
	//
	// Otto核心，如果有任何值将忽略
	FoundationResults map[string]*foundation.CompileResult `json:"foundation_results"`
}
