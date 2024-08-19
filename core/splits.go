// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"strconv"
	"strings"

	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/system"
	"cogentcore.org/core/tree"
)

// Splits allocates a certain proportion of its space to each of its children
// along [styles.Style.Direction]. It adds [Handle] widgets to its parts that
// allow the user to customize the amount of space allocated to each child.
type Splits struct {
	Frame

	// Splits is the proportion (0-1 normalized, enforced) of space
	// allocated to each element. 0 indicates that an element should
	// be completely collapsed. By default, each element gets the
	// same amount of space.
	Splits []float32

	// SavedSplits is a saved version of the splits that can be restored
	// for dynamic collapse/expand operations.
	SavedSplits []float32 `set:"-"`
}

func (sl *Splits) Init() {
	sl.Frame.Init()
	sl.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
		s.Margin.Zero()
		s.Padding.Zero()
		s.Min.Y.Em(10)

		if sl.SizeClass() == SizeCompact {
			s.Direction = styles.Column
		} else {
			s.Direction = styles.Row
		}
	})
	sl.OnWidgetAdded(func(w Widget) {
		if w.AsTree().Parent == sl.This && w != sl.Parts { // TODO(config): need some way to do this with the new config paradigm
			w.AsWidget().Styler(func(s *styles.Style) {
				// splits elements must scroll independently and grow
				s.Overflow.Set(styles.OverflowAuto)
				s.Grow.Set(1, 1)
			})
		}
	})

	sl.OnKeyChord(func(e events.Event) {
		kc := string(e.KeyChord())
		mod := "Control+"
		if TheApp.Platform() == system.MacOS {
			mod = "Meta+"
		}
		if !strings.HasPrefix(kc, mod) {
			return
		}
		kns := kc[len(mod):]

		knc, err := strconv.Atoi(kns)
		if err != nil {
			return
		}
		kn := int(knc)
		if kn == 0 {
			e.SetHandled()
			sl.EvenSplits()
		} else if kn <= len(sl.Children) {
			e.SetHandled()
			if sl.Splits[kn-1] <= 0.01 {
				sl.RestoreChild(kn - 1)
			} else {
				sl.CollapseChild(true, kn-1)
			}
		}
	})

	sl.Updater(func() {
		sl.UpdateSplits()
	})
	parts := sl.NewParts()
	parts.Maker(func(p *tree.Plan) {
		for i := range len(sl.Children) - 1 { // one fewer handle than children
			tree.AddAt(p, "handle-"+strconv.Itoa(i), func(w *Handle) {
				w.OnChange(func(e events.Event) {
					sl.SetSplitAction(w.IndexInParent(), w.Value())
				})
				w.Styler(func(s *styles.Style) {
					s.Direction = sl.Styles.Direction
				})
			})
		}
	})
}

// UpdateSplits normalizes the splits and ensures that there are as
// many split proportions as children.
func (sl *Splits) UpdateSplits() *Splits {
	sz := len(sl.Children)
	if sz == 0 {
		return sl
	}
	if sl.Splits == nil || len(sl.Splits) != sz {
		sl.Splits = make([]float32, sz)
	}
	sum := float32(0)
	for _, sp := range sl.Splits {
		sum += sp
	}
	if sum == 0 { // set default even splits
		sl.EvenSplits()
		sum = 1
	} else {
		norm := 1 / sum
		for i := range sl.Splits {
			sl.Splits[i] *= norm
		}
	}
	return sl
}

// EvenSplits splits space evenly across all panels
func (sl *Splits) EvenSplits() {
	sz := len(sl.Children)
	if sz == 0 {
		return
	}
	even := 1.0 / float32(sz)
	for i := range sl.Splits {
		sl.Splits[i] = even
	}
	sl.NeedsLayout()
}

// SaveSplits saves the current set of splits in SavedSplits, for a later RestoreSplits
func (sl *Splits) SaveSplits() {
	sz := len(sl.Splits)
	if sz == 0 {
		return
	}
	if sl.SavedSplits == nil || len(sl.SavedSplits) != sz {
		sl.SavedSplits = make([]float32, sz)
	}
	copy(sl.SavedSplits, sl.Splits)
}

// RestoreSplits restores a previously saved set of splits (if it exists), does an update
func (sl *Splits) RestoreSplits() {
	if sl.SavedSplits == nil {
		return
	}
	sl.SetSplits(sl.SavedSplits...).NeedsLayout()
}

// CollapseChild collapses given child(ren) (sets split proportion to 0),
// optionally saving the prior splits for later Restore function -- does an
// Update -- triggered by double-click of splitter
func (sl *Splits) CollapseChild(save bool, idxs ...int) {
	if save {
		sl.SaveSplits()
	}
	sz := len(sl.Children)
	for _, idx := range idxs {
		if idx >= 0 && idx < sz {
			sl.Splits[idx] = 0
		}
	}
	sl.UpdateSplits()
	sl.NeedsLayout()
}

