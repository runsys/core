// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"fmt"
	"image"

	"goki.dev/girl/states"
	"goki.dev/goosi/events"
	"goki.dev/ki/v2"
)

func (wb *WidgetBase) EventMgr() *EventMgr {
	if wb.Sc == nil {
		return nil
	}
	return &wb.Sc.EventMgr
}

// On adds an event listener function for the given event type
func (wb *WidgetBase) On(etype events.Types, fun func(e events.Event)) Widget {
	wb.Listeners.Add(etype, fun)
	return wb.This().(Widget)
}

// Send sends an NEW event of given type to this widget,
// optionally starting from values in the given original event
// (recommended to include where possible).
// Do NOT send an existing event using this method if you
// want the Handled state to persist throughout the call chain;
// call HandleEvent directly for any existing events.
func (wb *WidgetBase) Send(typ events.Types, orig events.Event) {
	var e events.Event
	if orig != nil {
		e = orig.Clone()
		e.AsBase().Typ = typ
	} else {
		e = &events.Base{Typ: typ}
	}
	wb.This().(Widget).HandleEvent(e)
}

// HandleEvent sends the given event to all Listeners for that event type.
// It also checks if the State has changed and calls ApplyStyle if so.
// If more significant Config level changes are needed due to an event,
// the event handler must do this itself.
func (wb *WidgetBase) HandleEvent(ev events.Event) {
	if EventTrace {
		if ev.Type() != events.MouseMove {
			fmt.Println("Event to Widget:", wb.Path(), ev.String())
		}
	}
	state := wb.Style.State
	wb.Listeners.Call(ev)
	if wb.Style.State != state {
		wb.ApplyStyleUpdate(wb.Sc)
	}
}

// WidgetHandlers adds the default events for Widget objects.
func (wb *WidgetBase) WidgetHandlers() {
	wb.WidgetStateFromMouse()
	wb.LongHoverTooltip()
	wb.WidgetStateFromFocus()
}

// PosInBBox returns true if given position is within
// this node's win bbox (under read lock)
func (wb *WidgetBase) PosInBBox(pos image.Point) bool {
	wb.BBoxMu.RLock()
	defer wb.BBoxMu.RUnlock()
	return pos.In(wb.ScBBox)
}

// WidgetStateFromMouse updates all standard State flags based on mouse events,
// such as MouseDown / Up -> Active and MouseEnter / Leave -> Hovered.
// None of these "consume" the event by setting Handled flag, as they are
// designed to work in conjunction with more specific handlers.
func (wb *WidgetBase) WidgetStateFromMouse() {
	wb.On(events.MouseDown, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Activatable) {
			// fmt.Println("active:", wb)
			wb.SetState(true, states.Active)
		}
	})
	wb.On(events.MouseUp, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Activatable) {
			wb.SetState(false, states.Active)
		}
	})
	wb.On(events.Click, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Checkable) {
			wb.SetState(!wb.StateIs(states.Checked), states.Checked)
		}
		// fmt.Println("click", wb)
		if wb.AbilityIs(states.Focusable) {
			wb.GrabFocus()
		} else {
			wb.FocusClear()
		}
	})
	wb.On(events.DoubleClick, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		// if we are not double clickable, we just treat
		// it as a click event
		if !wb.AbilityIs(states.DoubleClickable) {
			wb.Send(events.Click, e)
		}
	})
	wb.On(events.MouseEnter, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Hoverable) {
			wb.SetState(true, states.Hovered)
		}
	})
	wb.On(events.MouseLeave, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Hoverable) {
			wb.SetState(false, states.Hovered)
		}
	})
	wb.On(events.LongHoverStart, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.LongHoverable) {
			wb.SetState(true, states.LongHovered)
		}
	})
	wb.On(events.LongHoverEnd, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.LongHoverable) {
			wb.SetState(false, states.LongHovered)
		}
	})
	wb.On(events.SlideStart, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Slideable) {
			// fmt.Println("sliding:", wb)
			wb.SetState(true, states.Sliding)
		}
	})
	wb.On(events.SlideStop, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Slideable) {
			wb.SetState(false, states.Sliding, states.Active)
			// fmt.Println("done sliding:", wb)
		}
	})
	wb.On(events.DragStart, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Draggable) {
			wb.SetState(true, states.Dragging)
		}
	})
	wb.On(events.Drop, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Draggable) {
			wb.SetState(false, states.Dragging, states.Active)
		}
	})
}

// 	if bb.IsDisabled() {
// 		if !strings.HasSuffix(bb.Class, "-action") { // not for menu-action, bar-action
// 			bb.SetSelected(!bb.StateIs(states.Selected))
// 			// bb.EmitSelectedSignal()
// 			bb.UpdateSig()
// 		}

// LongHoverTooltip listens for LongHoverEvent and pops up a tooltip.
// Most widgets should call this as part of their event handler methods.
func (wb *WidgetBase) LongHoverTooltip() {
	wb.On(events.LongHoverStart, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.Tooltip == "" {
			return
		}
		fmt.Println("long hover tooltip")
		e.SetHandled()
		// pos := wbb.WinBBox.Max
		// pos.X -= 20
		// mvp := wbb.Sc
		// PopupTooltip(wbb.Tooltip, pos.X, pos.Y, mvp, wbb.Nm)
	})
}

// WidgetStateFromFocus updates standard State flags based on Focus events
func (wb *WidgetBase) WidgetStateFromFocus() {
	wb.On(events.Focus, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Focusable) {
			wb.ScrollToMe()
			wb.SetState(true, states.Focused)
		}
	})
	wb.On(events.FocusLost, func(e events.Event) {
		if wb.StateIs(states.Disabled) {
			return
		}
		if wb.AbilityIs(states.Focusable) {
			wb.SetState(false, states.Focused)
		}
	})
}

