// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"image"
	"sync"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/parse/complete"
)

//////////////////////////////////////////////////////////////////////////////
// Complete

// Complete holds the current completion data and functions to call for building
// the list of possible completions and for editing text after a completion is selected.
// It also holds the [PopupStage] associated with it.
type Complete struct { //types:add -setters
	// function to get the list of possible completions
	MatchFunc complete.MatchFunc

	// function to get the text to show for lookup
	LookupFunc complete.LookupFunc

	// function to edit text using the selected completion
	EditFunc complete.EditFunc

	// the object that implements complete.Func
	Context any

	// line number in source that completion is operating on, if relevant
	SrcLn int

	// character position in source that completion is operating on
	SrcCh int

	// the list of potential completions
	Completions complete.Completions

	// current completion seed
	Seed string

	// the user's completion selection
	Completion string

	// the event listeners for the completer (it sends Select events)
	Listeners events.Listeners `set:"-" display:"-"`

	// Stage is the [PopupStage] associated with the [Complete]
	Stage *Stage

	DelayTimer *time.Timer `set:"-"`
	DelayMu    sync.Mutex  `set:"-"`
	ShowMu     sync.Mutex  `set:"-"`
}

// CompleteSignals are signals that are sent by Complete
type CompleteSignals int32 //enums:enum -trim-prefix Complete

const (
	// CompleteSelect means the user chose one of the possible completions
	CompleteSelect CompleteSignals = iota

	// CompleteExtend means user has requested that the seed extend if all
	// completions have a common prefix longer than current seed
	CompleteExtend
)

// NewComplete returns a new [Complete] object. It does not show it; see [Complete.Show].
func NewComplete() *Complete {
	return &Complete{}
}

// IsAboutToShow returns true if the DelayTimer is started for
// preparing to show a completion.  note: don't really need to lock
func (c *Complete) IsAboutToShow() bool {
	c.DelayMu.Lock()
	defer c.DelayMu.Unlock()
	return c.DelayTimer != nil
}

// Show is the main call for listing completions.
// Has a builtin delay timer so completions are only shown after
// a delay, which resets every time it is called.
// After delay, Calls ShowNow, which calls MatchFunc
// to get a list of completions and builds the completion popup menu
func (c *Complete) Show(ctx Widget, pos image.Point, text string) {
	if c.MatchFunc == nil {
		return
	}
	wait := SystemSettings.CompleteWaitDuration
	if c.Stage != nil {
		c.Cancel()
	}
	if wait == 0 {
		c.ShowNow(ctx, pos, text)
		return
	}
	c.DelayMu.Lock()
	if c.DelayTimer != nil {
		c.DelayTimer.Stop()
	}
	c.DelayTimer = time.AfterFunc(wait,
		func() {
			c.ShowNowAsync(ctx, pos, text)
			c.DelayMu.Lock()
			c.DelayTimer = nil
			c.DelayMu.Unlock()
		})
	c.DelayMu.Unlock()
}

// ShowNow actually calls MatchFunc to get a list of completions and builds the
// completion popup menu.  This is the sync version called from
func (c *Complete) ShowNow(ctx Widget, pos image.Point, text string) {
	if c.Stage != nil {
		c.Cancel()
	}
	c.ShowMu.Lock()
	defer c.ShowMu.Unlock()
	if c.ShowNowImpl(ctx, pos, text) {
		c.Stage.RunPopup()
	}
}

// ShowNowAsync actually calls MatchFunc to get a list of completions and builds the
// completion popup menu.  This is the Async version for delayed AfterFunc call.
func (c *Complete) ShowNowAsync(ctx Widget, pos image.Point, text string) {
	if c.Stage != nil {
		c.CancelAsync()
	}
	c.ShowMu.Lock()
	defer c.ShowMu.Unlock()
	if c.ShowNowImpl(ctx, pos, text) {
		c.Stage.RunPopupAsync()
	}
}

