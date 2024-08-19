// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"image"
	"log/slog"
	"reflect"

	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/cursors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
)

// Slider is a slideable widget that provides slider functionality for two Types:
// Slider type provides a movable thumb that represents Value as the center of thumb
// Pos position, with room reserved at ends for 1/2 of the thumb size.
// Scrollbar has a VisiblePct factor that specifies the percent of the content
// currently visible, which determines the size of the thumb, and thus the range of motion
// remaining for the thumb Value (VisiblePct = 1 means thumb is full size, and no remaining
// range of motion).
// The Content size (inside the margin and padding) determines the outer bounds of
// the rendered area.
// The [styles.Style.Direction] determines the direction in which the slider slides.
type Slider struct {
	Frame

	// Type is the type of the slider, which determines its visual
	// and functional properties. The default type, [SliderSlider],
	// should work for most end-user use cases.
	Type SliderTypes

	// Value is the current value, represented by the position of the thumb.
	// It defaults to 0.5.
	Value float32 `set:"-"`

	// Min is the minimum possible value.
	// It defaults to 0.
	Min float32

	// Max is the maximum value supported.
	// It defaults to 1.
	Max float32

	// Step is the amount that the arrow keys increment/decrement the value by.
	// It defaults to 0.1.
	Step float32

	// EnforceStep is whether to ensure that the value is always
	// a multiple of [Slider.Step].
	EnforceStep bool

	// PageStep is the amount that the PageUp and PageDown keys
	// increment/decrement the value by.
	// It defaults to 0.2, and will be at least as big as [Slider.Step].
	PageStep float32

	// Icon is an optional icon to use for the dragging knob.
	Icon icons.Icon `display:"show-name"`

	// For Scrollbar type only: proportion (1 max) of the full range of scrolled data
	// that is currently visible.  This determines the thumb size and range of motion:
	// if 1, full slider is the thumb and no motion is possible.
	VisiblePct float32 `set:"-"`

	// Size of the thumb as a proportion of the slider thickness, which is
	// Content size (inside the padding).  This is for actual X,Y dimensions,
	// so must be sensitive to Dim dimension alignment.
	ThumbSize math32.Vector2

	// TrackSize is the proportion of slider thickness for the visible track
	// for the Slider type.  It is often thinner than the thumb, achieved by
	// values < 1 (.5 default)
	TrackSize float32

	// threshold for amount of change in scroll value before emitting an input event
	InputThreshold float32

	// specifies the precision of decimal places (total, not after the decimal point)
	// to use in representing the number. This helps to truncate small weird floating
	// point values in the nether regions.
	Prec int

	// The background color that is used for styling the selected value section of the slider.
	// It should be set in the StyleFuncs, just like the main style object is.
	// If it is set to transparent, no value is rendered, so the value section of the slider
	// just looks like the rest of the slider.
	ValueColor image.Image

	// The background color that is used for styling the thumb (handle) of the slider.
	// It should be set in the StyleFuncs, just like the main style object is.
	// If it is set to transparent, no thumb is rendered, so the thumb section of the slider
	// just looks like the rest of the slider.
	ThumbColor image.Image

	// If true, keep the slider (typically a Scrollbar) within the parent Scene
	// bounding box, if the parent is in view.  This is the default behavior
	// for Layout scrollbars, and setting this flag replicates that behavior
	// in other scrollbars.
	StayInView bool

	//////////////////////////////////////////////////////////////////
	// 	Computed values below

	// logical position of the slider relative to Size
	Pos float32 `edit:"-" set:"-"`

	// previous Change event emitted value - don't re-emit Change if it is the same
	LastValue float32 `edit:"-" copier:"-" xml:"-" json:"-" set:"-"`

	// previous sliding value - for computing the Input change
	PrevSlide float32 `edit:"-" copier:"-" xml:"-" json:"-" set:"-"`

	// Computed size of the slide box in the relevant dimension
	// range of motion, exclusive of spacing, based on layout allocation.
	Size float32 `edit:"-" set:"-"`

	// underlying drag position of slider -- not subject to snapping
	SlideStartPos float32 `edit:"-" set:"-"`
}

// SliderTypes are the different types of sliders
type SliderTypes int32 //enums:enum -trim-prefix Slider

