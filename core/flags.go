// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"cogentcore.org/core/enums"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/tree"
)

// StateIs returns whether the widget has the given [states.States] flag set.
func (wb *WidgetBase) StateIs(state states.States) bool {
	return wb.Styles.State.HasFlag(state)
}

// AbilityIs returns whether the widget has the given [abilities.Abilities] flag set.
func (wb *WidgetBase) AbilityIs(able abilities.Abilities) bool {
	return wb.Styles.Abilities.HasFlag(able)
}

// SetState sets the given [states.State] flags to the given value.
func (wb *WidgetBase) SetState(on bool, state ...states.States) *WidgetBase {
	bfs := make([]enums.BitFlag, len(state))
	for i, st := range state {
		bfs[i] = st
	}
	wb.Styles.State.SetFlag(on, bfs...)
	return wb
}

// SetAbilities sets the given [abilities.Abilities] flags to the given value.
func (wb *WidgetBase) SetAbilities(on bool, able ...abilities.Abilities) *WidgetBase {
	bfs := make([]enums.BitFlag, len(able))
	for i, st := range able {
		bfs[i] = st
	}
	wb.Styles.Abilities.SetFlag(on, bfs...)
	return wb
}

// SetSelected sets the [states.Selected] flag to given value for the entire Widget
// and calls [WidgetBase.Restyle] to apply any resultant style changes.
func (wb *WidgetBase) SetSelected(sel bool) *WidgetBase {
	wb.SetState(sel, states.Selected)
	wb.Restyle()
	return wb
}

// CanFocus returns whether this node can receive keyboard focus.
func (wb *WidgetBase) CanFocus() bool {
	return wb.Styles.Abilities.HasFlag(abilities.Focusable)
}

// SetEnabled sets the [states.Disabled] flag to the opposite of the given value.
func (wb *WidgetBase) SetEnabled(enabled bool) *WidgetBase {
	return wb.SetState(!enabled, states.Disabled)
}

// IsDisabled returns whether this node is flagged as [Disabled].
// If so, behave and style appropriately.
func (wb *WidgetBase) IsDisabled() bool {
	return wb.StateIs(states.Disabled)
}

// IsReadOnly returns whether this node is flagged as [ReadOnly] or [Disabled].
// If so, behave appropriately. Styling is based on each state separately,
// but behaviors are often the same for both of these states.
func (wb *WidgetBase) IsReadOnly() bool {
	return wb.StateIs(states.ReadOnly) || wb.StateIs(states.Disabled)
}

// SetReadOnly sets the [states.ReadOnly] flag to the given value.
func (wb *WidgetBase) SetReadOnly(ro bool) *WidgetBase {
	return wb.SetState(ro, states.ReadOnly)
}

// HasStateWithin returns whether the current node or any
// of its children have the given state flag.
func (wb *WidgetBase) HasStateWithin(state states.States) bool {
	got := false
	wb.WidgetWalkDown(func(wi Widget, wb *WidgetBase) bool {
		if wb.StateIs(state) {
			got = true
			return tree.Break
		}
		return tree.Continue
	})
	return got
}
