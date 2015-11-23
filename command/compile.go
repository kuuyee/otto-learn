package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/otto/appfile"
	"github.com/hashicorp/otto/appfile/detect"
	"github.com/hashicorp/otto/ui"
	"path/filepath"
)

// CompileCommand是一个编译命令，要来把
// Appfile编译成一组数据供其他命令使用
type CompileCommand struct {
	Meta
	Detectors []*detect.Detector
}

func (c *CompileCommand) Run(args []string) int {
	var flagAppfile string
	fs := c.FlagSet("compile", FlagSetNone)
	fs.Usage = func() { c.Ui.Error(c.Help()) }
	fs.StringVar(&flagAppfile, "appfile", "", "")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Load a UI
	ui := c.OttoUi()
	ui.Header("装载 Appfile...")

	app, appPath, err := loadAppfile(flagAppfile)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	// 如果没有Appfile，告诉用户发生了什么
	if app == nil {
		ui.Header("No Appfile found! Detecting project information...")
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
	detectorDir := filepath.Join(dataDir, DefaultLocalDataDetectorDir)
	log.Printf("[DEBUG] loading detectors from: %s", detectorDir)
	detectConfig, err := detect.ParseDir(detectorDir)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	if detectConfig == nil {
		detectConfig = &detect.Config{}
	}
	err = detectConfig.Merge(&detect.Config{Detectors: c.Detectors})
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	// 装载默认Appfile，我们可以合并任何的默认
	// appfile到已经装载的Appfile
	appDef, err := appfile.Default(appPath, detectConfig)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"装载Appfile报错：%s", err))
		return 1
	}

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
func loadAppfile(flagAppfile string) (*appfile.File, string, error) {
	return nil, "", nil
}

func findAppfile(flag string) (string, error) {
	return "", nil
}

func findAppfileInDir(path string) string {
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