const (
	// SliderSlider indicates a standard, user-controllable slider
	// for setting a numeric value
	SliderSlider SliderTypes = iota

	// SliderScrollbar indicates a slider acting as a scrollbar for content
	// This sets the
	SliderScrollbar
)

func (sr *Slider) WidgetValue() any { return &sr.Value }

func (sr *Slider) OnBind(value any) {
	kind := reflectx.NonPointerType(reflect.TypeOf(value)).Kind()
	if kind >= reflect.Int && kind <= reflect.Uintptr {
		sr.SetStep(1).SetEnforceStep(true).SetMax(100)
	}
}

func (sr *Slider) Init() {
	sr.Frame.Init()
	sr.Value = 0.5
	sr.Max = 1
	sr.VisiblePct = 1
	sr.Step = 0.1
	sr.PageStep = 0.2
	sr.Prec = 9
	sr.ThumbSize.Set(1, 1)
	sr.TrackSize = 0.5
	sr.Styler(func(s *styles.Style) {
		s.SetAbilities(true, abilities.Activatable, abilities.Focusable, abilities.Hoverable, abilities.Slideable)

		// we use a different color for the thumb and value color
		// (compared to the background color) so that they get the
		// correct state layer
		s.Color = colors.Scheme.Primary.On

		if sr.Type == SliderSlider {
			sr.ValueColor = colors.Scheme.Primary.Base
			sr.ThumbColor = colors.Scheme.Primary.Base
			s.Padding.Set(units.Dp(8))
			s.Background = colors.Scheme.SurfaceVariant
		} else {
			sr.ValueColor = colors.Scheme.OutlineVariant
			sr.ThumbColor = colors.Scheme.OutlineVariant
			s.Background = colors.Scheme.SurfaceContainerLow
		}

		// sr.ValueColor = s.StateBackgroundColor(sr.ValueColor)
		// sr.ThumbColor = s.StateBackgroundColor(sr.ThumbColor)
		s.Color = colors.Scheme.OnSurface

		s.Border.Radius = styles.BorderRadiusFull
		if !sr.IsReadOnly() {
			s.Cursor = cursors.Grab
			switch {
			case s.Is(states.Sliding):
				s.Cursor = cursors.Grabbing
			case s.Is(states.Active):
				s.Cursor = cursors.Grabbing
			}
		}
	})
	sr.FinalStyler(func(s *styles.Style) {
		if s.Direction == styles.Row {
			s.Min.X.Em(20)
			s.Min.Y.Em(1)
		} else {
			s.Min.Y.Em(20)
			s.Min.X.Em(1)
		}
		if sr.Type == SliderScrollbar {
			if s.Direction == styles.Row {
				s.Min.Y = s.ScrollBarWidth
			} else {
				s.Min.X = s.ScrollBarWidth
			}
		}
	})

	sr.On(events.MouseDown, func(e events.Event) {
		pos := sr.PointToRelPos(e.Pos())
		sr.SetSliderPosAction(pos)
		sr.SlideStartPos = sr.Pos
	})
	// note: not doing anything in particular on SlideStart
	sr.On(events.SlideMove, func(e events.Event) {
		del := e.StartDelta()
		if sr.Styles.Direction == styles.Row {
			sr.SetSliderPosAction(sr.SlideStartPos + float32(del.X))
		} else {
			sr.SetSliderPosAction(sr.SlideStartPos + float32(del.Y))
		}
	})
	// we need to send change events for both SlideStop and Click
	// to handle the no-slide click case
	change := func(e events.Event) {
		pos := sr.PointToRelPos(e.Pos())
		sr.SetSliderPosAction(pos)
		sr.SendChanged()
	}
	sr.On(events.SlideStop, change)
	sr.On(events.Click, change)
	sr.On(events.Scroll, func(e events.Event) {
		se := e.(*events.MouseScroll)
		se.SetHandled()
		var del float32
		// if we are scrolling in the y direction on an x slider,
		// we still count it
		if sr.Styles.Direction == styles.Row && se.Delta.X != 0 {
			del = se.Delta.X
		} else {
			del = se.Delta.Y
		}
		if sr.Type == SliderScrollbar {
			del = -del // invert for "natural" scroll
		}
		edel := sr.ScrollScale(del)
		sr.SetValueAction(sr.Value + edel)
		sr.SendChanged()
	})
	sr.OnKeyChord(func(e events.Event) {
		kf := keymap.Of(e.KeyChord())
		if DebugSettings.KeyEventTrace {
			slog.Info("SliderBase KeyInput", "widget", sr, "keyFunction", kf)
		}
		switch kf {
		case keymap.MoveUp:
			sr.SetValueAction(sr.Value - sr.Step)
			e.SetHandled()
		case keymap.MoveLeft:
			sr.SetValueAction(sr.Value - sr.Step)
			e.SetHandled()
		case keymap.MoveDown:
			sr.SetValueAction(sr.Value + sr.Step)
			e.SetHandled()
		case keymap.MoveRight:
			sr.SetValueAction(sr.Value + sr.Step)
			e.SetHandled()
		case keymap.PageUp:
			if sr.PageStep < sr.Step {
				sr.PageStep = 2 * sr.Step
			}
			sr.SetValueAction(sr.Value - sr.PageStep)
			e.SetHandled()
		case keymap.PageDown:
			if sr.PageStep < sr.Step {
				sr.PageStep = 2 * sr.Step
			}
			sr.SetValueAction(sr.Value + sr.PageStep)
			e.SetHandled()
		case keymap.Home:
			sr.SetValueAction(sr.Min)
			e.SetHandled()
		case keymap.End:
			sr.SetValueAction(sr.Max)
			e.SetHandled()
		}
	})

	sr.Maker(func(p *tree.Plan) {
		if sr.Icon.IsNil() {
			return
		}
		tree.AddAt(p, "icon", func(w *Icon) {
			w.Styler(func(s *styles.Style) {
				s.Font.Size.Dp(24)
				s.Color = sr.ThumbColor
			})
			w.Updater(func() {
				w.SetIcon(sr.Icon)
			})
		})
	})
}

