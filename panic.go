package main

import (
	"fmt"
	"github.com/mitchellh/panicwrap"
	"io"
	"os"
	"strings"
)

// 当发生一个panic是会输出下面的头内容
const panicOutput = `
!!!!!!!!!!!!!!!!!!!!!!!!!!! KuuYee OTTO CRASH !!!!!!!!!!!!!!!!!!!!!!!!!!!!

Otto crashed! This is always indicative of a bug within Otto.
A crash log has been placed at "crash.log" relative to your current
working directory. It would be immensely helpful if you could please
report the crash with Otto[1] so that we can fix this.

When reporting bugs, please include your otto version. That
information is available on the first line of crash.log. You can also
get it by running 'otto -version' on the command line.

[1]: https://github.com/hashicorp/otto/issues

!!!!!!!!!!!!!!!!!!!!!!!!!!! KuuYee OTTO CRASH !!!!!!!!!!!!!!!!!!!!!!!!!!!!
`

// panicwrap的处理函数
func panicHandler(logF *os.File) panicwrap.HandlerFunc {
	return func(m string) {

		fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", m))

		// 创建Log文件
		f, err := os.Create("crash.log")
		if err != nil {
			fmt.Fprintf(os.Stderr, "创建log文件报错：", err)
			return
		}
		defer f.Close()

		// 这是偏移未知到Log文件头部
		if _, err = logF.Seek(0, 0); err != nil {
			fmt.Fprintf(os.Stderr, "设置偏移位报错：", err)
			return
		}

		// 复制内容到log文件，包括刚产生的panic
		if _, err = io.Copy(f, logF); err != nil {
			fmt.Fprintf(os.Stderr, "写入log文件报错：", err)
			return
		}

		// 输出一些帮助信息
		fmt.Printf("\n\n")
		fmt.Println(strings.TrimSpace(panicOutput))
	}
}
