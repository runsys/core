// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"image"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
)

// AutoScrollRate determines the rate of auto-scrolling of layouts
var AutoScrollRate = float32(1)

// HasAnyScroll returns true if the frame has any scrollbars.
func (fr *Frame) HasAnyScroll() bool {
	return fr.HasScroll[math32.X] || fr.HasScroll[math32.Y]
}

// ScrollGeom returns the target position and size for scrollbars
func (fr *Frame) ScrollGeom(d math32.Dims) (pos, sz math32.Vector2) {
	sbw := math32.Ceil(fr.Styles.ScrollBarWidth.Dots)
	od := d.Other()
	bbmin := math32.Vector2FromPoint(fr.Geom.ContentBBox.Min)
	bbmax := math32.Vector2FromPoint(fr.Geom.ContentBBox.Max)
	if fr.This != fr.Scene.This { // if not the scene, keep inside the scene
		bbmin.SetMax(math32.Vector2FromPoint(fr.Scene.Geom.ContentBBox.Min))
		bbmax.SetMin(math32.Vector2FromPoint(fr.Scene.Geom.ContentBBox.Max).SubScalar(sbw))
	}
	pos.SetDim(d, bbmin.Dim(d))
	pos.SetDim(od, bbmax.Dim(od))
	bbsz := bbmax.Sub(bbmin)
	sz.SetDim(d, bbsz.Dim(d)-4)
	sz.SetDim(od, sbw)
	sz = sz.Ceil()
	return
}

// ConfigScrolls configures any scrollbars that have been enabled
// during the Layout process. This is called during Position, once
// the sizing and need for scrollbars has been established.
// The final position of the scrollbars is set during ScenePos in
// PositionScrolls.  Scrolls are kept around in general.
func (fr *Frame) ConfigScrolls() {
	for d := math32.X; d <= math32.Y; d++ {
		if fr.HasScroll[d] {
			fr.ConfigScroll(d)
		}
	}
}

// ConfigScroll configures scroll for given dimension
func (fr *Frame) ConfigScroll(d math32.Dims) {
	if fr.Scrolls[d] != nil {
		return
	}
	fr.Scrolls[d] = NewSlider()
	sb := fr.Scrolls[d]
	tree.SetParent(sb, fr)
	// sr.SetFlag(true, tree.Field) // note: do not turn on -- breaks pos
	sb.SetType(SliderScrollbar)
	sb.InputThreshold = 1
	sb.Min = 0.0
	sb.Styler(func(s *styles.Style) {
		s.Direction = styles.Directions(d)
		s.Padding.Zero()
		s.Margin.Zero()
		s.MaxBorder.Width.Zero()
		s.Border.Width.Zero()
		s.FillMargin = false
	})
	sb.FinalStyler(func(s *styles.Style) {
		od := d.Other()
		_, sz := fr.This.(Layouter).ScrollGeom(d)
		if sz.X > 0 && sz.Y > 0 {
			s.SetState(false, states.Invisible)
			s.Min.SetDim(d, units.Dot(sz.Dim(d)))
			s.Min.SetDim(od, units.Dot(sz.Dim(od)))
		} else {
			s.SetState(true, states.Invisible)
		}
		s.Max = s.Min
	})
	sb.OnInput(func(e events.Event) {
		e.SetHandled()
		fr.This.(Layouter).ScrollChanged(d, sb)
	})
	sb.Update()
}

// ScrollChanged is called in the OnInput event handler for updating,
// when the scrollbar value has changed, for given dimension.
// This is part of the Layouter interface.
func (fr *Frame) ScrollChanged(d math32.Dims, sb *Slider) {
	fr.Geom.Scroll.SetDim(d, -sb.Value)
	fr.This.(Layouter).ScenePos() // computes updated positions
	fr.NeedsRender()
}

// ScrollUpdateFromGeom updates the scrollbar for given dimension
// based on the current Geom.Scroll value for that dimension.
// This can be used to programatically update the scroll value.
func (fr *Frame) ScrollUpdateFromGeom(d math32.Dims) {
	if !fr.HasScroll[d] || fr.Scrolls[d] == nil {
		return
	}
	sb := fr.Scrolls[d]
	cv := fr.Geom.Scroll.Dim(d)
	sb.SetValueAction(-cv)
	fr.This.(Layouter).ScenePos() // computes updated positions
	fr.NeedsRender()
}

