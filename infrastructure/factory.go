package infrastructure

import (
	"fmt"
	"reflect"
)

// Factory是创建infrastructures的工厂函数
type Factory func() (Infrastructure, error)

// StructFactory返回一个工厂函数用来创建一个
// 拷贝自类型v的新实例
func StructFactory(v Infrastructure) Factory {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return func() (Infrastructure, error) {
		raw := reflect.New(t)
		v, ok := raw.Interface().(Infrastructure)
		if !ok {
			return nil, fmt.Errorf("错误的实例类型：%#v", raw.Interface())
		}
		return v, nil
	}
}
