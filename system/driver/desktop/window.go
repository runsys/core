// Copyright 2019 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// originally based on golang.org/x/exp/shiny:
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package desktop

import (
	"image"
	"log"

	"cogentcore.org/core/events"
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/gpu/gpudraw"
	"cogentcore.org/core/system"
	"cogentcore.org/core/system/driver/base"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Window is the implementation of [system.Window] for the desktop platform.
type Window struct {
	base.WindowMulti[*App, *gpudraw.Drawer]

	// Glw is the glfw window associated with this window
	Glw *glfw.Window

	// ScreenName is the name of the last known screen this window was on
	ScreenWindow string
}

func (w *Window) IsVisible() bool {
	return w.WindowMulti.IsVisible() && w.Glw != nil
}

// Activate() sets this window as the current render target for gpu rendering
// functions, and the current context for gpu state (equivalent to
// MakeCurrentContext on OpenGL).
// If it returns false, then window is not visible / valid and
// nothing further should happen.
// Must call this on app main thread using system.TheApp.RunOnMain
//
//	system.TheApp.RunOnMain(func() {
//	   if !win.Activate() {
//	       return
//	   }
//	   // do GPU calls here
//	})
func (w *Window) Activate() bool {
	// note: activate is only run on main thread so we don't need to check for mutex
	if w == nil || w.Glw == nil {
		return false
	}
	w.Glw.MakeContextCurrent()
	return true
}

// DeActivate() clears the current render target and gpu rendering context.
// Generally more efficient to NOT call this and just be sure to call
// Activate where relevant, so that if the window is already current context
// no switching is required.
// Must call this on app main thread using system.TheApp.RunOnMain
func (w *Window) DeActivate() {
	glfw.DetachCurrentContext()
}

// NewGlfwWindow makes a new glfw window.
// It must be run on main.
func NewGlfwWindow(opts *system.NewWindowOptions, sc *system.Screen) (*glfw.Window, error) {
	_, _, tool, fullscreen := system.WindowFlagsToBool(opts.Flags)
	// glfw.DefaultWindowHints()
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.Visible, glfw.False) // needed to position
	glfw.WindowHint(glfw.Focused, glfw.True)
	// glfw.WindowHint(glfw.ScaleToMonitor, glfw.True)
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	// glfw.WindowHint(glfw.Samples, 0) // don't do multisampling for main window -- only in sub-render
	if fullscreen {
		glfw.WindowHint(glfw.Maximized, glfw.True)
	}
	if tool {
		glfw.WindowHint(glfw.Decorated, glfw.False)
	} else {
		glfw.WindowHint(glfw.Decorated, glfw.True)
	}
	// glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	// todo: glfw.Floating for always-on-top -- could set for modal
	sz := opts.Size
	if TheApp.Platform() == system.MacOS {
		// on macOS, the size we pass to glfw must be in dots
		sz = sc.WinSizeFromPix(opts.Size)
	}
	win, err := glfw.CreateWindow(sz.X, sz.Y, opts.GetTitle(), nil, nil)
	if err != nil {
		return win, err
	}

	if !fullscreen {
		win.SetPos(opts.Pos.X, opts.Pos.Y)
	}
	if opts.Icon != nil {
		win.SetIcon(opts.Icon)
	}
	return win, err
}

// Screen gets the screen of the window, computing various window parameters.
func (w *Window) Screen() *system.Screen {
	if w == nil || w.Glw == nil {
		return TheApp.Screens[0]
	}
	w.Mu.Lock()
	defer w.Mu.Unlock()

	var sc *system.Screen
	mon := w.Glw.GetMonitor() // this returns nil for windowed windows -- i.e., most windows
	// that is super useless it seems. only works for fullscreen
	if mon != nil {
		if MonitorDebug {
			log.Printf("MonitorDebug: desktop.Window.Screen: %v: got screen: %v\n", w.Nm, mon.GetName())
		}
		sc = TheApp.ScreenByName(mon.GetName())
		if sc == nil {
			log.Printf("MonitorDebug: desktop.Window.Screen: could not find screen of name: %v\n", mon.GetName())
			sc = TheApp.Screens[0]
		}
		goto setScreen
	}
	sc = w.GetScreenOverlap()
	// if monitorDebug {
	// 	log.Printf("MonitorDebug: desktop.Window.GetScreenOverlap: %v: got screen: %v\n", w.Nm, sc.Name)
	// }
setScreen:
	w.ScreenWindow = sc.Name
	w.PhysDPI = sc.PhysicalDPI
	w.DevicePixelRatio = sc.DevicePixelRatio
	if w.LogDPI == 0 {
		w.LogDPI = sc.LogicalDPI
	}
	return sc
}

