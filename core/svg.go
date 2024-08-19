// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/fs"
	"strings"

	"cogentcore.org/core/cursors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/svg"
	"cogentcore.org/core/tree"
	"golang.org/x/image/draw"
)

// SVG is a Widget that renders an [svg.SVG] object.
// If it is not [states.ReadOnly], the user can pan and zoom the display.
// By default, it is [states.ReadOnly].
type SVG struct {
	WidgetBase

	// SVG is the SVG drawing to display in this widget
	SVG *svg.SVG `set:"-"`

	// prevSize is the cached allocated size for the last rendered image.
	prevSize image.Point `xml:"-" json:"-" set:"-"`
}

func (sv *SVG) Init() {
	sv.WidgetBase.Init()
	sv.SVG = svg.NewSVG(10, 10)
	sv.SetReadOnly(true)
	sv.Styler(func(s *styles.Style) {
		s.Min.Set(units.Dp(256))

		ro := sv.IsReadOnly()
		s.SetAbilities(!ro, abilities.Slideable, abilities.Activatable, abilities.Scrollable)
		if !ro {
			if s.Is(states.Active) {
				s.Cursor = cursors.Grabbing
				s.StateLayer = 0
			} else {
				s.Cursor = cursors.Grab
			}
		}
	})
	sv.FinalStyler(func(s *styles.Style) {
		sv.SVG.Root.ViewBox.PreserveAspectRatio.SetFromStyle(s)
	})

	sv.On(events.SlideMove, func(e events.Event) {
		if sv.IsReadOnly() {
			return
		}
		e.SetHandled()
		del := e.PrevDelta()
		sv.SVG.Translate.X += float32(del.X)
		sv.SVG.Translate.Y += float32(del.Y)
		sv.NeedsRender()
	})
	sv.On(events.Scroll, func(e events.Event) {
		if sv.IsReadOnly() {
			return
		}
		e.SetHandled()
		se := e.(*events.MouseScroll)
		sv.SVG.Scale += float32(se.Delta.Y) / 100
		if sv.SVG.Scale <= 0.0000001 {
			sv.SVG.Scale = 0.01
		}
		sv.NeedsRender()
	})
}

// Open opens an XML-formatted SVG file
func (sv *SVG) Open(filename Filename) error { //types:add
	return sv.SVG.OpenXML(string(filename))
}

// OpenSVG opens an XML-formatted SVG file from the given fs.
func (sv *SVG) OpenFS(fsys fs.FS, filename string) error {
	return sv.SVG.OpenFS(fsys, filename)
}

// Read reads an XML-formatted SVG file from the given reader.
func (sv *SVG) Read(r io.Reader) error {
	return sv.SVG.ReadXML(r)
}

// ReadBytes reads an XML-formatted SVG file from the given bytes.
func (sv *SVG) ReadBytes(b []byte) error {
	return sv.SVG.ReadXML(bytes.NewReader(b))
}

// ReadString reads an XML-formatted SVG file from the given string.
func (sv *SVG) ReadString(s string) error {
	return sv.SVG.ReadXML(strings.NewReader(s))
}

// SaveSVG saves the current SVG to an XML-encoded standard SVG file.
func (sv *SVG) SaveSVG(filename Filename) error { //types:add
	return sv.SVG.SaveXML(string(filename))
}

// SavePNG saves the current rendered SVG image to an PNG image file.
func (sv *SVG) SavePNG(filename Filename) error { //types:add
	return sv.SVG.SavePNG(string(filename))
}

func (sv *SVG) SizeFinal() {
	sv.WidgetBase.SizeFinal()
	sv.SVG.Resize(sv.Geom.Size.Actual.Content.ToPoint())
}

// RenderSVG renders the SVG
func (sv *SVG) RenderSVG() {
	if sv.SVG == nil {
		return
	}
	// need to make the image again to prevent it from
	// rendering over itself
	sv.SVG.Pixels = image.NewRGBA(sv.SVG.Pixels.Rect)
	sv.SVG.RenderState.Init(sv.SVG.Pixels.Rect.Dx(), sv.SVG.Pixels.Rect.Dy(), sv.SVG.Pixels)
	sv.SVG.Render()
	sv.prevSize = sv.SVG.Pixels.Rect.Size()
}

func (sv *SVG) Render() {
	sv.WidgetBase.Render()
	if sv.SVG == nil {
		return
	}
	needsRender := !sv.IsReadOnly()
	if !needsRender {
		if sv.SVG.Pixels == nil {
			needsRender = true
		} else {
			sz := sv.SVG.Pixels.Bounds().Size()
			if sz != sv.prevSize || sz == (image.Point{}) {
				needsRender = true
			}
		}
	}
	if needsRender {
		sv.RenderSVG()
	}
	r := sv.Geom.ContentBBox
	sp := sv.Geom.ScrollOffset()
	draw.Draw(sv.Scene.Pixels, r, sv.SVG.Pixels, sp, draw.Over)
}

func (sv *SVG) MakeToolbar(p *tree.Plan) {
	// TODO(kai): resolve svg panning and selection structure
	tree.Add(p, func(w *Button) {
		w.SetIcon(icons.PanTool)
		w.SetTooltip("toggle the ability to zoom and pan the view")
		w.OnClick(func(e events.Event) {
			sv.SetReadOnly(!sv.IsReadOnly())
			sv.Restyle()
		})
	})
	tree.Add(p, func(w *Button) {
		w.SetIcon(icons.ArrowForward)
		w.SetTooltip("turn on select mode for selecting SVG elements")
		w.OnClick(func(e events.Event) {
			fmt.Println("this will select select mode")
		})
	})
	tree.Add(p, func(w *Separator) {})
	tree.Add(p, func(w *FuncButton) {
		w.SetFunc(sv.Open)
	})
	tree.Add(p, func(w *FuncButton) {
		w.SetFunc(sv.SaveSVG).SetIcon(icons.Save)
	})
	tree.Add(p, func(w *FuncButton) {
		w.SetFunc(sv.SavePNG).SetIcon(icons.Save)
	})
}
