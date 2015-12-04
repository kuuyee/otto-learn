package appfile

// File对应单个Appfile的结构
type File struct {
	// ID是标志这个file的UUID，在第一次编译时生成，否则一直是空
	ID string

	// Path 是装载的root文件，如果从io.Reader解析，这个Path可能是空
	Path string

	// Source对于依赖不能为空，对于编译有用
	Source string

	Application    *Application
	Project        *Project
	Infrastructure []*Infrastructure
	Customization  *Customization

	// Imports is the list of imports that this File made. The imports
	// are realized during compilation, but this list won't be cleared
	// in case it wants to be inspected later.
	Import []*Import
}

// Application 是定义App的结构
type Application struct {
	Name         string
	Type         string
	Dependencies []*Dependency `mapstructure:"dependency"`
}

// Customization 是Appfile下Customization分区内容
type Customization struct {
	Type   string
	Config map[string]interface{}
}

// Dependency 是App依赖的另一个Appfile
type Dependency struct {
	Source string
}

// Project 是Project结构，很多app属于其中
type Project struct {
	Name           string
	Infrastructure string
}

// Infrastructure是定义Infrastructure的结构，App运行在其上
type Infrastructure struct {
	Name   string
	Type   string
	Flavor string

	Foundations []*Foundation
}

// Foundation是配置Infrastructure基本的结构单元
type Foundation struct {
	Name   string
	Config map[string]interface{}
}

// Import 导入其他的Appfile
type Import struct {
	Source string
}

// Merge 将合并外部的Appfile，外部Appfile内容将覆盖默认内容
func (f *File) Merge(other *File) error {
	if other.ID != "" {
		f.ID = other.ID
	}
	if other.Path != "" {
		f.Path = other.Path
	}

	// Application
	if f.Application == nil {
		f.Application = other.Application
	} else if other.Application != nil {
		// Note this won't copy dependencies properly
		f.Application.Merge(other.Application)
	}

	// Project
	if f.Project == nil {
		f.Project = other.Project
	} else if other.Project != nil {
		// Note this won't copy dependencies properly
		*f.Project = *other.Project
	}

	// Infrastructure
	infraMap := make(map[string]int)
	for i, infra := range f.Infrastructure {
		infraMap[infra.Name] = i
	}

	for _, i := range other.Infrastructure {
		idx, ok := infraMap[i.Name]
		if !ok {
			f.Infrastructure = append(f.Infrastructure, i)
			continue
		}

		old := f.Infrastructure[idx]
		if len(i.Foundations) == 0 {
			i.Foundations = old.Foundations
		}

		f.Infrastructure[idx] = i
	}

	// TODO: customizations
	f.Customization = other.Customization

	return nil
}

func (app *Application) Merge(other *Application) {
	if other.Name != "" {
		app.Name = other.Name
	}
	if other.Type != "" {
		app.Type = other.Type
	}
	if len(other.Dependencies) > 0 {
		app.Dependencies = other.Dependencies
	}
}
