package command

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	//"github.com/hashicorp/otto/appfile"
	"github.com/kuuyee/otto-learn/appfile"
	//"github.com/hashicorp/otto/appfile/detect"
	"github.com/hashicorp/otto/ui"
	"github.com/kuuyee/otto-learn/appfile/detect"
)

// CompileCommand是一个编译命令，要来把
// Appfile编译成一组数据供其他命令使用
type CompileCommand struct {
	Meta
	Detectors []*detect.Detector //在main.commands.go中初始化
}

func (c *CompileCommand) Run(args []string) int {
	var flagAppfile string
	fs := c.FlagSet("compile", FlagSetNone)
	fs.Usage = func() { c.Ui.Error(c.Help()) }
	//把参数--appfile的值写入&flagAppfile
	fs.StringVar(&flagAppfile, "appfile", "", "")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Load a UI
	ui := c.OttoUi()
	ui.Header("装载 Appfile...")

	fmt.Printf("[KuuYee]====> flagAppfile: %+v\n", flagAppfile)
	app, appPath, err := loadAppfile(flagAppfile)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	fmt.Printf("[KuuYee]====> appPath: %+v\n", appPath)

	// 如果没有Appfile，告诉用户发生了什么
	if app == nil {
		ui.Header("没有发现Appfile! Detecting project information...")
		ui.Message(fmt.Sprintf(
			"No Appfile was found. If there is no Appfile, Otto will do its best\n" +
				"to detect the type of application this is and set reasonable defaults.\n" +
				"This is a good way to get started with Otto, but over time we recommend\n" +
				"writing a real Appfile since this will allow more complex customizations,\n" +
				"the ability to reference dependencies, versioning, and more."))
	}

	// 解析
	dataDir, err := c.DataDir()
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	fmt.Printf("[KuuYee]====> dataDir: %+v\n", dataDir)
	detectorDir := filepath.Join(dataDir, DefaultLocalDataDetectorDir)
	fmt.Printf("[KuuYee]====> detectorDir: %+v\n", detectorDir)
	log.Printf("[DEBUG] loading detectors from: %s", detectorDir)
	detectConfig, err := detect.ParseDir(detectorDir) //如果没有找到定制配置，则从这里开始分析
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	if detectConfig == nil {
		detectConfig = &detect.Config{}
		fmt.Printf("[KuuYee]====> detectConfig: %+v\n", detectConfig)
	}
	err = detectConfig.Merge(&detect.Config{Detectors: c.Detectors})
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	//打印能够发现的类型
	for i, v := range detectConfig.Detectors {
		fmt.Printf("[KuuYee]====> detectConfig %d : %+v\n", i, v)
	}

	// 装载默认Appfile，我们可以合并任何的默认
	// appfile到已经装载的Appfile
	appDef, err := appfile.Default(appPath, detectConfig)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"装载Appfile报错：%s", err))
		return 1
	}
	fmt.Printf("[KuuYee]====> appDef: %+v\n", appDef)

	// 如果没有加载到appfile，那么认为没有可用的应用
	if app == nil && appDef.Application.Type == "" {
		c.Ui.Error(strings.TrimSpace(errCantDetectType))
		return 1
	}

	// 合并应用
	if app != nil {
		if err := appDef.Merge(app); err != nil {
			c.Ui.Error(fmt.Sprintf(
				"装载Appfile报错： %s", err))
			return 1
		}
	}
	app = appDef
	fmt.Printf("[KuuYee]====> app: %+v\n", app)

	// 编译Appfile
	ui.Header("获取所有的Appfile依赖...")
	capp, err := appfile.Compile(app, &appfile.CompileOpts{
		Dir:      filepath.Join(filepath.Dir(app.Path), DefaultOutputDir, DefaultOutputDirCompiledAppfile),
		Detect:   detectConfig,
		Callback: c.compileCallback(ui),
	})
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"编译Appfile报错：%s", err))
		return 1
	}

	// 取得一个Core
	core, err := c.Core(capp)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"装载Core报错：%s", err))
		return 1
	}

	// 取得可用的infrastucture，仅仅为了UI
	infra := app.ActiveInfrastructure()

	// 编译之前，告诉用户what is going on
	ui.Header("编译...")
	ui.Message(fmt.Sprintf(
		"Application:   %s (%s)",
		app.Application.Name,
		app.Application.Type))
	ui.Message(fmt.Sprintf("项目：    %s", app.Project.Name))
	ui.Message(fmt.Sprintf(
		"Infrastructure: %s (%s)",
		infra.Type,
		infra.Flavor))
	ui.Message("")

	// 开始编译
	if err := core.Compile(); err != nil {
		c.Ui.Error(fmt.Sprintf(
			"编译报错: %s", err))
		return 1
	}

	// Success!
	ui.Header("[green]编译成功!")
	ui.Message(fmt.Sprintf(
		"[green]This means that Otto is now ready to start a development environment,\n" +
			"deploy this application, build the supporting infrastructure, and\n" +
			"more. See the help for more information.\n\n" +
			"Supporting files to enable Otto to manage your application from\n" +
			"development to deployment have been placed in the output directory.\n" +
			"These files can be manually inspected to determine what Otto will do."))

	return 0
}

