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
	_ "github.com/hashicorp/otto/appfile"
	//"github.com/hashicorp/otto/context"
	_ "github.com/hashicorp/otto/foundation"
	_ "github.com/hashicorp/otto/ui"
	_ "github.com/kuuyee/otto-learn/context"
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
}

type CompileResult struct{}
