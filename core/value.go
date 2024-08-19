// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"reflect"
	"strings"

	"cogentcore.org/core/base/labels"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
)

// Value is a widget that has an associated value representation.
// It can be bound to a value using [Bind].
type Value interface {
	Widget

	// WidgetValue returns the pointer to the associated value of the widget.
	WidgetValue() any
}

// ValueSetter is an optional interface that [Value]s can implement
// to customize how the associated widget value is set from the given value.
type ValueSetter interface {

	// SetWidgetValue sets the associated widget value from the given value.
	SetWidgetValue(value any) error
}

// OnBinder is an optional interface that [Value]s can implement to
// do something when the widget is bound to the given value.
type OnBinder interface {

	// OnBind is called when the widget is bound to given value.
	OnBind(value any)
}

// Bind binds the given value to the given [Value] such that the values of
// the two will be linked and updated appropriately after [events.Change] events
// and during [Widget.UpdateWidget]. It returns the widget to enable method chaining.
func Bind[T Value](value any, vw T) T {
	wb := vw.AsWidget()
	alreadyBound := wb.ValueUpdate != nil
	wb.ValueUpdate = func() {
		if vws, ok := any(vw).(ValueSetter); ok {
			ErrorSnackbar(vw, vws.SetWidgetValue(value))
		} else {
			ErrorSnackbar(vw, reflectx.SetRobust(vw.WidgetValue(), value))
		}
	}
	wb.ValueOnChange = func() {
		ErrorSnackbar(vw, reflectx.SetRobust(value, vw.WidgetValue()))
	}
	if alreadyBound {
		resetWidgetValue(vw)
	}
	wb.ValueTitle = labels.FriendlyTypeName(reflectx.NonPointerType(reflect.TypeOf(value)))
	if ob, ok := any(vw).(OnBinder); ok {
		ob.OnBind(value)
	}
	wb.ValueUpdate() // we update it with the initial value immediately
	return vw
}

// If we were already bound to another value previously, we first need to
// reset the widget value to zero to avoid any issues with the pointer from
// the old value persisting and being updated. For example, that issue happened
// with slice and map pointers persisting in forms when a new struct was set.
func resetWidgetValue(vw Value) {
	rv := reflect.ValueOf(vw.WidgetValue())
	if rv.IsValid() && rv.Type().Kind() == reflect.Pointer {
		rv.Elem().SetZero()
	}
}

// JoinValueTitle returns a [WidgetBase.ValueTitle] string composed
// of two elements, with a • separator, handling the cases where
// either or both can be empty.
func JoinValueTitle(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + " • " + b
	}
}

const shiftNewWindow = "[Shift: new window]"

// InitValueButton configures the given [Value] to open a dialog representing
// its value in accordance with the given dialog construction function when clicked.
// It also sets the tooltip of the widget appropriately. If allowReadOnly is false,
// the dialog will not be opened if the widget is read only. It also takes an optional
// function to call after the dialog is accepted.
func InitValueButton(v Value, allowReadOnly bool, make func(d *Body), after ...func()) {
	wb := v.AsWidget()
	// windows are never new on mobile
	if !TheApp.Platform().IsMobile() {
		wb.SetTooltip(shiftNewWindow)
	}
	wb.OnClick(func(e events.Event) {
		if allowReadOnly || !wb.IsReadOnly() {
			wb.ValueNewWindow = e.HasAnyModifier(key.Shift)
			OpenValueDialog(v, make, after...)
		}
	})
}

// OpenValueDialog opens a new value dialog for the given [Value] using the
// given function for constructing the dialog and the optional given function
// to call after the dialog is accepted.
func OpenValueDialog(v Value, make func(d *Body), after ...func()) {
	opv := reflectx.UnderlyingPointer(reflect.ValueOf(v.WidgetValue()))
	if !opv.IsValid() {
		return
	}
	obj := opv.Interface()
	if RecycleDialog(obj) {
		return
	}
	wb := v.AsWidget()
	d := NewBody().AddTitle(wb.ValueTitle)
	if text := strings.ReplaceAll(wb.Tooltip, shiftNewWindow, ""); text != "" {
		d.AddText(text)
	}
	make(d)

	// if we don't have anything specific for ok events,
	// we just register an OnClose event and skip the
	// OK and Cancel buttons
	if len(after) == 0 {
		d.OnClose(func(e events.Event) {
			wb.SendChange()
			wb.Update()
		})
	} else {
		// otherwise, we have to make the bottom bar
		d.AddBottomBar(func(parent Widget) {
			d.AddCancel(parent)
			d.AddOK(parent).OnClick(func(e events.Event) {
				after[0]()
				wb.SendChange()
				wb.Update()
			})
		})
	}

	if wb.ValueNewWindow {
		d.RunWindowDialog(v)
	} else {
		d.RunFullDialog(v)
	}
}