// ScrollValues returns the maximum size that could be scrolled,
// the visible size (which could be less than the max size, in which
// case no scrollbar is needed), and visSize / maxSize as the VisiblePct.
// This is used in updating the scrollbar and determining whether one is
// needed in the first place
func (fr *Frame) ScrollValues(d math32.Dims) (maxSize, visSize, visPct float32) {
	sz := &fr.Geom.Size
	maxSize = sz.Internal.Dim(d)
	visSize = sz.Alloc.Content.Dim(d)
	visPct = visSize / maxSize
	return
}

// SetScrollParams sets scrollbar parameters.  Must set Step and PageStep,
// but can also set others as needed.
// Max and VisiblePct are automatically set based on ScrollValues maxSize, visPct.
func (fr *Frame) SetScrollParams(d math32.Dims, sb *Slider) {
	sb.Step = fr.Styles.Font.Size.Dots // step by lines
	sb.PageStep = 10.0 * sb.Step       // todo: more dynamic
}

// PositionScrolls arranges scrollbars
func (fr *Frame) PositionScrolls() {
	for d := math32.X; d <= math32.Y; d++ {
		if fr.HasScroll[d] && fr.Scrolls[d] != nil {
			fr.PositionScroll(d)
		} else {
			fr.Geom.Scroll.SetDim(d, 0)
		}
	}
}

func (fr *Frame) PositionScroll(d math32.Dims) {
	sb := fr.Scrolls[d]
	pos, ssz := fr.This.(Layouter).ScrollGeom(d)
	maxSize, _, visPct := fr.This.(Layouter).ScrollValues(d)
	if sb.Geom.Pos.Total == pos && sb.Geom.Size.Actual.Content == ssz && sb.VisiblePct == visPct {
		return
	}
	if ssz.X <= 0 || ssz.Y <= 0 {
		sb.SetState(true, states.Invisible)
		return
	}
	sb.SetState(false, states.Invisible)
	sb.Max = maxSize
	sb.SetVisiblePct(visPct)
	// fmt.Println(ly, d, "vis pct:", asz/csz)
	sb.SetValue(sb.Value) // keep in range
	fr.This.(Layouter).SetScrollParams(d, sb)

	sb.Update() // applies style
	sb.SizeUp()
	sb.Geom.Size.Alloc = fr.Geom.Size.Actual
	sb.SizeDown(0)

	sb.Geom.Pos.Total = pos
	sb.SetContentPosFromPos()
	// note: usually these are intersected with parent *content* bbox,
	// but scrolls are specifically outside of that.
	sb.SetBBoxesFromAllocs()
}

// RenderScrolls draws the scrollbars
func (fr *Frame) RenderScrolls() {
	for d := math32.X; d <= math32.Y; d++ {
		if fr.HasScroll[d] && fr.Scrolls[d] != nil {
			fr.Scrolls[d].RenderWidget()
		}
	}
}

// SetScrollsOff turns off the scrollbars
func (fr *Frame) SetScrollsOff() {
	for d := math32.X; d <= math32.Y; d++ {
		fr.HasScroll[d] = false
	}
}

// ScrollActionDelta moves the scrollbar in given dimension by given delta
// and emits a ScrollSig signal.
func (fr *Frame) ScrollActionDelta(d math32.Dims, delta float32) {
	if fr.HasScroll[d] && fr.Scrolls[d] != nil {
		sb := fr.Scrolls[d]
		nval := sb.Value + sb.ScrollScale(delta)
		sb.SetValueAction(nval)
		fr.NeedsRender() // only render needed -- scroll updates pos
	}
}

// ScrollActionPos moves the scrollbar in given dimension to given
// position and emits a ScrollSig signal.
func (fr *Frame) ScrollActionPos(d math32.Dims, pos float32) {
	if fr.HasScroll[d] && fr.Scrolls[d] != nil {
		sb := fr.Scrolls[d]
		sb.SetValueAction(pos)
		fr.NeedsRender()
	}
}

// ScrollToPos moves the scrollbar in given dimension to given
// position and DOES NOT emit a ScrollSig signal.
func (fr *Frame) ScrollToPos(d math32.Dims, pos float32) {
	if fr.HasScroll[d] && fr.Scrolls[d] != nil {
		sb := fr.Scrolls[d]
		sb.SetValueAction(pos)
		fr.NeedsRender()
	}
}

