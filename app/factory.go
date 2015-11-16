package app

import (
	"fmt"
	"reflect"
)

// Factory是工厂函数，用来创建基础设施
type Factory func() (App, error)

// StructFactory 返回一个工厂函数来创建一个
// 拷贝自类型v的新实例
func StructFactory(v App) Factory {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return func() (App, error) {
		raw := reflect.New(t)
		v, ok := raw.Interface().(App)
		if !ok {
			return nil, fmt.Errorf("实例类型错误: %#v", raw.Interface())
		}
		return v, nil
	}
}
