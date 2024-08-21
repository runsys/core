// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"image"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"github.com/stretchr/testify/assert"
)

func TestHandleWidgetState(t *testing.T) {
	b := NewBody()
	w := NewWidgetBase(b)

	test := func(ability abilities.Abilities, state states.States, event events.Types, endEvent events.Types) {
		expect := states.States(0)
		assert.Equal(t, expect, w.Styles.State)

		w.Send(event)
		assert.Equal(t, expect, w.Styles.State)

		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, ability)
		})
		w.Style()

		w.Send(event)
		expect.SetFlag(true, state)
		assert.Equal(t, expect, w.Styles.State)

		w.Send(endEvent)
		expect.SetFlag(false, state)
		assert.Equal(t, expect, w.Styles.State)
	}

	test(abilities.Activatable, states.Active, events.MouseDown, events.MouseUp)
	test(abilities.LongPressable, states.LongPressed, events.LongPressStart, events.LongPressEnd)
	test(abilities.Hoverable, states.Hovered, events.MouseEnter, events.MouseLeave)
	test(abilities.LongHoverable, states.LongHovered, events.LongHoverStart, events.LongHoverEnd)
	test(abilities.Slideable, states.Sliding, events.SlideStart, events.SlideStop)
	test(abilities.Draggable, states.Dragging, events.DragStart, events.Drop)
	test(abilities.Focusable, states.Focused, events.Focus, events.FocusLost)
	test(abilities.Checkable, states.Checked, events.Click, events.Click)
}

func TestWidgetEventManager(t *testing.T) {
	b := NewBody()
	w := NewWidgetBase(b)

	assert.Equal(t, &w.Scene.Events, w.Events())

	b.AssertRender(t, "events/event-manager", func() {
		assert.Equal(t, w.Scene.RenderWindow().SystemWindow.Events(), w.SystemEvents())
		assert.Equal(t, w.Events().Clipboard(), w.Clipboard())
	})
}

func TestSystemEvents(t *testing.T) {
	b := NewBody()
	bt := NewButton(b).SetText("Button")

	b.AssertRender(t, "events/system-events", func() {
		bt.SystemEvents().MouseButton(events.MouseDown, events.Left, image.Pt(20, 20), 0)
	}, func() {
		expect := states.States(0)
		expect.SetFlag(true, states.Active)
		assert.Equal(t, expect, bt.Styles.State)
	})
}

func TestWidgetPrev(t *testing.T) {
	b := NewBody()
	NewTextField(b).AddClearButton()
	NewTextField(b).SetLeadingIcon(icons.Search)
	lt := NewTextField(b)
	b.UpdateTree()

	paths := []string{
		"/body scene/body/text-field-1/lead-icon/icon",
		"/body scene/body/text-field-1/lead-icon",
		"/body scene/body/text-field-1",
		"/body scene/body/text-field-0/trail-icon/icon",
		"/body scene/body/text-field-0/trail-icon",
		"/body scene/body/text-field-0/trail-icon-stretch",
		"/body scene/body/text-field-0",
		"/body scene/body",
		"/body scene",
	}
	i := 0
	widgetPrevFunc(lt, func(w Widget) bool {
		have := w.AsTree().Path()
		want := paths[i]
		if have != want {
			t.Errorf("expected\n%s\n\tbut got\n%s", want, have)
		}
		i++
		return false
	})
}

func TestWidgetNext(t *testing.T) {
	b := NewBody()
	ft := NewTextField(b).AddClearButton()
	NewTextField(b).SetLeadingIcon(icons.Search)
	NewTextField(b)
	b.UpdateTree()

	paths := []string{
		"/body scene/body/text-field-0/trail-icon-stretch",
		"/body scene/body/text-field-0/trail-icon",
		"/body scene/body/text-field-0/trail-icon/icon",
		"/body scene/body/text-field-1",
		"/body scene/body/text-field-1/lead-icon",
		"/body scene/body/text-field-1/lead-icon/icon",
		"/body scene/body/text-field-2",
	}
	i := 0
	widgetNextFunc(ft, func(w Widget) bool {
		have := w.AsTree().Path()
		want := paths[i]
		if have != want {
			t.Errorf("expected\n%s\n\tbut got\n%s", want, have)
		}
		i++
		return false
	})
}

func ExampleWidgetBase_AddCloseDialog() {
	b := NewBody()
	b.AddCloseDialog(func(d *Body) bool {
		d.AddTitle("Are you sure?").AddText("Are you sure you want to close the Cogent Core Demo?")
		d.AddBottomBar(func(parent Widget) {
			d.AddOK(parent).SetText("Close").OnClick(func(e events.Event) {
				b.Scene.Close()
			})
		})
		return true
	})
	b.RunMainWindow()
}
