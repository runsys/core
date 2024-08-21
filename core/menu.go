// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"image"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
)

// StyleMenuScene configures the default styles
// for the given pop-up menu frame with the given parent.
// It should be called on menu frames when they are created.
func StyleMenuScene(msc *Scene) {
	msc.Styler(func(s *styles.Style) {
		s.Grow.Set(0, 0)
		s.Padding.Set(units.Dp(2))
		s.Border.Radius = styles.BorderRadiusExtraSmall
		s.Background = colors.Scheme.SurfaceContainer
		s.BoxShadow = styles.BoxShadow2()
		s.Gap.Zero()
	})
	msc.SetOnChildAdded(func(n tree.Node) {
		if bt := AsButton(n); bt != nil {
			bt.Type = ButtonMenu
			bt.OnKeyChord(func(e events.Event) {
				kf := keymap.Of(e.KeyChord())
				switch kf {
				case keymap.MoveRight:
					if bt.openMenu(e) {
						e.SetHandled()
					}
				case keymap.MoveLeft:
					// need to be able to use arrow keys to navigate in completer
					if msc.Stage.Type != CompleterStage {
						msc.Stage.ClosePopup()
						e.SetHandled()
					}
				}
			})
			return
		}
		if sp, ok := n.(*Separator); ok {
			sp.Styler(func(s *styles.Style) {
				s.Direction = styles.Row
			})
		}
	})
}

// newMenuScene constructs a [Scene] for displaying a menu, using the
// given menu constructor function. If no name is provided, it defaults
// to "menu".  If no menu items added, returns nil.
func newMenuScene(menu func(m *Scene), name ...string) *Scene {
	nm := "menu"
	if len(name) > 0 {
		nm = name[0] + "-menu"
	}
	msc := NewScene(nm)
	StyleMenuScene(msc)
	menu(msc)
	if !msc.HasChildren() {
		return nil
	}

	hasSelected := false
	msc.WidgetWalkDown(func(cw Widget, cwb *WidgetBase) bool {
		if cw == msc {
			return tree.Continue
		}
		if bt := AsButton(cw); bt != nil {
			if bt.Menu == nil {
				bt.handleClickDismissMenu()
			}
		}
		if !hasSelected && cwb.StateIs(states.Selected) {
			// fmt.Println("start focus sel:", wb)
			msc.Events.SetStartFocus(cwb)
			hasSelected = true
		}
		return tree.Continue
	})
	if !hasSelected && msc.HasChildren() {
		// fmt.Println("start focus first:", msc.Child(0).(Widget))
		msc.Events.SetStartFocus(msc.Child(0).(Widget))
	}
	return msc
}

// NewMenuStage returns a new Menu stage with given scene contents,
// in connection with given widget, which provides key context
// for constructing the menu, at given RenderWindow position
// (e.g., use ContextMenuPos or WinPos method on ctx Widget).
// Make further configuration choices using Set* methods, which
// can be chained directly after the New call.
// Use Run call at the end to start the Stage running.
func NewMenuStage(sc *Scene, ctx Widget, pos image.Point) *Stage {
	if sc == nil || !sc.HasChildren() {
		return nil
	}
	st := NewPopupStage(MenuStage, sc, ctx)
	if pos != (image.Point{}) {
		st.Pos = pos
	}
	return st
}

// NewMenu returns a new menu stage based on the given menu constructor
// function, in connection with given widget, which provides key context
// for constructing the menu at given RenderWindow position
// (e.g., use ContextMenuPos or WinPos method on ctx Widget).
// Make further configuration choices using Set* methods, which
// can be chained directly after the New call.
// Use Run call at the end to start the Stage running.
func NewMenu(menu func(m *Scene), ctx Widget, pos image.Point) *Stage {
	return NewMenuStage(newMenuScene(menu, ctx.AsTree().Name), ctx, pos)
}

// AddContextMenu adds the given context menu to [WidgetBase.ContextMenus].
// It is the main way that code should modify a widget's context menus.
// Context menu functions are run in reverse order, and separators are
// automatically added between each context menu function.
func (wb *WidgetBase) AddContextMenu(menu func(m *Scene)) {
	wb.ContextMenus = append(wb.ContextMenus, menu)
}

// applyContextMenus adds the [Widget.ContextMenus] to the given menu scene
// in reverse order. It also adds separators between each context menu function.
func (wb *WidgetBase) applyContextMenus(m *Scene) {
	for i := len(wb.ContextMenus) - 1; i >= 0; i-- {
		wb.ContextMenus[i](m)

		nc := m.NumChildren()
		// we delete any extra separator
		if nc > 0 {
			if _, ok := m.Child(nc - 1).(*Separator); ok {
				m.DeleteChildAt(nc - 1)
			}
		}
		if i != 0 {
			NewSeparator(m)
		}
	}
}

// ContextMenuPos returns the default position for the context menu
// upper left corner.  The event will be from a mouse ContextMenu
// event if non-nil: should handle both cases.
func (wb *WidgetBase) ContextMenuPos(e events.Event) image.Point {
	if e != nil {
		return e.WindowPos()
	}
	return wb.winPos(.5, .5) // center
}

func (wb *WidgetBase) handleWidgetContextMenu() {
	wb.On(events.ContextMenu, func(e events.Event) {
		wi := wb.This.(Widget)
		wi.ShowContextMenu(e)
	})
}

func (wb *WidgetBase) ShowContextMenu(e events.Event) {
	e.SetHandled() // always
	wi := wb.This.(Widget)
	nm := NewMenu(wi.AsWidget().applyContextMenus, wi, wi.ContextMenuPos(e))
	if nm == nil { // no items
		return
	}
	nm.Run()
}

// NewMenuFromStrings constructs a new menu from given list of strings,
// calling the given function with the index of the selected string.
// if string == sel, that menu item is selected initially.
func NewMenuFromStrings(strs []string, sel string, fun func(idx int)) *Scene {
	return newMenuScene(func(m *Scene) {
		for i, s := range strs {
			b := NewButton(m).SetText(s)
			b.OnClick(func(e events.Event) {
				fun(i)
			})
			if s == sel {
				b.SetSelected(true)
			}
		}
	})
}
