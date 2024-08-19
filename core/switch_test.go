// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"testing"

	"cogentcore.org/core/events"
	"github.com/stretchr/testify/assert"
)

func TestSwitch(t *testing.T) {
	b := NewBody()
	NewSwitch(b)
	b.AssertRender(t, "switch/basic")
}

func TestSwitchText(t *testing.T) {
	b := NewBody()
	NewSwitch(b).SetText("Remember me")
	b.AssertRender(t, "switch/text")
}

func TestSwitchChip(t *testing.T) {
	b := NewBody()
	NewSwitch(b).SetType(SwitchChip).SetText("Remember me")
	b.AssertRender(t, "switch/chip")
}

func TestSwitchCheckbox(t *testing.T) {
	b := NewBody()
	NewSwitch(b).SetType(SwitchCheckbox).SetText("Remember me")
	b.AssertRender(t, "switch/checkbox")
}

func TestSwitchRadioButton(t *testing.T) {
	b := NewBody()
	NewSwitch(b).SetType(SwitchRadioButton).SetText("Remember me")
	b.AssertRender(t, "switch/radio-button")
}

func TestSwitchChipChecked(t *testing.T) {
	b := NewBody()
	NewSwitch(b).SetType(SwitchChip).SetText("Remember me").SetChecked(true)
	b.AssertRender(t, "switch/chip-checked")
}

func TestSwitchChange(t *testing.T) {
	b := NewBody()
	sw := NewSwitch(b)
	checked := false
	sw.OnChange(func(e events.Event) {
		checked = sw.IsChecked()
	})
	b.AssertRender(t, "switch/change", func() {
		sw.Send(events.Click)
		assert.True(t, checked)
		sw.Send(events.Click)
		assert.False(t, checked)
		sw.Send(events.Click)
		assert.True(t, checked)
	})
}
