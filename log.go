package main

import (
	"io"
	"os"
)

// 日志的环境变量
const Envlog = "OTTO_LOG"
const EnvLogFile = "OTTOLOG_PATH"

func logOutput() (logOutput io.Writer, err error) {
	logOutput = nil
	if os.Getenv(Envlog) != "" {
		logOutput = os.Stderr

		if logPath := os.Getenv(EnvLogFile); logPath != "" {
			var err error
			logOutput, err = os.Create(logPath)
			if err != nil {
				return nil, err
			}
		}
	}
	return
}