// SnapValue snaps the value to [Slider.Step] if [Slider.EnforceStep] is on.
func (sr *Slider) SnapValue() {
	if !sr.EnforceStep {
		return
	}
	// round to the nearest step
	sr.Value = sr.Step * math32.Round(sr.Value/sr.Step)
}

// SendChanged sends a Changed message if given new value is
// different from the existing Value.
func (sr *Slider) SendChanged(e ...events.Event) bool {
	if sr.Value == sr.LastValue {
		return false
	}
	sr.LastValue = sr.Value
	sr.SendChange(e...)
	return true
}

// SliderSize returns the size available for sliding, based on allocation
func (sr *Slider) SliderSize() float32 {
	sz := sr.Geom.Size.Actual.Content.Dim(sr.Styles.Direction.Dim())
	if sr.Type != SliderScrollbar {
		thsz := sr.ThumbSizeDots()
		sz -= thsz.Dim(sr.Styles.Direction.Dim()) // half on each size
	}
	return sz
}

// SliderThickness returns the thickness of the slider: Content size in other dim.
func (sr *Slider) SliderThickness() float32 {
	return sr.Geom.Size.Actual.Content.Dim(sr.Styles.Direction.Dim().Other())
}

// ThumbSizeDots returns the thumb size in dots, based on ThumbSize
// and the content thickness
func (sr *Slider) ThumbSizeDots() math32.Vector2 {
	return sr.ThumbSize.MulScalar(sr.SliderThickness())
}

// SlideThumbSize returns thumb size, based on type
func (sr *Slider) SlideThumbSize() float32 {
	if sr.Type == SliderScrollbar {
		minsz := sr.SliderThickness()
		return max(math32.Clamp(sr.VisiblePct, 0, 1)*sr.SliderSize(), minsz)
	}
	return sr.ThumbSizeDots().Dim(sr.Styles.Direction.Dim())
}

// EffectiveMax returns the effective maximum value represented.
// For the Slider type, it it is just Max.
// for the Scrollbar type, it is Max - Value of thumb size
func (sr *Slider) EffectiveMax() float32 {
	if sr.Type == SliderScrollbar {
		return sr.Max - math32.Clamp(sr.VisiblePct, 0, 1)*(sr.Max-sr.Min)
	}
	return sr.Max
}