// GetScreenOverlap gets the monitor for given window
// based on overlap of geometry, using limited glfw 3.3 api,
// which does not provide this functionality.
// See: https://github.com/glfw/glfw/issues/1699
// This is adapted from slawrence2302's code posted there.
func (w *Window) GetScreenOverlap() *system.Screen {
	w.App.Mu.Lock()
	defer w.App.Mu.Unlock()

	var wgeom image.Rectangle
	wgeom.Min.X, wgeom.Min.Y = w.Glw.GetPos()
	var sz image.Point
	sz.X, sz.Y = w.Glw.GetSize()
	wgeom.Max = wgeom.Min.Add(sz)

	var csc *system.Screen
	var ovlp int
	for _, sc := range TheApp.Screens {
		isect := sc.Geometry.Intersect(wgeom).Size()
		ov := isect.X * isect.Y
		if ov > ovlp || ovlp == 0 {
			csc = sc
			ovlp = ov
		}
	}
	return csc
}

func (w *Window) Position() image.Point {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	if w.Glw == nil {
		return w.Pos
	}
	var ps image.Point
	ps.X, ps.Y = w.Glw.GetPos()
	w.Pos = ps
	return ps
}

func (w *Window) SetTitle(title string) {
	if w.IsClosed() {
		return
	}
	w.Titl = title
	w.App.RunOnMain(func() {
		if w.Glw == nil { // by time we got to main, could be diff
			return
		}
		w.Glw.SetTitle(title)
	})
}

func (w *Window) SetIcon(images []image.Image) {
	if w.IsClosed() {
		return
	}
	w.App.RunOnMain(func() {
		if w.Glw == nil { // by time we got to main, could be diff
			return
		}
		w.Glw.SetIcon(images)
	})
}

func (w *Window) SetWinSize(sz image.Point) {
	if w.IsClosed() || w.Is(system.Fullscreen) {
		return
	}
	// note: anything run on main only doesn't need lock -- implicit lock
	w.App.RunOnMain(func() {
		if w.Glw == nil { // by time we got to main, could be diff
			return
		}
		w.Glw.SetSize(sz.X, sz.Y)
	})
}

func (w *Window) SetPos(pos image.Point) {
	if w.IsClosed() || w.Is(system.Fullscreen) {
		return
	}
	// note: anything run on main only doesn't need lock -- implicit lock
	w.App.RunOnMain(func() {
		if w.Glw == nil { // by time we got to main, could be diff
			return
		}
		w.Glw.SetPos(pos.X, pos.Y)
	})
}

func (w *Window) SetGeom(pos image.Point, sz image.Point) {
	if w.IsClosed() || w.Is(system.Fullscreen) {
		return
	}
	sc := w.Screen()
	sz = sc.WinSizeFromPix(sz)
	// note: anything run on main only doesn't need lock -- implicit lock
	w.App.RunOnMain(func() {
		if w.Glw == nil { // by time we got to main, could be diff
			return
		}
		w.Glw.SetSize(sz.X, sz.Y)
		w.Glw.SetPos(pos.X, pos.Y)
	})
}

func (w *Window) Show() {
	if w.IsClosed() {
		return
	}
	// note: anything run on main only doesn't need lock -- implicit lock
	w.App.RunOnMain(func() {
		if w.Glw == nil { // by time we got to main, could be diff
			return
		}
		w.Glw.Show()
	})
}

func (w *Window) Raise() {
	if w.IsClosed() {
		return
	}
	// note: anything run on main only doesn't need lock -- implicit lock
	w.App.RunOnMain(func() {
		if w.Glw == nil { // by time we got to main, could be diff
			return
		}
		if w.Is(system.Minimized) {
			w.Glw.Restore()
		} else {
			w.Glw.Focus()
		}
	})
}

func (w *Window) Minimize() {
	if w.IsClosed() {
		return
	}
	// note: anything run on main only doesn't need lock -- implicit lock
	w.App.RunOnMain(func() {
		if w.Glw == nil { // by time we got to main, could be diff
			return
		}
		w.Glw.Iconify()
	})
}

