// Copyright 2023 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on golang.org/x/exp/shiny:
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"image"

	"cogentcore.org/core/events"
	"cogentcore.org/core/system"
)

// WindowMulti contains the data and logic common to all implementations of [system.Window]
// on multi-window platforms (desktop), as opposed to single-window
// platforms (mobile, web, and offscreen), for which you should use [WindowSingle].
// A WindowMulti is associated with a corresponding [system.App] type.
// The [system.App] type should embed [AppMulti].
type WindowMulti[A system.App, D system.Drawer] struct {
	Window[A]

	// Event is the event manager for the window
	Event events.Source `label:"Event manger"`

	// Draw is the [system.Drawer] used for this window.
	Draw D `label:"Drawer"`

	// Pos is the position of the window
	Pos image.Point `label:"Position"`

	// WnSize is the size of the window in window manager coordinates
	WnSize image.Point `label:"Window manager size"`

	// PixSize is the pixel size of the window in raw display dots
	PixSize image.Point `label:"Pixel size"`

	// DevicePixelRatio is a factor that scales the screen's
	// "natural" pixel coordinates into actual device pixels.
	// On OS-X, it is backingScaleFactor = 2.0 on "retina"
	DevicePixelRatio float32

	// PhysicalDPI is the physical dots per inch of the screen,
	// for generating true-to-physical-size output.
	// It is computed as 25.4 * (PixSize.X / PhysicalSize.X)
	// where 25.4 is the number of mm per inch.
	PhysDPI float32 `label:"Physical DPI"`

	// LogicalDPI is the logical dots per inch of the screen,
	// which is used for all rendering.
	// It is: transient zoom factor * screen-specific multiplier * PhysicalDPI
	LogDPI float32 `label:"Logical DPI"`
}

// NewWindowMulti makes a new [WindowMulti] for the given app with the given options.
func NewWindowMulti[A system.App, D system.Drawer](a A, opts *system.NewWindowOptions) WindowMulti[A, D] {
	return WindowMulti[A, D]{
		Window: NewWindow(a, opts),
	}
}

func (w *WindowMulti[A, D]) Events() *events.Source {
	return &w.Event
}

func (w *WindowMulti[A, D]) Drawer() system.Drawer {
	return w.Draw
}

func (w *WindowMulti[A, D]) Size() image.Point {
	// w.Mu.Lock() // this prevents race conditions but also locks up
	// defer w.Mu.Unlock()
	return w.PixSize
}

func (w *WindowMulti[A, D]) WinSize() image.Point {
	// w.Mu.Lock() // this prevents race conditions but also locks up
	// defer w.Mu.Unlock()
	return w.WnSize
}

func (w *WindowMulti[A, D]) Position() image.Point {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	return w.Pos
}

func (w *WindowMulti[A, D]) PhysicalDPI() float32 {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	return w.PhysDPI
}

func (w *WindowMulti[A, D]) LogicalDPI() float32 {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	return w.LogDPI
}

func (w *WindowMulti[A, D]) SetLogicalDPI(dpi float32) {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	w.LogDPI = dpi
}

func (w *WindowMulti[A, D]) SetWinSize(sz image.Point) {
	if w.This.IsClosed() {
		return
	}
	w.WnSize = sz
}

func (w *WindowMulti[A, D]) SetSize(sz image.Point) {
	if w.This.IsClosed() {
		return
	}
	sc := w.This.Screen()
	sz = sc.WinSizeFromPix(sz)
	w.SetWinSize(sz)
}

func (w *WindowMulti[A, D]) SetPos(pos image.Point) {
	if w.This.IsClosed() {
		return
	}
	w.Pos = pos
}

func (w *WindowMulti[A, D]) SetGeom(pos image.Point, sz image.Point) {
	if w.This.IsClosed() {
		return
	}
	sc := w.This.Screen()
	sz = sc.WinSizeFromPix(sz)
	w.SetWinSize(sz)
	w.Pos = pos
}

func (w *WindowMulti[A, D]) IsVisible() bool {
	return w.Window.IsVisible() && w.App.NScreens() != 0
}
