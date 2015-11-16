package command

import (
	_ "fmt"
	"github.com/kuuyee/otto-learn/otto"
	"github.com/mitchellh/cli"
)

const (
	// DefaultAppfile 是Appfile的默认文件名
	DefaultAppfile = "Appfile"

	// DefaultLocalDataDir 是local数据的默认路径
	DefaultLocalData            = "~/.otto.d"
	DefaultLocalDataDetectorDir = "detect"

	// DefaultOutputDir 是输出目录的默认路径
	DefaultOutputDir                = ".otto"
	DefaultOutputDirCompiledAppfile = "appfile"
	DefaultOutputDirCompiledData    = "compiled"
	DefaultOutputDirLocalData       = "data"

	// DefaultDataDir 是数据目录的默认，
	// 如果在Appfile没有指定的话
	DafaultDataDir = "otto-data"
)

// Meta是命令的元选项
type Meta struct {
	CoreConfig *otto.CoreConfig
	Ui         cli.Ui
}