func (w *Window) Close() {
	if w == nil {
		return
	}
	w.Window.Close()

	w.Mu.Lock()
	defer w.Mu.Unlock()

	w.App.RunOnMain(func() {
		w.Draw.System.WaitDone()
		if w.DestroyGPUFunc != nil {
			w.DestroyGPUFunc()
		}
		var surf *gpu.Surface
		if sf, ok := w.Draw.Renderer().(*gpu.Surface); ok {
			surf = sf
		}
		w.Draw.Release()
		if surf != nil { // note: release surface after draw
			surf.Release()
		}
		w.Glw.Destroy()
		w.Glw = nil // marks as closed for all other calls
		w.Draw = nil
	})
}

func (w *Window) SetMousePos(x, y float64) {
	if !w.IsVisible() {
		return
	}
	w.Mu.Lock()
	defer w.Mu.Unlock()
	if TheApp.Platform() == system.MacOS {
		w.Glw.SetCursorPos(x/float64(w.DevicePixelRatio), y/float64(w.DevicePixelRatio))
	} else {
		w.Glw.SetCursorPos(x, y)
	}
}

func (w *Window) SetCursorEnabled(enabled, raw bool) {
	w.CursorEnabled = enabled
	if enabled {
		w.Glw.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	} else {
		w.Glw.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		if raw && glfw.RawMouseMotionSupported() {
			w.Glw.SetInputMode(glfw.RawMouseMotion, glfw.True)
		}
	}
}

/////////////////////////////////////////////////////////
//  Window Callbacks

func (w *Window) Moved(gw *glfw.Window, x, y int) {
	w.Mu.Lock()
	w.Pos = image.Pt(x, y)
	w.Mu.Unlock()
	// w.app.GetScreens() // this can crash here on win disconnect..
	w.Screen() // gets parameters
	w.UpdateFullscreen()
	w.Event.Window(events.WinMove)
}

func (w *Window) WinResized(gw *glfw.Window, width, height int) {
	// w.app.GetScreens()  // this can crash here on win disconnect..
	w.UpdateFullscreen()
	w.UpdateGeom()
}

func (w *Window) UpdateFullscreen() {
	w.Flgs.SetFlag(w.Glw.GetAttrib(glfw.Maximized) == glfw.True, system.Fullscreen)
}

func (w *Window) UpdateGeom() {
	w.Mu.Lock()
	cursc := w.ScreenWindow
	w.Mu.Unlock()
	sc := w.Screen() // gets parameters
	w.Mu.Lock()
	var wsz image.Point
	wsz.X, wsz.Y = w.Glw.GetSize()
	// fmt.Printf("win size: %v\n", wsz)
	w.WnSize = wsz
	var fbsz image.Point
	fbsz.X, fbsz.Y = w.Glw.GetFramebufferSize()
	w.PixSize = fbsz
	w.PhysDPI = sc.PhysicalDPI
	w.LogDPI = sc.LogicalDPI
	w.Mu.Unlock()
	// w.App.RunOnMain(func() {
	// note: getting non-main thread warning.
	w.Draw.System.Renderer.SetSize(w.PixSize)
	// })
	if cursc != w.ScreenWindow {
		if MonitorDebug {
			log.Printf("desktop.Window.UpdateGeom: %v: got new screen: %v (was: %v)\n", w.Nm, w.ScreenWindow, cursc)
		}
	}
	w.Event.WindowResize()
}

func (w *Window) FbResized(gw *glfw.Window, width, height int) {
	fbsz := image.Point{width, height}
	if w.PixSize != fbsz {
		w.UpdateGeom()
	}
}

func (w *Window) OnCloseReq(gw *glfw.Window) {
	go w.CloseReq()
}

func (w *Window) Focused(gw *glfw.Window, focused bool) {
	// w.Flgs.SetFlag(focused, system.Focused)
	if focused {
		w.Event.Window(events.WinFocus)
	} else {
		w.Event.Last.MousePos = image.Point{-1, -1} // key for preventing random click to same location
		w.Event.Window(events.WinFocusLost)
	}
}

func (w *Window) Iconify(gw *glfw.Window, iconified bool) {
	w.Flgs.SetFlag(iconified, system.Minimized)
	if iconified {
		w.Event.Window(events.WinMinimize)
	} else {
		w.Screen() // gets parameters
		w.Event.Window(events.WinMinimize)
	}
}
