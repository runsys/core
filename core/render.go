// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/profile"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/cam/hct"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
)

// Rendering logic:
//
// Key principles:
//
// * Async updates (animation, mouse events, etc) change state, _set only flags_
//   using thread-safe atomic bitflag operations. True async rendering
//   is really hard to get right, and requires tons of mutexes etc. Async updates
//	 must go through [WidgetBase.AsyncLock] and [WidgetBase.AsyncUnlock].
// * Synchronous, full-tree render updates do the layout, rendering,
//   at regular FPS (frames-per-second) rate -- nop unless flag set.
//
// Three main steps:
// * Config: (re)configures widgets based on current params
//   typically by making Parts.  Always calls ApplyStyle.
// * Layout: does sizing and positioning on tree, arranging widgets.
//   Needed for whole tree after any Config changes anywhere
//   See layout.go for full details and code.
// * Render: just draws with current config, layout.
//
// ApplyStyle is always called after Config, and after any
// current state of the Widget changes via events, animations, etc
// (e.g., a Hover started or a Button is pushed down).
// Use NeedsRender() to drive the rendering update at next DoNeedsRender call.
//
// The initial configuration of a scene can skip calling
// Config and ApplyStyle because these will be called automatically
// during the Run() process for the Scene.
//
// For dynamic reconfiguration after initial display,
// Update() is the key method, calling Config then
// ApplyStyle on the node and all of its children.
//
// For nodes with dynamic content that doesn't require styling or config,
// a simple NeedsRender call will drive re-rendering.
//
// Updating is _always_ driven top-down by RenderWindow at FPS sampling rate,
// in the DoUpdate() call on the Scene.
// Three types of updates can be triggered, in order of least impact
// and highest frequency first:
// * ScNeedsRender: does NeedsRender on nodes.
// * ScNeedsLayout: does GetSize, DoLayout, then Render -- after Config.
//

// AsyncLock must be called before making any updates in a separate goroutine
// outside of the main configuration, rendering, and event handling structure.
// It must have a matching [WidgetBase.AsyncUnlock] after it.
func (wb *WidgetBase) AsyncLock() {
	rc := wb.Scene.RenderContext()
	if rc == nil {
		// if there is no render context, we are probably
		// being deleted, so we just block forever
		select {}
	}
	rc.Lock()
	if wb.This == nil {
		rc.Unlock()
		select {}
	}
	wb.Scene.updating = true
}

// AsyncUnlock must be called after making any updates in a separate goroutine
// outside of the main configuration, rendering, and event handling structure.
// It must have a matching [WidgetBase.AsyncLock] before it.
func (wb *WidgetBase) AsyncUnlock() {
	rc := wb.Scene.RenderContext()
	if rc == nil {
		return
	}
	rc.Unlock()
	if wb.Scene != nil {
		wb.Scene.updating = false
	}
}

// NeedsRender specifies that the widget needs to be rendered.
func (wb *WidgetBase) NeedsRender() {
	if DebugSettings.UpdateTrace {
		fmt.Println("\tDebugSettings.UpdateTrace: NeedsRender:", wb)
	}
	wb.needsRender = true
	if wb.Scene != nil {
		wb.Scene.sceneNeedsRender = true
	}
}

// NeedsLayout specifies that the widget's scene needs to do a layout.
// This needs to be called after any changes that affect the structure
// and/or size of elements.
func (wb *WidgetBase) NeedsLayout() {
	if DebugSettings.UpdateTrace {
		fmt.Println("\tDebugSettings.UpdateTrace: NeedsLayout:", wb)
	}
	if wb.Scene != nil {
		wb.Scene.needsLayout = true
	}
}

// NeedsRebuild returns true if the RenderContext indicates
// a full rebuild is needed.
func (wb *WidgetBase) NeedsRebuild() bool {
	if wb.This == nil || wb.Scene == nil || wb.Scene.Stage == nil {
		return false
	}
	rc := wb.Scene.RenderContext()
	if rc == nil {
		return false
	}
	return rc.Rebuild
}