// ScrollDelta processes a scroll event.  If only one dimension is processed,
// and there is a non-zero in other, then the consumed dimension is reset to 0
// and the event is left unprocessed, so a higher level can consume the
// remainder.
func (fr *Frame) ScrollDelta(e events.Event) {
	se := e.(*events.MouseScroll)
	fdel := se.Delta

	hasShift := e.HasAnyModifier(key.Shift, key.Alt) // shift or alt indicates to scroll horizontally
	if hasShift {
		if !fr.HasScroll[math32.X] { // if we have shift, we can only horizontal scroll
			return
		}
		fr.ScrollActionDelta(math32.X, fdel.Y)
		return
	}

	if fr.HasScroll[math32.Y] && fr.HasScroll[math32.X] {
		fr.ScrollActionDelta(math32.Y, fdel.Y)
		fr.ScrollActionDelta(math32.X, fdel.X)
	} else if fr.HasScroll[math32.Y] {
		fr.ScrollActionDelta(math32.Y, fdel.Y)
		if se.Delta.X != 0 {
			se.Delta.Y = 0
		}
	} else if fr.HasScroll[math32.X] {
		if se.Delta.X != 0 {
			fr.ScrollActionDelta(math32.X, fdel.X)
			if se.Delta.Y != 0 {
				se.Delta.X = 0
			}
		}
	}
}

// ParentScrollFrame returns the first parent frame that has active scrollbars.
func (wb *WidgetBase) ParentScrollFrame() *Frame {
	ly := tree.ParentByType[Layouter](wb)
	if ly == nil {
		return nil
	}
	fr := ly.AsFrame()
	if fr.HasAnyScroll() {
		return fr
	}
	return fr.ParentScrollFrame()
}

// ScrollToThis tells this widget's parent frame to scroll to keep
// this widget in view. It returns whether any scrolling was done.
func (wb *WidgetBase) ScrollToThis() bool {
	ly := wb.ParentScrollFrame()
	if ly == nil {
		return false
	}
	return ly.ScrollToWidget(wb.This.(Widget))
}

// ScrollToWidget scrolls the layout to ensure that the given widget is in view.
// It returns whether scrolling was needed.
func (fr *Frame) ScrollToWidget(w Widget) bool {
	// note: critical to NOT use BBox b/c it is zero for invisible items!
	return fr.ScrollToBox(w.AsWidget().Geom.TotalRect())
}

// AutoScrollDim auto-scrolls along one dimension, based on the current
// position value, which is in the current scroll value range.
func (fr *Frame) AutoScrollDim(d math32.Dims, pos float32) bool {
	if !fr.HasScroll[d] || fr.Scrolls[d] == nil {
		return false
	}
	sb := fr.Scrolls[d]
	smax := sb.EffectiveMax()
	ssz := sb.ScrollThumbValue()
	dst := sb.Step * AutoScrollRate

	mind := max(0, (pos - sb.Value))
	maxd := max(0, (sb.Value+ssz)-pos)

	if mind <= maxd {
		pct := mind / ssz
		if pct < .1 && sb.Value > 0 {
			dst = min(dst, sb.Value)
			sb.SetValueAction(sb.Value - dst)
			return true
		}
	} else {
		pct := maxd / ssz
		if pct < .1 && sb.Value < smax {
			dst = min(dst, (smax - sb.Value))
			sb.SetValueAction(sb.Value + dst)
			return true
		}
	}
	return false
}

var LayoutLastAutoScroll time.Time

// AutoScroll scrolls the layout based on given position in scroll
// coordinates (i.e., already subtracing the BBox Min for a mouse event).
func (fr *Frame) AutoScroll(pos math32.Vector2) bool {
	now := time.Now()
	lag := now.Sub(LayoutLastAutoScroll)
	if lag < SystemSettings.LayoutAutoScrollDelay {
		return false
	}
	did := false
	if fr.HasScroll[math32.Y] && fr.HasScroll[math32.X] {
		did = fr.AutoScrollDim(math32.Y, pos.Y)
		did = did || fr.AutoScrollDim(math32.X, pos.X)
	} else if fr.HasScroll[math32.Y] {
		did = fr.AutoScrollDim(math32.Y, pos.Y)
	} else if fr.HasScroll[math32.X] {
		did = fr.AutoScrollDim(math32.X, pos.X)
	}
	if did {
		LayoutLastAutoScroll = time.Now()
	}
	return did
}

