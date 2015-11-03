package foundation

import (
	"fmt"
	"reflect"
)

// Factory 是工厂函数，用来创建foundation
type Factory func() (Foundation, error)

// StructFactory 返回一个工厂函数，用来创建一个v 类型拷贝的实例
func StructFactory(v Foundation) Factory {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return func() (Foundation, error) {
		raw := reflect.New(t)
		v, ok := raw.Interface().(Foundation)
		if !ok {
			return nil, fmt.Errorf("实例类型错误：%#v", raw.Interface())
		}
		return v, nil
	}
}
