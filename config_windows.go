package main

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	shell         = syscall.MustLoadDLL("Shell32.dll")
	getFolderPath = shell.MustFindProc("SHGetFolderPathW")
)

const CSIDL_APPDATA = 26

func configFile() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", err
	}
	fmt.Printf("[KuuYee]====> homedir: %s",dir)
	return filepath.Join(dir, "otto.rc"), nil
}

func configDir() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", err
	}
	getdir := filepath.Join(dir, "otto.d")
	fmt.Println(getdir)
	return getdir, nil
}

func homeDir() (string, error) {
	b := make([]uint16, syscall.MAX_PATH)

	r, _, err := getFolderPath.Call(0, CSIDL_APPDATA, 0, 0, uintptr(unsafe.Pointer(&b[0])))
	if uint32(r) != 0 {
		return "", err
	}
	home := syscall.UTF16ToString(b)
	fmt.Printf("%v\n", syscall.MAX_PATH)
	return home, nil
}