// LayoutScene does a layout of the scene: Size, Position
func (sc *Scene) LayoutScene() {
	if DebugSettings.LayoutTrace {
		fmt.Println("\n############################\nLayoutScene SizeUp start:", sc)
	}
	sc.SizeUp()
	sz := &sc.Geom.Size
	sz.Alloc.Total.SetPoint(sc.SceneGeom.Size)
	sz.SetContentFromTotal(&sz.Alloc)
	// sz.Actual = sz.Alloc // todo: is this needed??
	if DebugSettings.LayoutTrace {
		fmt.Println("\n############################\nSizeDown start:", sc)
	}
	maxIter := 3
	for iter := 0; iter < maxIter; iter++ { // 3  > 2; 4 same as 3
		redo := sc.SizeDown(iter)
		if redo && iter < maxIter-1 {
			if DebugSettings.LayoutTrace {
				fmt.Println("\n############################\nSizeDown redo:", sc, "iter:", iter+1)
			}
		} else {
			break
		}
	}
	if DebugSettings.LayoutTrace {
		fmt.Println("\n############################\nSizeFinal start:", sc)
	}
	sc.SizeFinal()
	if DebugSettings.LayoutTrace {
		fmt.Println("\n############################\nPosition start:", sc)
	}
	sc.Position()
	if DebugSettings.LayoutTrace {
		fmt.Println("\n############################\nScenePos start:", sc)
	}
	sc.ScenePos()
}

// LayoutRenderScene does a layout and render of the tree:
// GetSize, DoLayout, Render.  Needed after Config.
func (sc *Scene) LayoutRenderScene() {
	sc.LayoutScene()
	sc.RenderWidget()
}

// DoNeedsRender calls Render on tree from me for nodes
// with NeedsRender flags set
func (wb *WidgetBase) DoNeedsRender() {
	if wb.This == nil {
		return
	}
	wb.WidgetWalkDown(func(w Widget, cwb *WidgetBase) bool {
		if cwb.needsRender {
			w.RenderWidget()
			return tree.Break // done
		}
		if ly := AsFrame(w); ly != nil {
			for d := math32.X; d <= math32.Y; d++ {
				if ly.HasScroll[d] && ly.Scrolls[d] != nil {
					ly.Scrolls[d].DoNeedsRender()
				}
			}
		}
		return tree.Continue
	})
}

//////////////////////////////////////////////////////////////////
//		Scene

var SceneShowIters = 2

// DoUpdate checks scene Needs flags to do whatever updating is required.
// returns false if already updating.
// This is the main update call made by the RenderWindow at FPS frequency.
func (sc *Scene) DoUpdate() bool {
	if sc.updating {
		return false
	}
	sc.updating = true // prevent rendering
	defer func() { sc.updating = false }()

	rc := sc.RenderContext()

	if sc.showIter < SceneShowIters {
		sc.needsLayout = true
		sc.showIter++
	}

	switch {
	case rc.Rebuild:
		pr := profile.Start("rebuild")
		sc.DoRebuild()
		sc.needsLayout = false
		sc.sceneNeedsRender = false
		sc.imageUpdated = true
		pr.End()
	case sc.lastRender.NeedsRestyle(rc):
		pr := profile.Start("restyle")
		sc.ApplyStyleScene()
		sc.LayoutRenderScene()
		sc.needsLayout = false
		sc.sceneNeedsRender = false
		sc.imageUpdated = true
		sc.lastRender.SaveRender(rc)
		pr.End()
	case sc.needsLayout:
		pr := profile.Start("layout")
		sc.LayoutRenderScene()
		sc.needsLayout = false
		sc.sceneNeedsRender = false
		sc.imageUpdated = true
		pr.End()
	case sc.sceneNeedsRender:
		pr := profile.Start("render")
		sc.DoNeedsRender()
		sc.sceneNeedsRender = false
		sc.imageUpdated = true
		pr.End()
	default:
		return false
	}

	if sc.showIter == SceneShowIters { // end of first pass
		sc.showIter++
		if !sc.prefSizing {
			sc.Events.ActivateStartFocus()
		}
	}

	return true
}

