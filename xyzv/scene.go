// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xyzv

//go:generate core generate

import (
	"image"
	"image/draw"
	"log"
	"log/slog"

	"cogentcore.org/core/abilities"
	"cogentcore.org/core/events"
	"cogentcore.org/core/gi"
	"cogentcore.org/core/goosi"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/units"
	"cogentcore.org/core/vgpu"
	"cogentcore.org/core/xyz"
)

// Scene is a gi.Widget that manages a xyz.Scene,
// providing the basic rendering logic for the 3D scene
// in the 2D gi gui context.
type Scene struct {
	gi.WidgetBase

	// XYZ is the 3D xyz.Scene
	XYZ *xyz.Scene `set:"-"`

	// how to deal with selection / manipulation events
	SelMode SelModes

	// currently selected node
	CurSel xyz.Node `copier:"-" json:"-" xml:"-" view:"-"`

	// currently selected manipulation control point
	CurManipPt *ManipPt `copier:"-" json:"-" xml:"-" view:"-"`

	// parameters for selection / manipulation box
	SelParams SelParams `view:"inline"`
}

func (sw *Scene) OnInit() {
	sw.XYZ = xyz.NewScene("Scene")
	sw.XYZ.Defaults()
	sw.SelParams.Defaults()
	sw.WidgetBase.OnInit()
	sw.HandleEvents()
	sw.SetStyles()
}

// SceneXYZ returns the xyz.Scene
func (sw *Scene) SceneXYZ() *xyz.Scene {
	return sw.XYZ
}

func (sw *Scene) SetStyles() {
	sw.Style(func(s *styles.Style) {
		s.SetAbilities(true, abilities.Clickable, abilities.Focusable, abilities.Activatable, abilities.Slideable, abilities.LongHoverable, abilities.DoubleClickable)
		s.Grow.Set(1, 1)
		s.Min.Set(units.Em(20))
	})
}

func (sw *Scene) HandleEvents() {
	sw.On(events.Scroll, func(e events.Event) {
		// pos := sw.Geom.ContentBBox.Min
		// fmt.Println("loc off:", e.LocalOff(), "pos:", pos, "e pos:", e.WindowPos())
		// e.SetLocalOff(e.LocalOff().Add(pos))
		sw.XYZ.MouseScrollEvent(e.(*events.MouseScroll))
		sw.NeedsRender(true)
	})
	sw.On(events.KeyChord, func(e events.Event) {
		sw.XYZ.KeyChordEvent(e)
		sw.NeedsRender(true)
	})
	sw.HandleSlideEvents()
	sw.HandleSelectEvents()
}

func (sw *Scene) ConfigWidget() {
	sw.ConfigFrame()
}

// ConfigFrame configures the framebuffer for GPU rendering,
// using the RenderWin GPU and Device.
func (sw *Scene) ConfigFrame() {
	zp := image.Point{}
	sz := sw.Geom.Size.Actual.Content.ToPointFloor()
	if sz == zp {
		return
	}
	sw.XYZ.Geom.Size = sz

	doConfig := false
	if sw.XYZ.Frame != nil {
		cursz := sw.XYZ.Frame.Format.Size
		if cursz == sz {
			return
		}
	} else {
		doConfig = true
	}

	win := sw.WidgetBase.Scene.EventMgr.RenderWin()
	if win == nil {
		return
	}
	drw := win.GoosiWin.Drawer()
	goosi.TheApp.RunOnMain(func() {
		sw.XYZ.ConfigFrameFromSurface(drw.Surface().(*vgpu.Surface))
		if doConfig {
			sw.XYZ.Config()
		}
	})
	sw.XYZ.SetFlag(true, xyz.ScNeedsRender)
	sw.NeedsRender(true)
}