// ScrollThumbValue returns the current scroll VisiblePct
// in terms of the Min - Max range of values.
func (sr *Slider) ScrollThumbValue() float32 {
	return math32.Clamp(sr.VisiblePct, 0, 1) * (sr.Max - sr.Min)
}

// SetSliderPos sets the position of the slider at the given
// relative position within the usable Content sliding range,
// in pixels, and updates the corresponding Value based on that position.
func (sr *Slider) SetSliderPos(pos float32) {
	sz := sr.Geom.Size.Actual.Content.Dim(sr.Styles.Direction.Dim())
	if sz <= 0 {
		return
	}

	thsz := sr.SlideThumbSize()
	thszh := .5 * thsz
	sr.Pos = math32.Clamp(pos, thszh, sz-thszh)
	prel := (sr.Pos - thszh) / (sz - thsz)
	effmax := sr.EffectiveMax()
	val := math32.Truncate(sr.Min+prel*(effmax-sr.Min), sr.Prec)
	val = math32.Clamp(val, sr.Min, effmax)
	// fmt.Println(pos, thsz, prel, val)
	sr.Value = val
	sr.SnapValue()
	sr.SetPosFromValue(sr.Value) // go back the other way to be fully consistent
	sr.NeedsRender()
}

// SetSliderPosAction sets the position of the slider at the given position in pixels,
// and updates the corresponding Value based on that position.
// This version sends input events.
func (sr *Slider) SetSliderPosAction(pos float32) {
	sr.SetSliderPos(pos)
	if math32.Abs(sr.PrevSlide-sr.Value) > sr.InputThreshold {
		sr.PrevSlide = sr.Value
		sr.Send(events.Input)
	}
}

// SetPosFromValue sets the slider position based on the given value
// (typically rs.Value)
func (sr *Slider) SetPosFromValue(val float32) {
	sz := sr.Geom.Size.Actual.Content.Dim(sr.Styles.Direction.Dim())
	if sz <= 0 {
		return
	}

	effmax := sr.EffectiveMax()
	val = math32.Clamp(val, sr.Min, effmax)
	prel := (val - sr.Min) / (effmax - sr.Min) // relative position 0-1
	thsz := sr.SlideThumbSize()
	thszh := .5 * thsz
	sr.Pos = 0.5*thsz + prel*(sz-thsz)
	sr.Pos = math32.Clamp(sr.Pos, thszh, sz-thszh)
	sr.NeedsRender()
}

// SetVisiblePct sets the visible pct value for Scrollbar type.
func (sr *Slider) SetVisiblePct(val float32) *Slider {
	sr.VisiblePct = math32.Clamp(val, 0, 1)
	return sr
}

// SetValue sets the value and updates the slider position,
// but does not send a Change event (see Action version)
func (sr *Slider) SetValue(val float32) *Slider {
	effmax := sr.EffectiveMax()
	val = math32.Clamp(val, sr.Min, effmax)
	if sr.Value != val {
		sr.Value = val
		sr.SnapValue()
		sr.SetPosFromValue(val)
	}
	sr.NeedsRender()
	return sr
}

// SetValueAction sets the value and updates the slider representation, and
// emits an input and change event
func (sr *Slider) SetValueAction(val float32) {
	if sr.Value == val {
		return
	}
	sr.SetValue(val)
	sr.Send(events.Input)
	sr.SendChange()
}

func (sr *Slider) WidgetTooltip(pos image.Point) (string, image.Point) {
	res := sr.Tooltip
	if sr.Type == SliderScrollbar {
		return res, sr.DefaultTooltipPos()
	}
	if res != "" {
		res += " "
	}
	res += fmt.Sprintf("(value: %.4g, minimum: %.4g, maximum: %.4g)", sr.Value, sr.Min, sr.Max)
	return res, sr.DefaultTooltipPos()
}

// PointToRelPos translates a point in scene local pixel coords into relative
// position within the slider content range
func (sr *Slider) PointToRelPos(pt image.Point) float32 {
	ptf := math32.Vector2FromPoint(pt).Dim(sr.Styles.Direction.Dim())
	return ptf - sr.Geom.Pos.Content.Dim(sr.Styles.Direction.Dim())
}