// MakeSceneWidgets calls Config on all widgets in the Scene,
// which will set NeedsLayout to drive subsequent layout and render.
// This is a top-level call, typically only done when the window
// is first drawn, once the full sizing information is available.
func (sc *Scene) MakeSceneWidgets() {
	sc.updating = true // prevent rendering
	defer func() { sc.updating = false }()

	sc.UpdateTree()
}

// ApplyStyleScene calls ApplyStyle on all widgets in the Scene,
// This is needed whenever the window geometry, DPI,
// etc is updated, which affects styling.
func (sc *Scene) ApplyStyleScene() {
	sc.updating = true // prevent rendering
	defer func() { sc.updating = false }()

	sc.StyleTree()
	sc.needsLayout = true
}

// DoRebuild does the full re-render and RenderContext Rebuild flag
// should be used by Widgets to rebuild things that are otherwise
// cached (e.g., Icon, TextCursor).
func (sc *Scene) DoRebuild() {
	sc.MakeSceneWidgets()
	sc.ApplyStyleScene()
	sc.LayoutRenderScene()
}

// PrefSize computes the preferred size of the scene based on current contents.
// initSz is the initial size -- e.g., size of screen.
// Used for auto-sizing windows.
func (sc *Scene) PrefSize(initSz image.Point) image.Point {
	sc.updating = true // prevent rendering
	defer func() { sc.updating = false }()

	sc.prefSizing = true
	sc.MakeSceneWidgets()
	sc.ApplyStyleScene()
	sc.LayoutScene()
	sz := &sc.Geom.Size
	psz := sz.Actual.Total
	sc.prefSizing = false
	sc.showIter = 0
	return psz.ToPointFloor()
}

//////////////////////////////////////////////////////////////////
//		Widget local rendering

// PushBounds pushes our bounding box bounds onto the bounds stack
// if they are non-empty. This automatically limits our drawing to
// our own bounding box. This must be called as the first step in
// Render implementations. It returns whether the new bounds are
// empty or not; if they are empty, then don't render.
func (wb *WidgetBase) PushBounds() bool {
	if wb == nil || wb.This == nil {
		return false
	}
	wb.needsRender = false             // done!
	if !wb.This.(Widget).IsVisible() { // checks deleted etc
		return false
	}
	if wb.Geom.TotalBBox.Empty() {
		if DebugSettings.RenderTrace {
			fmt.Printf("Render empty bbox: %v at %v\n", wb.Path(), wb.Geom.TotalBBox)
		}
		return false
	}
	wb.Styles.ComputeActualBackground(wb.ParentActualBackground())
	pc := &wb.Scene.PaintContext
	if pc.State == nil || pc.Image == nil {
		return false
	}
	radius := wb.Styles.Border.Radius.Dots()
	if wb.renderContainer {
		radius.Set(-1)
	}
	pc.PushBoundsGeom(wb.Geom.TotalBBox, wb.Geom.ContentBBox, radius)
	pc.Defaults() // start with default values
	if DebugSettings.RenderTrace {
		fmt.Printf("Render: %v at %v\n", wb.Path(), wb.Geom.TotalBBox)
	}
	return true
}

