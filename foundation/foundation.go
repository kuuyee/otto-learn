package foundation

import (
	"github.com/kuuyee/otto-learn/context"
)

// Foundation接口，每个foundation必须实现。
// 一个foundation绑定一个真实的基础设施。
// 并能够给服务发现和安全呈现一个菜单。
//
// Foundation一定是一个(name,infra type,infra flavor) 3-tuple
type Foundation interface {
	// Compile在编译文件时调用
	Compile(*Context) (*CompileResult, error)

	// Infra用来构建和销毁基础设施。
	// Context中的Action字段决定期望的动作
	// 目前只支持 build 和 destroy操作
	Infra(*Context) error
}

// Context是Foundation的操作上下文
type Context struct {
	context.Shared

	// Action是一个子动作
	//
	// ActionArgs是动作的参数列表
	//
	// 只能被当前调用的Infra的设置
	Action     string
	ActionArgs []string

	// Config 是Appfile中对于foundation原始配置
	Config map[string]interface{}

	// AppConfig是我们正在使用APP的foundation配置
	// 如果我们正在应用中使用编译，这个功能才能用
	//
	// 编译时需要，可能是nil. nil不是好的定义，但是
	// nil会什么都不做，除了deploy.
	AppConfig *Config

	// Dir是编译时可用来做持久化存储的目录。当编译完成
	// 时，就清理这个目录
	Dir string

	// Tuple是foundation用的元数据
	Tuple Tuple
}

// CompileResult是一个包含编译结果值的struct
//
// 现在是空的，但是将来会用到
type CompileResult struct{}
