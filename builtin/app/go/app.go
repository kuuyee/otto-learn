package goapp

import (
	"fmt"
	"github.com/kuuyee/otto-learn/app"
	"strings"
)

// go:generate go-bindata -pkg=goapp -nomemcopy -nometadata ./data/...

// App实现了app.App
type App struct{}

func (a *App) Compile(ctx *app.Context) (*app.CompileResult, error) {
	return nil, nil
}

func (a *App) Build(ctx *app.Context) error {
	return fmt.Errorf(strings.TrimSpace(buildErr))
}

func (a *App) Deploy(ctx *app.Context) error {
	return nil
}

func (a *App) Dev(ctx *app.Context) error {
	return nil
}

func (a *App) DevDep(dst, src *app.Context) (*app.DevDep, error) {
	return nil, nil
}

const buildErr = `
Build isn't supported yet for Go!

Early versions of Otto are focusing on creating a fantastic development
experience. Because of this, build/deploy are still lacking for many
application types. These will be fixed very soon in upcoming versions of
Otto. Sorry!
`