// PopBounds pops our bounding box bounds. This is the last step
// in Render implementations after rendering children.
func (wb *WidgetBase) PopBounds() {
	if wb == nil || wb.This == nil {
		return
	}
	pc := &wb.Scene.PaintContext

	isSelw := wb.Scene.selectedWidget == wb.This
	if wb.Scene.renderBBoxes || isSelw {
		pos := math32.Vector2FromPoint(wb.Geom.TotalBBox.Min)
		sz := math32.Vector2FromPoint(wb.Geom.TotalBBox.Size())
		// node: we won't necc. get a push prior to next update, so saving these.
		pcsw := pc.StrokeStyle.Width
		pcsc := pc.StrokeStyle.Color
		pcfc := pc.FillStyle.Color
		pcop := pc.FillStyle.Opacity
		pc.StrokeStyle.Width.Dot(1)
		pc.StrokeStyle.Color = colors.C(hct.New(wb.Scene.renderBBoxHue, 100, 50))
		pc.FillStyle.Color = nil
		if isSelw {
			fc := pc.StrokeStyle.Color
			pc.FillStyle.Color = fc
			pc.FillStyle.Opacity = 0.2
		}
		pc.DrawRectangle(pos.X, pos.Y, sz.X, sz.Y)
		pc.FillStrokeClear()
		// restore
		pc.FillStyle.Opacity = pcop
		pc.FillStyle.Color = pcfc
		pc.StrokeStyle.Width = pcsw
		pc.StrokeStyle.Color = pcsc

		wb.Scene.renderBBoxHue += 10
		if wb.Scene.renderBBoxHue > 360 {
			rmdr := (int(wb.Scene.renderBBoxHue-360) + 1) % 9
			wb.Scene.renderBBoxHue = float32(rmdr)
		}
	}

	pc.PopBounds()
}

// Render is the method that widgets should implement to define their
// custom rendering steps. It should not typically be called outside of
// [Widget.RenderWidget], which also does other steps applicable
// for all widgets. The base [WidgetBase.Render] implementation
// renders the standard box model.
func (wb *WidgetBase) Render() {
	wb.RenderStandardBox()
}

// RenderWidget renders the widget and any parts and children that it has.
// It does not render if the widget is invisible. It calls Widget.Render]
// for widget-specific rendering.
func (wb *WidgetBase) RenderWidget() {
	if wb.PushBounds() {
		wb.This.(Widget).Render()
		wb.RenderParts()
		wb.RenderChildren()
		wb.PopBounds()
	}
}

func (wb *WidgetBase) RenderParts() {
	if wb.Parts != nil {
		wb.Parts.RenderWidget()
	}
}

// RenderChildren renders all of the widget's children.
func (wb *WidgetBase) RenderChildren() {
	wb.WidgetKidsIter(func(i int, kwi Widget, kwb *WidgetBase) bool {
		kwi.RenderWidget()
		return tree.Continue
	})
}

////////////////////////////////////////////////////////////////////////////////
//  Standard Box Model rendering

// RenderBoxImpl implements the standard box model rendering, assuming all
// paint parameters have already been set.
func (wb *WidgetBase) RenderBoxImpl(pos math32.Vector2, sz math32.Vector2, bs styles.Border) {
	wb.Scene.PaintContext.DrawBorder(pos.X, pos.Y, sz.X, sz.Y, bs)
}

// RenderStandardBox renders the standard box model.
func (wb *WidgetBase) RenderStandardBox() {
	pos := wb.Geom.Pos.Total
	sz := wb.Geom.Size.Actual.Total
	wb.Scene.PaintContext.DrawStandardBox(&wb.Styles, pos, sz, wb.ParentActualBackground())
}

//////////////////////////////////////////////////////////////////
//		Widget position functions

// PointToRelPos translates a point in Scene pixel coords
// into relative position within node, based on the Content BBox
func (wb *WidgetBase) PointToRelPos(pt image.Point) image.Point {
	return pt.Sub(wb.Geom.ContentBBox.Min)
}

// WinBBox returns the RenderWindow based bounding box for the widget
// by adding the Scene position to the ScBBox
func (wb *WidgetBase) WinBBox() image.Rectangle {
	bb := wb.Geom.TotalBBox
	if wb.Scene != nil {
		return bb.Add(wb.Scene.SceneGeom.Pos)
	}
	return bb
}