// RestoreChild restores given child(ren) -- does an Update
func (sl *Splits) RestoreChild(idxs ...int) {
	sz := len(sl.Children)
	for _, idx := range idxs {
		if idx >= 0 && idx < sz {
			sl.Splits[idx] = 1.0 / float32(sz)
		}
	}
	sl.UpdateSplits()
	sl.NeedsLayout()
}

// IsCollapsed returns true if given split number is collapsed
func (sl *Splits) IsCollapsed(idx int) bool {
	sz := len(sl.Children)
	if idx >= 0 && idx < sz {
		return sl.Splits[idx] < 0.01
	}
	return false
}

// SetSplitAction sets the new splitter value, for given splitter.
// New value is 0..1 value of position of that splitter.
// It is a sum of all the positions up to that point.
// Splitters are updated to ensure that selected position is achieved,
// while dividing remainder appropriately.
func (sl *Splits) SetSplitAction(idx int, nwval float32) {
	sz := len(sl.Splits)
	oldsum := float32(0)
	for i := 0; i <= idx; i++ {
		oldsum += sl.Splits[i]
	}
	delta := nwval - oldsum
	oldval := sl.Splits[idx]
	uval := oldval + delta
	if uval < 0 {
		uval = 0
		delta = -oldval
		nwval = oldsum + delta
	}
	rmdr := 1 - nwval
	if idx < sz-1 {
		oldrmdr := 1 - oldsum
		if oldrmdr <= 0 {
			if rmdr > 0 {
				dper := rmdr / float32((sz-1)-idx)
				for i := idx + 1; i < sz; i++ {
					sl.Splits[i] = dper
				}
			}
		} else {
			for i := idx + 1; i < sz; i++ {
				curval := sl.Splits[i]
				sl.Splits[i] = rmdr * (curval / oldrmdr) // proportional
			}
		}
	}
	sl.Splits[idx] = uval
	sl.UpdateSplits()
	sl.NeedsLayout()
}

func (sl *Splits) SizeDownSetAllocs(iter int) {
	sz := &sl.Geom.Size
	csz := sz.Alloc.Content
	dim := sl.Styles.Direction.Dim()
	od := dim.Other()
	cszd := csz.Dim(dim)
	cszod := csz.Dim(od)
	sl.UpdateSplits()
	sl.WidgetKidsIter(func(i int, kwi Widget, kwb *WidgetBase) bool {
		sw := math32.Round(sl.Splits[i] * cszd)
		ksz := &kwb.Geom.Size
		ksz.Alloc.Total.SetDim(dim, sw)
		ksz.Alloc.Total.SetDim(od, cszod)
		ksz.SetContentFromTotal(&ksz.Alloc)
		// ksz.Actual = ksz.Alloc
		return tree.Continue
	})
}

func (sl *Splits) Position() {
	if !sl.HasChildren() {
		sl.Frame.Position()
		return
	}
	sl.UpdateSplits()
	sl.ConfigScrolls()
	sl.PositionSplits()
	sl.PositionChildren()
}

func (sl *Splits) PositionSplits() {
	if sl.NumChildren() <= 1 {
		return
	}
	if sl.Parts != nil {
		sl.Parts.Geom.Size = sl.Geom.Size // inherit: allows bbox to include handle
	}
	dim := sl.Styles.Direction.Dim()
	od := dim.Other()
	csz := sl.Geom.Size.Alloc.Content // key to use Alloc here!  excludes gaps
	cszd := csz.Dim(dim)
	pos := float32(0)

	hand := sl.Parts.Child(0).(*Handle)
	hwd := hand.Geom.Size.Actual.Total.Dim(dim)
	hht := hand.Geom.Size.Actual.Total.Dim(od)
	mid := (csz.Dim(od) - hht) / 2

	sl.WidgetKidsIter(func(i int, kwi Widget, kwb *WidgetBase) bool {
		kwb.Geom.RelPos.SetZero()
		if i == 0 {
			return tree.Continue
		}
		sw := math32.Round(sl.Splits[i-1] * cszd)
		pos += sw + hwd
		kwb.Geom.RelPos.SetDim(dim, pos)
		hl := sl.Parts.Child(i - 1).(*Handle)
		hl.Geom.RelPos.SetDim(dim, pos-hwd)
		hl.Geom.RelPos.SetDim(od, mid)
		hl.Min = 0
		hl.Max = cszd
		hl.Pos = pos
		return tree.Continue
	})
}

func (sl *Splits) RenderWidget() {
	if sl.PushBounds() {
		sl.WidgetKidsIter(func(i int, kwi Widget, kwb *WidgetBase) bool {
			sp := sl.Splits[i]
			if sp <= 0.01 {
				kwb.SetState(true, states.Invisible)
			} else {
				kwb.SetState(false, states.Invisible)
			}
			kwi.RenderWidget()
			return tree.Continue
		})
		sl.RenderParts()
		sl.PopBounds()
	}
}
