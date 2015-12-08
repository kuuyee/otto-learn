package oneline

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// Read 只从给定文件中读取第一行，去除两侧的空白
func Read(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	result, err := bufio.NewReader(f).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	return strings.TrimSpace(result), nil
}
