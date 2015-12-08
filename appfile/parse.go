package appfile

import (
	"io"
	"os"
	"path/filepath"
)

func Parse(r io.Reader) (*File, error) {
	return nil, nil
}

// ParseFile 解析Appfile
func ParseFile(path string) (*File, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result, err := Parse(f)
	if result != nil {
		result.Path = path
		if err := result.loadID(); err != nil {
			return nil, err
		}
	}
	return result, err
}
