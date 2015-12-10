package appfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuuyee/otto-learn/helper/oneline"
	"github.com/kuuyee/otto-learn/helper/uuid"
)

const (
	IDFile = ".ottoid"
)

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
	Imports []*Import
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

// hasID 检查是否有ID文件，如果文件系统错误则直接返回
func (f *File) hasID() (bool, error) {
	path := filepath.Join(filepath.Dir(f.Path), IDFile)
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	return err == nil, nil
}

// initID 创建一个新的UUID并写入文件，这个会覆盖任何原来的ID文件
func (f *File) initID() error {
	path := filepath.Join(filepath.Dir(f.Path), IDFile)
	uuid := uuid.GenerateUUID()
	data := strings.TrimSpace(fmt.Sprintf(idFileTemplate, uuid)) + "\n"
	return ioutil.WriteFile(path, []byte(data), 0644)
}

// loadID 获取这个文件的ID
func (appF *File) loadID() error {
	hasID, err := appF.hasID()
	if err != nil {
		return err
	}
	if !hasID {
		appF.ID = ""
		return nil
	}

	path := filepath.Join(filepath.Dir(appF.Path), IDFile)
	uuid, err := oneline.Read(path)
	if err != nil {
		return err
	}
	appF.ID = uuid
	return nil
}

//-------------------------------------------------------------------
// Helper Methods
//-------------------------------------------------------------------

// ActiveInfrastructure返回一个在Appfile使用的Infrastructure
func (f *File) ActiveInfrastructure() *Infrastructure {
	for _, i := range f.Infrastructure {
		if i.Name == f.Project.Infrastructure {
			return i
		}
	}
	return nil
}

const idFileTemplate = `
%s

DO NOT MODIFY OR DELETE THIS FILE!

This file should be checked in to version control. Do not ignore this file.

The first line is a unique UUID that represents the Appfile in this directory.
This UUID is used globally across your projects to identify this specific
Appfile. This UUID allows you to modify the name of an application, or have
duplicate application names without conflicting.

If you delete this file, then deploys may duplicate this application since
Otto will be unable to tell that the application is deployed.
`