// WidgetMouseEvents connects to either or both mouse events -- IMPORTANT: if
// you need to also connect to other mouse events, you must copy this code --
// all processing of a mouse event must happen within one function b/c there
// can only be one registered per receiver and event type.  sel = Left button
// events.Press event, toggles the selected state, and emits a SelectedEvent.
// ctxtMenu = connects to Right button events.Press event, and sends a
// WidgetSig WidgetContextMenu signal, followed by calling ContextMenu method
// -- signal can be used to change state prior to generating context menu,
// including setting a CtxtMenuFunc that removes all items and thus negates
// the presentation of any menu
/*
func (wb *WidgetBase) WidgetMouseEvents(sel, ctxtMenu bool) {
	if !sel && !ctxtMenu {
		return
	}
	wbwe.AddFunc(events.MouseUp, RegPri, func(recv, send ki.Ki, sig int64, d any) {
		me := d.(events.Event)
		if sel {
			if me.Action == events.Press && me.Button == events.Left {
				me.SetHandled()
				_, wbb := AsWidget(recv)
				wbb.SetSelected(!wbb.StateIs(states.Selected))
				wbb.EmitSelectedSignal()
				wbb.ApplyStyleUpdate(wbb.Sc)
			}
		}
		if ctxtMenu {
			if me.Action == events.Release && me.Button == events.Right {
				me.SetHandled()
				wi, wbb := AsWidget(recv)
				wbb.EmitContextMenuSignal()
				wi.ContextMenu()
			}
		}
	})
}
*/

/*
// EmitSelectedSignal emits the WidgetSelected signal for this widget
func (wb *WidgetBase) EmitSelectedSignal() {
	wb.WidgetSig.Emit(wb.This(), int64(WidgetSelected), nil)
}

// EmitFocusedSignal emits the WidgetFocused signal for this widget
func (wb *WidgetBase) EmitFocusedSignal() {
	wb.WidgetSig.Emit(wb.This(), int64(WidgetFocused), nil)
}

// EmitContextMenuSignal emits the WidgetContextMenu signal for this widget
func (wb *WidgetBase) EmitContextMenuSignal() {
	wb.WidgetSig.Emit(wb.This(), int64(WidgetContextMenu), nil)
}
*/

// FirstContainingPoint finds the first node whose WinBBox contains the given
// point -- nil if none.  If leavesOnly is set then only nodes that have no
// nodes (leaves, terminal nodes) will be considered
func (wb *WidgetBase) FirstContainingPoint(pt image.Point, leavesOnly bool) ki.Ki {
	var rval ki.Ki
	wb.WalkPre(func(k ki.Ki) bool {
		if k == wb.This() {
			return ki.Continue
		}
		if leavesOnly && k.HasChildren() {
			return ki.Continue
		}
		_, w := AsWidget(k)
		if w == nil || w.Is(ki.Deleted) || w.Is(ki.Destroyed) {
			// 3D?
			return ki.Break
		}
		if w.PosInBBox(pt) {
			rval = w.This()
			return ki.Break
		}
		return ki.Continue
	})
	return rval
}

///////////////////////////////////////////////////////////////////
//		Focus

// GrabFocus grabs the keyboard input focus on this item or the first item within it
// that can be focused (if none, then goes ahead and sets focus to this object)
func (wb *WidgetBase) GrabFocus() {
	foc := wb.This().(Widget)
	if !foc.AbilityIs(states.Focusable) {
		foc = wb.FocusableInMe()
	}
	em := wb.EventMgr()
	if em != nil {
		// fmt.Println("grab focus:", foc)
		em.GrabFocus(foc.(Widget)) // doesn't send event
	}
}

// FocusableInMe returns the first Focusable element within this widget
func (wb *WidgetBase) FocusableInMe() Widget {
	var foc Widget
	wb.WalkPre(func(k ki.Ki) bool {
		kwi, kwb := AsWidget(k)
		if kwb == nil || kwb.This() == nil || kwb.Is(ki.Deleted) || kwb.Is(ki.Destroyed) {
			return ki.Break
		}
		if !kwb.AbilityIs(states.Focusable) {
			return ki.Continue
		}
		foc = kwi
		return ki.Break // done
	})
	return foc
}

// FocusNext moves the focus onto the next item
func (wb *WidgetBase) FocusNext() {
	em := wb.EventMgr()
	if em != nil {
		em.FocusNext()
	}
}

// FocusPrev moves the focus onto the previous item
func (wb *WidgetBase) FocusPrev() {
	em := wb.EventMgr()
	if em != nil {
		em.FocusPrev()
	}
}

// FocusClear resets focus to nil, but keeps the previous focus to pick up next time..
func (wb *WidgetBase) FocusClear() {
	em := wb.EventMgr()
	if em != nil {
		em.FocusClear()
	}
}

// StartFocus specifies this widget to give focus to when the window opens
func (wb *WidgetBase) StartFocus() {
	em := wb.EventMgr()
	if em != nil {
		em.SetStartFocus(wb.This().(Widget))
	}
}

// ContainsFocus returns true if this widget contains the current focus widget
// as maintained in the RenderWin
func (wb *WidgetBase) ContainsFocus() bool {
	em := wb.EventMgr()
	if em == nil {
		return false
	}
	cur := em.Focus
	if cur == nil {
		return false
	}
	if cur == wb.This() {
		return true
	}
	plev := cur.ParentLevel(wb.This())
	if plev < 0 {
		return false
	}
	return true
}
