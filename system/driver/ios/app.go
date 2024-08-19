// Copyright 2023 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ios

// Package ios implements system interfaces on iOS mobile devices
package ios

import (
	"log"
	"os/user"
	"path/filepath"
	"runtime"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/system"
	"cogentcore.org/core/system/driver/base"
	"cogentcore.org/core/vgpu"
	"cogentcore.org/core/vgpu/vdraw"
	vk "github.com/goki/vulkan"
)

func Init() {
	// Lock the goroutine responsible for initialization to an OS thread.
	// This means the goroutine running main (and calling the run function
	// below) is locked to the OS thread that started the program. This is
	// necessary for the correct delivery of UIKit events to the process.
	//
	// A discussion on this topic:
	// https://groups.google.com/forum/#!msg/golang-nuts/IiWZ2hUuLDA/SNKYYZBelsYJ
	runtime.LockOSThread()

	system.OnSystemWindowCreated = make(chan struct{})
	TheApp.InitVk()
	base.Init(TheApp, &TheApp.App)
}

// TheApp is the single [system.App] for the iOS platform
var TheApp = &App{AppSingle: base.NewAppSingle[*vdraw.Drawer, *Window]()}

// App is the [system.App] implementation for the iOS platform
type App struct {
	base.AppSingle[*vdraw.Drawer, *Window]

	// GPU is the system GPU used for the app
	GPU *vgpu.GPU

	// Winptr is the pointer to the underlying system window
	Winptr uintptr
}

// InitVk initializes Vulkan things for the app
func (a *App) InitVk() {
	err := vk.SetDefaultGetInstanceProcAddr()
	if err != nil {
		// TODO(kai): maybe implement better error handling here
		log.Fatalln("system/driver/ios.App.InitVk: failed to set Vulkan DefaultGetInstanceProcAddr")
	}
	err = vk.Init()
	if err != nil {
		log.Fatalln("system/driver/ios.App.InitVk: failed to initialize vulkan")
	}

	winext := vk.GetRequiredInstanceExtensions()
	a.GPU = vgpu.NewGPU()
	a.GPU.AddInstanceExt(winext...)
	a.GPU.Config(a.Name())
}

// DestroyVk destroys vulkan things (the drawer and surface of the window) for when the app becomes invisible
func (a *App) DestroyVk() {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	vk.DeviceWaitIdle(a.Draw.Surf.Device.Device)
	a.Draw.Destroy()
	a.Draw.Surf.Destroy()
	a.Draw = nil
}

// FullDestroyVk destroys all vulkan things for when the app is fully quit
func (a *App) FullDestroyVk() {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.GPU.Destroy()
}

// NewWindow creates a new window with the given options.
// It waits for the underlying system window to be created first.
// Also, it hides all other windows and shows the new one.
func (a *App) NewWindow(opts *system.NewWindowOptions) (system.Window, error) {
	defer func() { system.HandleRecover(recover()) }()
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.Win = &Window{base.NewWindowSingle(a, opts)}
	a.Win.This = a.Win
	a.Event.Window(events.WinShow)
	a.Event.Window(events.WinFocus)

	a.Event.WindowResize()
	a.Event.WindowPaint()

	go a.Win.WinLoop()

	return a.Win, nil
}

// SetSystemWindow sets the underlying system window pointer, surface, system, and drawer.
// It should only be called when [App.Mu] is already locked.
func (a *App) SetSystemWindow(winptr uintptr) error {
	defer func() { system.HandleRecover(recover()) }()
	var vsf vk.Surface
	// we have to remake the surface, system, and drawer every time someone reopens the window
	// because the operating system changes the underlying window
	ret := vk.CreateWindowSurface(a.GPU.Instance, winptr, nil, &vsf)
	if err := vk.Error(ret); err != nil {
		return err
	}
	sf := vgpu.NewSurface(a.GPU, vsf)

	sys := a.GPU.NewGraphicsSystem(a.Name(), &sf.Device)
	sys.ConfigRender(&sf.Format, vgpu.UndefType)
	sf.SetRender(&sys.Render)
	// sys.Mem.Vars.NDescs = vgpu.MaxTexturesPerSet
	sys.Config()
	a.Draw = &vdraw.Drawer{
		Sys:     *sys,
		YIsDown: true,
	}
	// a.Draw.ConfigSys()
	a.Draw.ConfigSurface(sf, vgpu.MaxTexturesPerSet)

	a.Winptr = winptr

	// if the window already exists, we are coming back to it, so we need to show it
	// again and send a screen update
	if a.Win != nil {
		a.Event.Window(events.WinShow)
		a.Event.Window(events.ScreenUpdate)
	}
	return nil
}

func (a *App) DataDir() string {
	usr, err := user.Current()
	if errors.Log(err) != nil {
		return "/tmp"
	}
	return filepath.Join(usr.HomeDir, "Library")
}

func (a *App) Platform() system.Platforms {
	return system.IOS
}

func (a *App) OpenURL(url string) {
	// TODO(kai): implement OpenURL on iOS
}

func (a *App) Clipboard(win system.Window) system.Clipboard {
	return &Clipboard{}
}