// ScrollToBoxDim scrolls to ensure that given target [min..max] range
// along one dimension is in view. Returns true if scrolling was needed
func (fr *Frame) ScrollToBoxDim(d math32.Dims, tmini, tmaxi int) bool {
	if !fr.HasScroll[d] || fr.Scrolls[d] == nil {
		return false
	}
	sb := fr.Scrolls[d]
	if sb == nil || sb.This == nil {
		return false
	}
	tmin, tmax := float32(tmini), float32(tmaxi)
	cmin, cmax := fr.Geom.ContentRangeDim(d)
	if tmin >= cmin && tmax <= cmax {
		return false
	}
	h := fr.Styles.Font.Size.Dots
	if tmin < cmin { // favors scrolling to start
		trg := sb.Value + tmin - cmin - h
		if trg < 0 {
			trg = 0
		}
		sb.SetValueAction(trg)
		return true
	} else {
		if (tmax - tmin) < sb.ScrollThumbValue() { // only if whole thing fits
			trg := sb.Value + float32(tmax-cmax) + h
			sb.SetValueAction(trg)
			return true
		}
	}
	return false
}

// ScrollToBox scrolls the layout to ensure that given rect box is in view.
// Returns true if scrolling was needed
func (fr *Frame) ScrollToBox(box image.Rectangle) bool {
	did := false
	if fr.HasScroll[math32.Y] && fr.HasScroll[math32.X] {
		did = fr.ScrollToBoxDim(math32.Y, box.Min.Y, box.Max.Y)
		did = did || fr.ScrollToBoxDim(math32.X, box.Min.X, box.Max.X)
	} else if fr.HasScroll[math32.Y] {
		did = fr.ScrollToBoxDim(math32.Y, box.Min.Y, box.Max.Y)
	} else if fr.HasScroll[math32.X] {
		did = fr.ScrollToBoxDim(math32.X, box.Min.X, box.Max.X)
	}
	if did {
		fr.NeedsRender()
	}
	return did
}

// ScrollDimToStart scrolls to put the given child coordinate position (eg.,
// top / left of a view box) at the start (top / left) of our scroll area, to
// the extent possible. Returns true if scrolling was needed.
func (fr *Frame) ScrollDimToStart(d math32.Dims, posi int) bool {
	if !fr.HasScroll[d] {
		return false
	}
	pos := float32(posi)
	cmin, _ := fr.Geom.ContentRangeDim(d)
	if pos == cmin {
		return false
	}
	sb := fr.Scrolls[d]
	trg := math32.Clamp(sb.Value+(pos-cmin), 0, sb.EffectiveMax())
	sb.SetValueAction(trg)
	return true
}

// ScrollDimToEnd scrolls to put the given child coordinate position (eg.,
// bottom / right of a view box) at the end (bottom / right) of our scroll
// area, to the extent possible. Returns true if scrolling was needed.
func (fr *Frame) ScrollDimToEnd(d math32.Dims, posi int) bool {
	if !fr.HasScroll[d] || fr.Scrolls[d] == nil {
		return false
	}
	pos := float32(posi)
	_, cmax := fr.Geom.ContentRangeDim(d)
	if pos == cmax {
		return false
	}
	sb := fr.Scrolls[d]
	trg := math32.Clamp(sb.Value+(pos-cmax), 0, sb.EffectiveMax())
	sb.SetValueAction(trg)
	return true
}

// ScrollDimToContentEnd is a helper function that scrolls the layout to the
// end of its content (ie: moves the scrollbar to the very end).
func (fr *Frame) ScrollDimToContentEnd(d math32.Dims) bool {
	end := fr.Geom.Pos.Content.Dim(d) + fr.Geom.Size.Internal.Dim(d)
	return fr.ScrollDimToEnd(d, int(end))
}

// ScrollDimToCenter scrolls to put the given child coordinate position (eg.,
// middle of a view box) at the center of our scroll area, to the extent
// possible. Returns true if scrolling was needed.
func (fr *Frame) ScrollDimToCenter(d math32.Dims, posi int) bool {
	if !fr.HasScroll[d] || fr.Scrolls[d] == nil {
		return false
	}
	pos := float32(posi)
	cmin, cmax := fr.Geom.ContentRangeDim(d)
	mid := 0.5 * (cmin + cmax)
	if pos == mid {
		return false
	}
	sb := fr.Scrolls[d]
	trg := math32.Clamp(sb.Value+(pos-mid), 0, sb.EffectiveMax())
	sb.SetValueAction(trg)
	return true
}
