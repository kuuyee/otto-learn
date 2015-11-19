package command

import (
	"bytes"
	"fmt"
)

// VersionCommand 实现了一个打印版本的命令
type VersionCommand struct {
	Meta

	Reversion         string
	Version           string
	VersionPrerelease string
	CheckFunc         VersionCheckFunc
}

// VersionCheckFunc回调函数用来检查是否为新版本
type VersionCheckFunc func() (VersionCheckInfo, error)

// VersinCheckInfo是回调函数VersionCheckFunc的返回值，描述
// Otto最终版本信息
type VersionCheckInfo struct {
	Outdated bool
	Lastest  string
	Alerts   []string
}

func (v *VersionCommand) Help() string {
	return ""
}

func (v *VersionCommand) Run(args []string) int {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "Otto v%s", v.Version)
	if v.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", v.VersionPrerelease)

		if v.Reversion != "" {
			fmt.Fprintf(&versionString, " (%s)", v.Version)
		}
	}

	v.Ui.Output(versionString.String())

	// 如果有版本检查函数，那么还要检查最终版本
	if v.CheckFunc != nil {
		// 分隔输出到新的行
		v.Ui.Output("")

		// 检查最终版本
		info, err := v.CheckFunc()
		if err != nil {
			v.Ui.Error(fmt.Sprintf("检查最终版本报错: %s", err))
		}
		if info.Outdated {
			v.Ui.Output(fmt.Sprintf(
				"你的Otto版本已经被弃用！最新版是: %s. 你可以"+
					"从www.ottoproject.io下载升级", info.Lastest))
		}
	}
	return 0
}

func (v *VersionCommand) Synopsis() string {
	return "打印 Otto 版本"
}
