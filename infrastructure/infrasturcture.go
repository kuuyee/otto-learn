package infrastructure

import (
	"github.com/hashicorp/otto/ui"
	"github.com/kuuyee/otto-learn/appfile"
	"github.com/kuuyee/otto-learn/context"
)

// Infrastructure 是每个infrasturcture必须
// 实现的接口
type Infrastructure interface {
	// Creds 是当otto需要对infrastructure供应商提供的认证
	// Infra需要查询用户(或环境)的认证并返回它们。Otto会处理
	// 加密、存储和接收认证.
	Creds(*Context) (map[string]string, error)

	// VerifyCreds 当接收缓存的认证结果被调用
	// 在执行任何操作之前这里给Infrastructure实现
	// 一个机会来检查认证是否OK
	VerifyCreds(*Context) error

	Execute(*Context) error
	Compile(*Context) (*CompileResult, error)
	Flavors() []string
}

// Context是操作infrastructure的上下文环境。下面的字段
// 只是针对某一个操作有效
type Context struct {
	context.Shared

	// Action是子操作
	//
	// ActionArgs 是操作的参数列表
	//
	// 这两个字段都只是用来设置执行调用的
	Action     string
	ActionArgs []string

	// Dir是执行任务编译时可用来做持久化存储的目录。
	Dir string

	// appfile中配置infrastructure。这包括我们期望的
	// infrastructure flavor
	Infra *appfile.Infrastructure
}

// RouteName实现了router.Context接口，所以我们可以使用路由
func (c *Context) RouteName() string {
	return c.Action
}

// RouteArgs实现了router.Context接口，所以我们可以使用路由
func (c *Context) RouteArgs() []string {
	return c.ActionArgs
}

// Ui实现了router.Context接口，所以我们可以使用路由
func (c *Context) UI() ui.Ui {
	return c.Ui
}

// CompileResult 是编译结果的结构
type CompileResult struct{}
