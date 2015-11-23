package command

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/hashicorp/otto/ui"
	"github.com/hashicorp/vault/helper/password"
	"github.com/mitchellh/cli"
)

var defaultInputReader io.Reader
var defaultInputWriter io.Writer

// 返回一个otto UI实现,封装cli.Ui
func NewUi(raw cli.Ui) ui.Ui {
	return &ui.Styled{
		Ui: &cliUi{
			CliUi: raw,
		},
	}
}

// cliUi是cli.Ui的封装并实现了otto.Ui接口。自从有了NewUI方法，
// 它就成为一个非导出结构
type cliUi struct {
	CliUi cli.Ui

	// 用于Input的读写
	Reader io.Reader
	Writer io.Writer

	interrupted bool
	l           sync.Mutex
}

func (u *cliUi) Header(msg string) {
	u.CliUi.Output(ui.Colorize(msg))
}

func (u *cliUi) Message(msg string) {
	u.CliUi.Output(ui.Colorize(msg))
}

func (u *cliUi) Raw(msg string) {
	fmt.Print(msg)
}

func (i *cliUi) Input(opts *ui.InputOpts) (string, error) {
	// 如何设置了环境变量，我们就不询问提示input
	if value := opts.EnvVarValue(); value != "" {
		return value, nil
	}

	r := i.Reader
	w := i.Writer
	if r == nil {
		r = defaultInputReader
	}
	if w == nil {
		w = defaultInputWriter
	}
	if r == nil {
		r = os.Stdin
	}
	if w == nil {
		w = os.Stdout
	}

	// 确保我们只在第一次询问input.Terraform应该确保这个
	// 但是不会破坏验证
	i.l.Lock()
	defer i.l.Unlock()

	// 如果终止，那么就不询问input
	if i.interrupted {
		return "", errors.New("中断")
	}

	// 监听中断操作，那么就不再询问input
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	// 格式化询问输出
	var buf bytes.Buffer
	buf.WriteString("[reset]")
	buf.WriteString(fmt.Sprintf("[bold]%s[reset]\n", opts.Query))
	if opts.Description != "" {
		s := bufio.NewScanner(strings.NewReader(opts.Description))
		for s.Scan() {
			buf.WriteString(fmt.Sprintf("   %s\n", s.Text()))
		}
		buf.WriteString("\n")
	}
	if opts.Default != "" {
		buf.WriteString("  [bold]Default:[reset] ")
		buf.WriteString(opts.Default)
		buf.WriteString("\n")
	}
	buf.WriteString("  [bold]输入一个值：[reset] ")

	// 询问用户输入
	if _, err := fmt.Fprint(w, ui.Colorize(buf.String())); err != nil {
		return "", err
	}

	// 监听goroutine输入。这允许我们中断
	result := make(chan string, 1)
	if opts.Hide {
		f, ok := r.(*os.File)
		if !ok {
			return "", fmt.Errorf("必须读取一个文件")
		}

		line, err := password.Read(f)
		if err != nil {
			return "", err
		}
		result <- line
	} else {
		go func() {
			var line string
			if _, err := fmt.Fscanln(r, &line); err != nil {
				log.Printf("[ERR] UIInput扫描错误: %s", err)
			}

			result <- line
		}()
	}

	select {
	case line := <-result:

		fmt.Fprint(w, "\n")

		if line == "" {
			line = opts.Default
		}
		return line, nil

	case <-sigCh:
		// 新起一行
		fmt.Fprintln(w)

		// Mark that we were interrupted so future Ask calls fail.
		i.interrupted = true

		return "", errors.New("中断")
	}
}