// WinPos returns the RenderWindow based position within the
// bounding box of the widget, where the x, y coordinates
// are the proportion across the bounding box to use:
// 0 = left / top, 1 = right / bottom
func (wb *WidgetBase) WinPos(x, y float32) image.Point {
	bb := wb.WinBBox()
	sz := bb.Size()
	var pt image.Point
	pt.X = bb.Min.X + int(math32.Round(float32(sz.X)*x))
	pt.Y = bb.Min.Y + int(math32.Round(float32(sz.Y)*y))
	return pt
}

/////////////////////////////////////////////////////////////////////////////
//	Profiling and Benchmarking, controlled by app bar buttons and keyboard shortcuts

// ProfileToggle turns profiling on or off, which does both
// targeted and global CPU and Memory profiling.
func ProfileToggle() { //types:add
	if profile.Profiling {
		EndTargetedProfile()
		EndCPUMemoryProfile()
	} else {
		StartTargetedProfile()
		StartCPUMemoryProfile()
	}
}

// cpuProfileFile is the file created by [StartCPUMemoryProfile],
// which needs to be stored so that it can be closed in [EndCPUMemoryProfile].
var cpuProfileFile *os.File

// StartCPUMemoryProfile starts the standard Go cpu and memory profiling.
func StartCPUMemoryProfile() {
	fmt.Println("Starting standard cpu and memory profiling")
	f, err := os.Create("cpu.prof")
	if errors.Log(err) == nil {
		cpuProfileFile = f
		errors.Log(pprof.StartCPUProfile(f))
	}
}

// EndCPUMemoryProfile ends the standard Go cpu and memory profiling.
func EndCPUMemoryProfile() {
	fmt.Println("Ending standard cpu and memory profiling")
	pprof.StopCPUProfile()
	errors.Log(cpuProfileFile.Close())
	f, err := os.Create("mem.prof")
	if errors.Log(err) == nil {
		runtime.GC() // get up-to-date statistics
		errors.Log(pprof.WriteHeapProfile(f))
		errors.Log(f.Close())
	}
}

// StartTargetedProfile starts targeted profiling using the prof package.
func StartTargetedProfile() {
	fmt.Println("Starting targeted profiling")
	profile.Reset()
	profile.Profiling = true
}

// EndTargetedProfile ends targeted profiling and prints report.
func EndTargetedProfile() {
	profile.Report(time.Millisecond)
	profile.Profiling = false
}

// BenchmarkFullRender runs benchmark of 50 full re-renders (full restyling, layout,
// and everything), reporting targeted profile results and generating standard
// Go cpu.prof and mem.prof outputs.
func (sc *Scene) BenchmarkFullRender() {
	fmt.Println("Starting BenchmarkFullRender")
	StartCPUMemoryProfile()
	StartTargetedProfile()
	ts := time.Now()
	n := 50
	for i := 0; i < n; i++ {
		sc.LayoutScene()
		sc.RenderWidget()
	}
	td := time.Since(ts)
	fmt.Printf("Time for %v full renders: %12.2f s\n", n, float64(td)/float64(time.Second))
	EndTargetedProfile()
	EndCPUMemoryProfile()
}

// BenchmarkReRender runs benchmark of 50 re-render-only updates of display
// (just the raw rendering, no styling or layout), reporting targeted profile
// results and generating standard Go cpu.prof and mem.prof outputs.
func (sc *Scene) BenchmarkReRender() {
	fmt.Println("Starting BenchmarkReRender")
	StartTargetedProfile()
	start := time.Now()
	n := 50
	for range n {
		sc.RenderWidget()
	}
	td := time.Since(start)
	fmt.Printf("Time for %v renders: %12.2f s\n", n, float64(td)/float64(time.Second))
	EndTargetedProfile()
}
