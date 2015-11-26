package context

import (
	"github.com/hashicorp/otto/appfile"
	"github.com/hashicorp/otto/directory"
	"github.com/hashicorp/otto/ui"
)

// Shared用来在app/infra中共享上下文
type Shared struct {
	// InfraCreds 是在infrastructure中使用的证书.
	// 确保在调用下列函数时被使用:
	//
	// 	 App.Build
	//
	InfraCreds map[string]string

	// Ui是UI对象用来于用户交互
	Ui ui.Ui

	// Directory是目录服务.在执行和编译期间提供服务的
	Directory directory.Backend

	// InstallDir是放置二进制文件的目录
	InstallDir string

	// appfile
	Appfile *appfile.File

	// FoundationDirs是放置各种基础脚本的目录
	//
	// 这些目录会包含 dev,deploy子目录,将会在环境中被装载
	// 目录中将存在一个"main.sh"并被调用
	FoundationDirs []string
}