// ScrollScale returns scaled value of scroll delta
// as a function of the step size.
func (sr *Slider) ScrollScale(del float32) float32 {
	return del * sr.Step
}

func (sr *Slider) Render() {
	sr.SetPosFromValue(sr.Value)

	pc := &sr.Scene.PaintContext
	st := &sr.Styles

	dim := sr.Styles.Direction.Dim()
	od := dim.Other()
	sz := sr.Geom.Size.Actual.Content
	pos := sr.Geom.Pos.Content

	pabg := sr.ParentActualBackground()

	if sr.Type == SliderScrollbar {
		pc.DrawStandardBox(st, pos, sz, pabg) // track
		if sr.ValueColor != nil {
			thsz := sr.SlideThumbSize()
			osz := sr.ThumbSizeDots().Dim(od)
			tpos := pos
			tpos = tpos.AddDim(dim, sr.Pos)
			tpos = tpos.SubDim(dim, thsz*.5)
			tsz := sz
			tsz.SetDim(dim, thsz)
			origsz := sz.Dim(od)
			tsz.SetDim(od, osz)
			tpos = tpos.AddDim(od, 0.5*(osz-origsz))
			vabg := sr.Styles.ComputeActualBackgroundFor(sr.ValueColor, pabg)
			pc.FillStyle.Color = vabg
			sr.RenderBoxImpl(tpos, tsz, styles.Border{Radius: st.Border.Radius}) // thumb
		}
	} else {
		prevbg := st.Background
		prevsl := st.StateLayer
		// use surrounding background with no state layer for surrounding box
		st.Background = pabg
		st.StateLayer = 0
		st.ComputeActualBackground(pabg)
		// surrounding box (needed to prevent it from rendering over itself)
		sr.RenderStandardBox()
		st.Background = prevbg
		st.StateLayer = prevsl
		st.ComputeActualBackground(pabg)

		trsz := sz.Dim(od) * sr.TrackSize
		bsz := sz
		bsz.SetDim(od, trsz)
		bpos := pos
		bpos = bpos.AddDim(od, .5*(sz.Dim(od)-trsz))
		pc.FillStyle.Color = st.ActualBackground
		sr.RenderBoxImpl(bpos, bsz, styles.Border{Radius: st.Border.Radius}) // track

		if sr.ValueColor != nil {
			bsz.SetDim(dim, sr.Pos)
			vabg := sr.Styles.ComputeActualBackgroundFor(sr.ValueColor, pabg)
			pc.FillStyle.Color = vabg
			sr.RenderBoxImpl(bpos, bsz, styles.Border{Radius: st.Border.Radius})
		}

		thsz := sr.ThumbSizeDots()
		tpos := pos
		tpos.SetDim(dim, pos.Dim(dim)+sr.Pos)
		tpos = tpos.AddDim(od, 0.5*sz.Dim(od)) // ctr

		// render thumb as icon or box
		if sr.Icon.IsSet() && sr.HasChildren() {
			ic := sr.Child(0).(*Icon)
			tpos.SetSub(thsz.MulScalar(.5))
			ic.Geom.Pos.Total = tpos
			ic.SetContentPosFromPos()
			ic.SetBBoxes()
		} else {
			tabg := sr.Styles.ComputeActualBackgroundFor(sr.ThumbColor, pabg)
			pc.FillStyle.Color = tabg
			tpos.SetSub(thsz.MulScalar(0.5))
			sr.RenderBoxImpl(tpos, thsz, styles.Border{Radius: st.Border.Radius})
		}
	}
}

func (sr *Slider) ScenePos() {
	sr.WidgetBase.ScenePos()
	if !sr.StayInView {
		return
	}
	pwb := sr.ParentWidget()
	zr := image.Rectangle{}
	if !pwb.IsVisible() || pwb.Geom.TotalBBox == zr {
		return
	}
	sbw := math32.Ceil(sr.Styles.ScrollBarWidth.Dots)
	scmax := math32.Vector2FromPoint(sr.Scene.Geom.ContentBBox.Max).SubScalar(sbw)
	sr.Geom.Pos.Total.SetMin(scmax)
	sr.SetContentPosFromPos()
	sr.SetBBoxesFromAllocs()
}
