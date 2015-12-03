package detect

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl"
	hclobj "github.com/hashicorp/hcl/hcl"
	"github.com/mitchellh/mapstructure"
)

// Parse 解析发现的配置
//
// 由于当前内部限制，配置文件的实体内容在解析前会先
// 拷贝到内存
func Parse(r io.Reader) (*Config, error) {
	// 在使用HCL进行解析之前，首先拷贝文件内容到内存缓冲
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}

	// 解析Buffer
	obj, err := hcl.Parse(buf.String())
	if err != nil {
		return nil, fmt.Errorf("解析错误：%s", err)
	}
	buf.Reset()

	var result Config

	// 解析
	if o := obj.Get("detect", false); o != nil {
		if err := parseDetect(&result, o); err != nil {
			return nil, fmt.Errorf("解析'import'错误: %s", err)
		}
	}
	return &result, nil
}

// ParseFile 分析发现的单个配置
func ParseFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// ParseDir解析目录下所有".hcl"为后缀的文件，按照字母顺序
func ParseDir(path string) (*Config, error) {
	// 读取所有文件
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) { //判断这个错误是不是文件不存在类型的错
			return nil, nil
		}
		return nil, err
	}

	//func (f *File) Readdirnames(n int) (names []string, err error)
	//Readdir读取目录f的内容，返回一个有n个成员的[]string，切片成员为目录中文件对象的名字，
	//采用目录顺序。对本函数的下一次调用会返回上一次调用剩余未读取的内容的信息。

	//如果n>0，Readdir函数会返回一个最多n个成员的切片。这时，如果Readdir返回一个空切片，它会返回
	//一个非nil的错误说明原因。如果到达了目录f的结尾，返回值err会是io.EOF。

	//如果n<=0，Readdir函数返回目录中剩余所有文件对象的名字构成的切片。此时，如果Readdir
	//调用成功（读取所有内容直到结尾），它会返回该切片和nil的错误值。如果在到达结尾前遇到
	//错误，会返回之前成功读取的名字构成的切片和该错误。
	files, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	// 排序
	sort.Strings(files)

	// 逐个检查，解析和合并
	var result Config
	for _, f := range files {
		// 只处理HCL文件
		if filepath.Ext(f) != ".hcl" {
			continue
		}

		// 如果是目录，忽略
		path := filepath.Join(path, f)
		fi, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			continue
		}

		// 解析
		current, err := ParseFile(path)
		if err != nil {
			return nil, fmt.Errorf("解析错误 %s : %s", path, err)
		}

		// 合并
		if err := result.Merge(current); err != nil {
			return nil, fmt.Errorf("合并报错 %s : %s", path, err)
		}
	}

	return &result, nil
}

func parseDetect(result *Config, obj *hclobj.Object) error {
	// 从map中获取所有实际对象的key值
	objects := make([]*hclobj.Object, 0, 2)
	for _, o1 := range obj.Elem(false) {
		for _, o2 := range o1.Elem(true) {
			objects = append(objects, o2)
		}
	}

	if len(objects) == 0 {
		return nil
	}

	// 检查每个对象，返回实际结果
	collection := make([]*Detector, 0, len(objects))
	for _, o := range objects {
		var m map[string]interface{}
		if err := hcl.DecodeObject(&m, o); err != nil {
			return err
		}

		var d Detector
		if err := mapstructure.WeakDecode(m, &d); err != nil {
			return fmt.Errorf("解析detector错误 '%s' : %s", o.Key, err)
		}

		d.Type = o.Key
		collection = append(collection, &d)
	}
	result.Detectors = collection
	return nil
}
