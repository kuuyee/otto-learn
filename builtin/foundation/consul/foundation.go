package consul

import (
	"github.com/kuuyee/otto-learn/foundation"
)

// go:generate go-bindate -pkg=consul -nomemcopy -nometadata ./data/...

// Foundation实现了foundation.Foundation
type Foundation struct{}

func (f *Foundation) Compile(ctx *foundation.Context) (*foundation.CompileResult, error) {
	return nil, nil
}

func (f *Foundation) Infra(ctx *foundation.Context) error {
	return nil
}
