// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"

	"github.com/goki/gear"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

type App struct {

	// the name of the user
	Name string `desc:"the name of the user"`

	// the age of the user
	Age int `desc:"the age of the user"`

	// whether the user likes Go
	LikesGo bool `desc:"whether the user likes Go"`

	// the target platform to build for
	BuildTarget string `desc:"the target platform to build for"`
}

var TheApp App

var TypeApp = kit.Types.AddType(&App{}, AppProps)

var AppProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"BuildCmd", ki.Props{
			"label": "Build",
		}},
	},
}

func main() {
	err := gear.Run(&TheApp, "basic", "Basic is an example app made with Gear", "config.toml")
	if err != nil {
		fmt.Println(err)
	}
}

func (a *App) BuildCmd() error {
	if a.BuildTarget == "" {
		return errors.New("missing build target")
	}
	fmt.Println("Building for platform", a.BuildTarget)
	return nil
}
