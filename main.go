package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/panicwrap"
	"github.com/mitchellh/prefixedio"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	//使用panicwrap封装panic
	var wrapConfig panicwrap.WrapConfig
	wrapConfig.CookieKey = "OTTO_PANICWRAP_COOKIE"
	wrapConfig.CookieValue = fmt.Sprintf("otto-%s-%s-%s", Version, VersionPrerelease, GitCommit)
	//fmt.Printf("warpConfig : %+v\n", wrapConfig)

	if !panicwrap.Wrapped(&wrapConfig) {
		fmt.Println("没有封装wrapConfig配置")
		logWriter, err := logOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "不能设置Log输出到：%s", err)
			return 1
		}
		if logWriter == nil {
			logWriter = ioutil.Discard //Discard实现了io.Writer接口
		}

		//当发生panic时我们发送log到临时文件，否则就删除
		logTempFile, err := ioutil.TempFile("", "kuuyee-otto-log")
		if err != nil {
			fmt.Fprintf(os.Stderr, "不能设置Log临时文件：%s", err)
			return 1
		}
		defer os.Remove(logTempFile.Name())
		defer logTempFile.Close()

		// 告诉logger把日志输出到这个文件
		os.Setenv(Envlog, "")
		os.Setenv(EnvLogFile, "")

		// 设置发送到stdout/stderr数据的读取前缀
		doneCh := make(chan struct{})
		outR, outW := io.Pipe()
		go copyOutput(outR, doneCh)

		// 创建panicwrap配置
		wrapConfig.Handler = panicHandler(logTempFile)
		wrapConfig.Writer = io.MultiWriter(logTempFile, logWriter)
		wrapConfig.Stdout = outW
		exitStatus, err := panicwrap.Wrap(&wrapConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "不能启动Otto: %s", err)
			return 1
		}

		// 如果>=0，那么我们是父进程，所以退出
		if exitStatus >= 0 {
			outW.Close()

			<-doneCh

			return exitStatus
		}

		// 我们是子，所以关闭tempfile
		logTempFile.Close()

	}

	fmt.Printf("[KuuYee_DEBUG]====> %s\n", "main.go/realMain/81")
	//调用真正的main函数
	return wrappedMain()
}

func wrappedMain() int {
	log.SetOutput(os.Stderr)
	log.Printf("[INFO] KuuYee Otto version: %s %s %s", Version, VersionPrerelease, GitCommit)

	// 设置信号处理器
	initSignalHandlers()

	// 载入配置
	config := BuiltinConfig
	fmt.Printf("[KuuYee]====> config: %+v",config)

	// 运行检查点
	go runCheckpoint(&config)

	// 获取命令行参数。通过"--version"和"-v"显示版本
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-v" || arg == "-version" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}
	cli := &cli.CLI{
		Args:     args,
		Commands: Commands,
		HelpFunc: cli.FilteredHelpFunc(
			CommandsInclude, cli.BasicHelpFunc("otto")),
		HelpWriter: os.Stdout,
	}

	exitCode, err := cli.Run()
	if err != nil {
		Ui.Error(fmt.Sprintf("Error executing CLI: %s", err.Error()))
		return 1
	}

	return exitCode
}

// copyOutput
func copyOutput(r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)

	pr, err := prefixedio.NewReader(r)
	if err != nil {
		panic(err)
	}

	pr.FlushTimeout = 5 * time.Millisecond

	stderrR, err := pr.Prefix(ErrorPrefix)
	if err != nil {
		panic(err)
	}

	stdoutR, err := pr.Prefix(OutputPrefix)
	if err != nil {
		panic(err)
	}

	defaultR, err := pr.Prefix("")
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		io.Copy(os.Stderr, stderrR)
	}()

	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, stdoutR)
	}()

	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, defaultR)
	}()

	wg.Wait()
}
