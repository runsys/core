// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"errors"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/stretchr/testify/assert"
)

func TestTextField(t *testing.T) {
	b := NewBody()
	NewTextField(b)
	b.AssertRender(t, "text-field/basic")
}

func TestTextFieldPlaceholder(t *testing.T) {
	b := NewBody()
	NewTextField(b).SetPlaceholder("Jane Doe")
	b.AssertRender(t, "text-field/placeholder")
}

func TestTextFieldText(t *testing.T) {
	b := NewBody()
	NewTextField(b).SetText("Hello, world!")
	b.AssertRender(t, "text-field/text")
}

func TestTextFieldOverflow(t *testing.T) {
	b := NewBody()
	NewTextField(b).SetText("This is a long sentence that demonstrates how text field content can overflow onto multiple lines")
	b.AssertRender(t, "text-field/overflow")
}

func TestTextFieldOutlined(t *testing.T) {
	b := NewBody()
	NewTextField(b).SetType(TextFieldOutlined)
	b.AssertRender(t, "text-field/outlined")
}

func TestTextFieldPassword(t *testing.T) {
	b := NewBody()
	NewTextField(b).SetTypePassword().SetText("my password")
	b.AssertRender(t, "text-field/password")
}

func TestTextFieldPasswordClick(t *testing.T) {
	b := NewBody()
	tf := NewTextField(b).SetTypePassword().SetText("my password")
	b.AssertRender(t, "text-field/password-click", func() {
		tf.TrailingIconButton().Send(events.Click)
	})
}

func TestTextFieldClear(t *testing.T) {
	b := NewBody()
	NewTextField(b).AddClearButton().SetText("Hello, world!")
	b.AssertRender(t, "text-field/clear")
}

func TestTextFieldClearClick(t *testing.T) {
	b := NewBody()
	tf := NewTextField(b).AddClearButton().SetText("Hello, world!")
	b.AssertRender(t, "text-field/clear-click", func() {
		tf.TrailingIconButton().Send(events.Click)
		assert.Equal(t, "", tf.Text())
	})
}

func TestTextFieldIcons(t *testing.T) {
	b := NewBody()
	clicked := false
	tf := NewTextField(b).SetTrailingIcon(icons.Euro).SetLeadingIcon(icons.OpenInNew, func(e events.Event) {
		clicked = true
	})
	b.AssertRender(t, "text-field/icons", func() {
		tf.LeadingIconButton().Send(events.Click)
		assert.True(t, clicked)
	})
}

func TestTextFieldValidatorValid(t *testing.T) {
	b := NewBody()
	tf := NewTextField(b).SetText("my secure password")
	tf.SetValidator(func() error {
		if len(tf.Text()) < 12 {
			return errors.New("password must be at least 12 characters")
		}
		return nil
	})
	b.AssertRender(t, "text-field/validator-valid", func() {
		tf.SendChange() // trigger validation
	})
}

func TestTextFieldValidatorInvalid(t *testing.T) {
	b := NewBody()
	b.Styler(func(s *styles.Style) {
		s.Min.Set(units.Em(15), units.Em(10))
	})
	tf := NewTextField(b).SetText("my password")
	tf.SetValidator(func() error {
		if len(tf.Text()) < 12 {
			return errors.New("password must be at least 12 characters")
		}
		return nil
	})
	b.AssertRenderScreen(t, "text-field/validator-invalid", func() {
		tf.SendChange() // trigger validation
	})
}

func TestTextFieldChange(t *testing.T) {
	b := NewBody()
	tf := NewTextField(b)
	text := ""
	tf.OnChange(func(e events.Event) {
		text = tf.Text()
	})
	b.AssertRender(t, "text-field/change", func() {
		tf.HandleEvent(events.NewKey(events.KeyChord, 'G', 0, 0))
		tf.HandleEvent(events.NewKey(events.KeyChord, 'o', 0, 0))
		assert.Equal(t, "", text)
		tf.HandleEvent(events.NewKey(events.KeyChord, 0, key.CodeReturnEnter, 0))
		assert.Equal(t, "Go", text)
	})
}

func TestTextFieldInput(t *testing.T) {
	b := NewBody()
	tf := NewTextField(b)
	n := 0
	text := ""
	tf.OnInput(func(e events.Event) {
		n++
		text = tf.Text()
	})
	b.AssertRender(t, "text-field/input", func() {
		tf.HandleEvent(events.NewKey(events.KeyChord, 'G', 0, 0))
		assert.Equal(t, 1, n)
		assert.Equal(t, "G", text)
		tf.HandleEvent(events.NewKey(events.KeyChord, 'o', 0, 0))
		assert.Equal(t, 2, n)
		assert.Equal(t, "Go", text)
		tf.HandleEvent(events.NewKey(events.KeyChord, 0, key.CodeReturnEnter, 0))
		assert.Equal(t, 2, n)
		assert.Equal(t, "Go", text)
	})
}
