package foundation

import (
	"fmt"
)

// Tuple 用来在Appfile中查阅foundation实现的配置
// 这个结果通常用在飞指针表单，是map的一个key
type Tuple struct {
	Type        string // Type是foundation的类型，比如"consule"
	Infra       string // Infra是一个基础设施的类型，比如"aws"
	InfraFlavor string // InfraFlavor是AWS VPC(Virtual Private Cloud)的配置
}

func (t *Tuple) String() string {
	return fmt.Sprintf("(%q, %q, %q)", t.Type, t.Infra, t.InfraFlavor)
}

// TupleSlice 是实现了sort.Interface接口的[]Tuple
// 可以在tuple_test.go中查看排序
type TupleSlice []Tuple

// sort.Interface impl
func (s TupleSlice) Len() int      { return len(s) }
func (s TupleSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TupleSlice) Less(i, j int) bool {
	if s[i].Type != s[j].Type {
		return s[i].Type < s[j].Type
	}
	if s[i].Infra != s[j].Infra {
		return s[i].Infra < s[j].Infra
	}
	if s[i].InfraFlavor != s[j].InfraFlavor {
		return s[i].InfraFlavor < s[j].InfraFlavor
	}
	return false
}

// Map把TupleSlice灌入一个map,使每个tuple对应一个工厂函数
func (s TupleSlice) Map(f Factory) TupleMap {
	m := make(TupleMap, len(s))
	for _, t := range s {
		m[t] = f
	}
	return m
}

// TupleMap是一些app tuple附加的辅助方法
type TupleMap map[Tuple]Factory

// Lookup查询一个Tuple，用来替代直接的[]数组访问
// since it respects wildcards ('*') within the Tuple.
func (m TupleMap) Lookup(t Tuple) Factory {
	// 如果存在直接返回
	if f, ok := m[t]; ok {
		return f
	}

	// 嗯，这不复杂
	for h, f := range m {
		if h.Type != "*" && h.Type != t.Type {
			continue
		}

		if h.Infra != "*" && h.Infra != t.Infra {
			continue
		}
		if h.InfraFlavor != "*" && h.InfraFlavor != t.InfraFlavor {
			continue
		}
		return f
	}
	return nil
}

// Add是一个辅助方法，在map中加入其它的map
func (m TupleMap) Add(m2 TupleMap) {
	for k, v := range m2 {
		m[k] = v
	}
}
