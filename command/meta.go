package command

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/otto/appfile"
	"github.com/hashicorp/otto/directory"
	"github.com/hashicorp/otto/ui"
	"github.com/kuuyee/otto-learn/otto"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
)

const (
	// DefaultAppfile 是Appfile的默认文件名
	DefaultAppfile = "Appfile"

	// DefaultLocalDataDir 是local数据的默认路径
	DefaultLocalDataDir         = "~/.otto.d"
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

var (
	// AltAppfiles是APPfile的别名，Otto可以通过别名发现和装载。
	AltAppfiles = []string{"appfile.hcl"}
)

// FlagSetFlags 是枚举，用来定义默认FlagSet
type FlagSetFlags uint

const (
	FlagSetNone FlagSetFlags = 0
)

// Meta是命令的元选项
type Meta struct {
	CoreConfig *otto.CoreConfig
	Ui         cli.Ui
}

// Core返回一个Appfile的Core.Appfile应该从appfile.File.Path装载
// root appfile路径将作为Otto的默认输出路径
func (m *Meta) Core(f *appfile.Compiled) (*otto.Core, error) {
	if f.File == nil || f.File.Path == "" {
		return nil, fmt.Errorf("不能确定Appfile目录")
	}

	rootDir, err := m.RootDir(filepath.Dir(f.File.Path))
	if err != nil {
		return nil, err
	}

	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	dataDir, err := m.DataDir()
	if err != nil {
		return nil, err
	}

	config := *m.CoreConfig
	config.Appfile = f
	config.DataDir = dataDir
	config.LocalDir = filepath.Join(
		rootDir, DafaultDataDir, DefaultOutputDirLocalData)
	config.CompileDir = filepath.Join(
		rootDir, DefaultOutputDir, DefaultOutputDirCompiledData)
	config.Ui = m.OttoUi()
	config.Directory, err = m.Directory(&config)
	if err != nil {
		return nil, err
	}

	return otto.NewCore(&config)
}

// DataDir返回Otto用户本地数据目录
func (m *Meta) DataDir() (string, error) {
	return homedir.Expand(DefaultLocalDataDir)
}

// RootDir定位"root"目录，这是Otto、Appfile的工作目录
// 我们一直往上层目录查找，直到".otto"目录，假定在这个目录
func (m *Meta) RootDir(startDir string) (string, error) {
	current := startDir

	// 向上查找，这是一个无限循环
	// We also protect this
	// loop with a basic infinite loop guard.
	i := 0
	prev := ""
	for prev != current && i < 1000 {
		if _, err := os.Stat(filepath.Join(current, DefaultOutputDir)); err == nil {
			// 找到appfile
			return current, nil
		}

		prev = current
		current = filepath.Dir(current)
		i++
	}

	return "", fmt.Errorf(
		"Otto doesn't appear to have compiled your Appfile yet!\n\n" +
			"Run `otto compile` in the directory with the Appfile or\n" +
			"with the `-appfile` flag in order to compile the files for\n" +
			"developing, building, and deploying your application.\n\n" +
			"Once the Appfile is compiled, you can run `otto` in any\n" +
			"subdirectory.")
}

// Directory返回Otto后端目录，如果没有指定，将使用Local目录
func (m *Meta) Directory(config *otto.CoreConfig) (directory.Backend, error) {
	return &directory.BoltBackend{
		Dir: filepath.Join(config.DataDir, "directory"),
	}, nil
}

// FlagSet 返回每个命令的公共flag。FlagSet的具体行为
// 可以通过第二个参数来设定
func (m *Meta) FlagSet(n string, fs FlagSetFlags) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)

	// 给Ui错误创建一个io.Writer。这是一个hack，但是其处理job。基本上：
	// 创建一个pipe,然后扫描每一行，并输出每一行到UI，不断的循环处理。
	errR, errW := io.Pipe()
	errScanner := bufio.NewScanner(errR)
	go func() {
		for errScanner.Scan() {
			m.Ui.Error(errScanner.Text())
		}
	}()
	f.SetOutput(errW)

	return f
}

// OttoUi返回ui.Ui对象
func (m *Meta) OttoUi() ui.Ui {
	return NewUi(m.Ui)
}