// ShowNowImpl is the implementation of ShowNow, presenting completions.
// Returns false if nothing to show.
func (c *Complete) ShowNowImpl(ctx Widget, pos image.Point, text string) bool {
	md := c.MatchFunc(c.Context, text, c.SrcLn, c.SrcCh)
	c.Completions = md.Matches
	c.Seed = md.Seed
	if len(c.Completions) == 0 {
		return false
	}
	if len(c.Completions) > SystemSettings.CompleteMaxItems {
		c.Completions = c.Completions[0:SystemSettings.CompleteMaxItems]
	}

	sc := NewScene(ctx.AsTree().Name + "-complete")
	MenuSceneConfigStyles(sc)
	c.Stage = NewPopupStage(CompleterStage, sc, ctx).SetPos(pos)
	// we forward our key events to the context object
	// so that you can keep typing while in a completer
	// sc.OnKeyChord(ctx.HandleEvent)

	for i := 0; i < len(c.Completions); i++ {
		cmp := &c.Completions[i]
		text := cmp.Text
		if cmp.Label != "" {
			text = cmp.Label
		}
		icon := cmp.Icon
		mi := NewButton(sc).SetText(text).SetIcon(icons.Icon(icon))
		mi.SetTooltip(cmp.Desc)
		mi.OnClick(func(e events.Event) {
			c.Complete(cmp.Text)
		})
		mi.OnKeyChord(func(e events.Event) {
			kf := keymap.Of(e.KeyChord())
			if e.KeyRune() == ' ' {
				ctx.AsWidget().HandleEvent(e) // bypass button handler!
			}
			if kf == keymap.Enter {
				e.SetHandled()
				c.Complete(cmp.Text)
			}
		})
		if i == 0 {
			sc.Events.SetStartFocus(mi)
		}
	}
	return true
}

// Cancel cancels any existing *or* pending completion.
// Call when new events nullify prior completions.
// Returns true if canceled.
func (c *Complete) Cancel() bool {
	if c.Stage == nil {
		return false
	}
	st := c.Stage
	c.Stage = nil
	st.ClosePopup()
	return true
}

// CancelAsync cancels any existing *or* pending completion,
// inside a delayed callback function (Async)
// Call when new events nullify prior completions.
// Returns true if canceled.
func (c *Complete) CancelAsync() bool {
	if c.Stage == nil {
		return false
	}
	st := c.Stage
	c.Stage = nil
	st.ClosePopupAsync()
	return true
}

// Abort aborts *only* pending completions, but does not close existing window.
// Returns true if aborted.
func (c *Complete) Abort() bool {
	c.DelayMu.Lock()
	// c.Sc = nil
	if c.DelayTimer != nil {
		c.DelayTimer.Stop()
		c.DelayTimer = nil
		c.DelayMu.Unlock()
		return true
	}
	c.DelayMu.Unlock()
	return false
}

// Lookup is the main call for doing lookups
func (c *Complete) Lookup(text string, posLine, posChar int, sc *Scene, pt image.Point) {
	if c.LookupFunc == nil || sc == nil {
		return
	}
	// c.Sc = nil
	c.LookupFunc(c.Context, text, posLine, posChar) // this processes result directly
}

// Complete sends Select event to listeners, indicating that the user has made a
// selection from the list of possible completions.
// This is called inside the main event loop.
func (c *Complete) Complete(s string) {
	c.Cancel()
	c.Completion = s
	c.Listeners.Call(&events.Base{Typ: events.Select})
}

// OnSelect registers given listener function for Select events on Value.
// This is the primary notification event for all Complete elements.
func (c *Complete) OnSelect(fun func(e events.Event)) {
	c.On(events.Select, fun)
}

// On adds an event listener function for the given event type
func (c *Complete) On(etype events.Types, fun func(e events.Event)) {
	c.Listeners.Add(etype, fun)
}

func (c *Complete) GetCompletion(s string) complete.Completion {
	for _, cc := range c.Completions {
		if s == cc.Text {
			return cc
		}
	}
	return complete.Completion{}
}

// CompleteEditText is a chance to modify the completion selection before it is inserted
func CompleteEditText(text string, cp int, completion string, seed string) (ed complete.Edit) {
	ed.NewText = completion
	return ed
}