func (sw *Scene) DrawIntoScene() {
	if sw.XYZ.Frame == nil {
		return
	}
	r := sw.Geom.ContentBBox
	sp := image.Point{}
	if sw.Par != nil { // use parents children bbox to determine where we can draw
		_, pwb := gi.AsWidget(sw.Par)
		pbb := pwb.Geom.ContentBBox
		nr := r.Intersect(pbb)
		sp = nr.Min.Sub(r.Min)
		if sp.X < 0 || sp.Y < 0 || sp.X > 10000 || sp.Y > 10000 {
			slog.Error("xyzv.Scene bad bounding box", "path", sw, "startPos", sp, "bbox", r, "parBBox", pbb)
			return
		}
		r = nr
	}
	img, err := sw.XYZ.Image() // direct access
	if err != nil {
		log.Println("frame image err:", err)
		return
	}
	draw.Draw(sw.WidgetBase.Scene.Pixels, r, img, sp, draw.Src) // note: critical to not use Over here!
	sw.XYZ.ImageDone()
}

// Render renders the Frame Image
func (sw *Scene) Render3D() {
	sw.ConfigFrame() // nop if all good
	if sw.XYZ.Frame == nil {
		return
	}
	if sw.XYZ.Is(xyz.ScNeedsConfig) {
		goosi.TheApp.RunOnMain(func() {
			sw.XYZ.Config()
		})
	}
	sw.XYZ.DoUpdate()
}

func (sw *Scene) Render() {
	if sw.PushBounds() {
		sw.Render3D()
		sw.DrawIntoScene()
		sw.RenderChildren()
		sw.PopBounds()
	}
}

// UpdateStart3D calls UpdateStart on the 3D Scene:
// sets the scene ScUpdating flag to prevent
// render updates during construction on a scene.
// if already updating, returns false.
// Pass the result to UpdateEnd* methods.
func (sw *Scene) UpdateStart3D() bool {
	return sw.XYZ.UpdateStart()
}

// UpdateEnd3D calls UpdateEnd on the 3D Scene:
// resets the scene ScUpdating flag if updt = true
func (sw *Scene) UpdateEnd3D(updt bool) {
	sw.XYZ.UpdateEnd(updt)
}

// UpdateEndRender3D calls UpdateEndRender on the 3D Scene
// and calls gi NeedsRender.
// resets the scene ScUpdating flag if updt = true
// and sets the ScNeedsRender flag; updt is from UpdateStart().
// Render only updates based on camera changes, not any node-level
// changes. See [UpdateEndUpdate].
func (sw *Scene) UpdateEndRender3D(updt bool) {
	if updt {
		sw.XYZ.UpdateEndRender(updt)
		sw.NeedsRender(updt)
	}
}

// UpdateEndUpdate3D calls UpdateEndUpdate on the 3D Scene
// and calls gi NeedsRender.
// UpdateEndUpdate resets the scene ScUpdating flag if updt = true
// and sets the ScNeedsUpdate flag; updt is from UpdateStart().
// Update is for when any node Pose or material changes happen.
// See [UpdateEndConfig] for major changes.
func (sw *Scene) UpdateEndUpdate3D(updt bool) {
	if updt {
		sw.XYZ.UpdateEndUpdate(updt)
		sw.NeedsRender(updt)
	}
}

// UpdateEndConfig3D calls UpdateEndConfig on the 3D Scene
// and calls gi NeedsRender.
// UpdateEndConfig resets the scene ScUpdating flag if updt = true
// and sets the ScNeedsConfig flag; updt is from UpdateStart().
// Config is for Texture, Lighting Meshes or more complex nodes).
func (sw *Scene) UpdateEndConfig3D(updt bool) {
	if updt {
		sw.XYZ.UpdateEndConfig(updt)
		sw.NeedsRender(updt)
	}
}

// Direct render to Drawer frame
// drw := sc.Win.OSWin.Drawer()
// drw.SetFrameImage(sc.DirUpIdx, sc.Frame.Frames[0])
// sc.Win.DirDraws.SetWinBBox(sc.DirUpIdx, sc.WinBBox)
// drw.SyncImages()
