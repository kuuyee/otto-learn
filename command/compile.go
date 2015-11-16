package command

import (
	_ "fmt"
	"github.com/hashicorp/otto/appfile"
	"github.com/hashicorp/otto/appfile/detect"
	"github.com/hashicorp/otto/ui"
)

// CompileCommand是一个编译命令，要来把
// Appfile编译成一组数据供其他命令使用
type CompileCommand struct {
	Meta
	Detectors []*detect.Detector
}

func (c *CompileCommand) Run(args []string) int {
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