func (c *CompileCommand) Synopsis() string {
	return "Prepares your project for being run."
}

func (c *CompileCommand) Help() string {
	return ""
}

func (c *CompileCommand) compileCallback(ui ui.Ui) func(appfile.CompileEvent) {
	return func(raw appfile.CompileEvent) {}
}

// 返回装载任何appfile.File的拷贝，否则返回nil,自从Otto能够
// 发现app类型，这是有效的。还能够返回appfile所在目录信息，就是
// 当前WD目录(没有发现appfile)
func loadAppfile(flagAppfile string) (*appfile.File, string, error) {
	appfilePath, err := findAppfile(flagAppfile)
	if err != nil {
		return nil, "", err
	}

	if appfilePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, "", err
		}
		return nil, wd, nil
	}
	app, err := appfile.ParseFile(appfilePath) //解析当前项目目录
	if err != nil {
		return nil, "", err
	}
	return app, filepath.Dir(app.Path), nil
}

// findAppfile返回已存在的Appfile所在路径，依赖于检查flag值结果进行判断
// 如果flag是空，返回空
func findAppfile(flag string) (string, error) {
	// 首先，如果在命令行指定了appfile，那么我们就要验证是否存在
	if flag != "" {
		fi, err := os.Stat(flag)
		if err != nil {
			return "", fmt.Errorf("装载Appfile报错：%s", err)
		}

		if fi.IsDir() {
			return findAppfileInDir(flag), nil
		} else {
			return flag, nil
		}
	}

	// 检索当前目录
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("加载当前目录报错: %s", err)
	}

	return findAppfileInDir(wd), nil
}

// findAppfileInDir 返回一个目录，首先查找默认的appfile目录，否则查找
// 其它能找到的appfile
func findAppfileInDir(path string) string {
	if _, err := os.Stat(filepath.Join(path, DefaultAppfile)); err == nil {
		return filepath.Join(path, DefaultAppfile)
	}
	for _, aaf := range AltAppfiles {
		if _, err := os.Stat(filepath.Join(path, aaf)); err == nil {
			return filepath.Join(path, aaf)
		}
	}
	return ""
}

const errCantDetectType = `
No Appfile is present and Otto couldn't detect the project type automatically.
Otto does its best without an Appfile to detect what kind of project this is
automatically, but sometimes this fails if the project is in a structure
Otto doesn't recognize or its a project type that Otto doesn't yet support.

Please create an Appfile and specify at a minimum the project name and type. Below
is an example minimal Appfile specifying the "my-app" application name and "go"
project type:

    application {
	name = "my-app"
	type = "go"
    }

If you believe Otto should've been able to automatically detect your
project type, then please open an issue with the Otto project.
`
