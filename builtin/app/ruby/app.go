package rubyapp

import (
	"github.com/kuuyee/otto-learn/app"
	_ "strings"
)

// App是app.App接口的Ruby版实现
type App struct{}

func (a *App) Compile(ctx *app.Context) (*app.CompileResult, error) {
	return nil, nil
}

func (a *App) Build(ctx *app.Context) error {
	return nil
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

const devInstructions = `
A development environment has been created for writing a generic
Ruby-based app.

Ruby is pre-installed. To work on your project, edit files locally on your
own machine. The file changes will be synced to the development environment.

When you're ready to build your project, run 'otto dev ssh' to enter
the development environment. You'll be placed directly into the working
directory where you can run 'bundle' and 'ruby' as you normally would.

You can access any running web application using the IP above.
`
