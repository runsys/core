// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package texteditor

import (
	"strings"
	"unicode"

	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/parse/lexer"
	"cogentcore.org/core/parse/token"
	"cogentcore.org/core/texteditor/text"
)

///////////////////////////////////////////////////////////////////////////////
//    Complete and Spell

// offerComplete pops up a menu of possible completions
func (ed *Editor) offerComplete() {
	if ed.Buffer.Complete == nil || ed.ISearch.On || ed.QReplace.On || ed.IsDisabled() {
		return
	}
	ed.Buffer.Complete.Cancel()
	if !ed.Buffer.Options.Completion {
		return
	}
	if ed.Buffer.InComment(ed.CursorPos) || ed.Buffer.InLitString(ed.CursorPos) {
		return
	}

	ed.Buffer.Complete.SrcLn = ed.CursorPos.Ln
	ed.Buffer.Complete.SrcCh = ed.CursorPos.Ch
	st := lexer.Pos{ed.CursorPos.Ln, 0}
	en := lexer.Pos{ed.CursorPos.Ln, ed.CursorPos.Ch}
	tbe := ed.Buffer.Region(st, en)
	var s string
	if tbe != nil {
		s = string(tbe.ToBytes())
		s = strings.TrimLeft(s, " \t") // trim ' ' and '\t'
	}

	//	count := ed.Buf.ByteOffs[ed.CursorPos.Ln] + ed.CursorPos.Ch
	cpos := ed.charStartPos(ed.CursorPos).ToPoint() // physical location
	cpos.X += 5
	cpos.Y += 10
	// ed.Buffer.setByteOffs() // make sure the pos offset is updated!!
	// todo: why? for above
	ed.Buffer.currentEditor = ed
	ed.Buffer.Complete.SrcLn = ed.CursorPos.Ln
	ed.Buffer.Complete.SrcCh = ed.CursorPos.Ch
	ed.Buffer.Complete.Show(ed, cpos, s)
}

// CancelComplete cancels any pending completion.
// Call this when new events have moved beyond any prior completion scenario.
func (ed *Editor) CancelComplete() {
	if ed.Buffer == nil {
		return
	}
	if ed.Buffer.Complete == nil {
		return
	}
	if ed.Buffer.Complete.Cancel() {
		ed.Buffer.currentEditor = nil
	}
}

// Lookup attempts to lookup symbol at current location, popping up a window
// if something is found.
func (ed *Editor) Lookup() { //types:add
	if ed.Buffer.Complete == nil || ed.ISearch.On || ed.QReplace.On || ed.IsDisabled() {
		return
	}

	var ln int
	var ch int
	if ed.HasSelection() {
		ln = ed.SelectRegion.Start.Ln
		if ed.SelectRegion.End.Ln != ln {
			return // no multiline selections for lookup
		}
		ch = ed.SelectRegion.End.Ch
	} else {
		ln = ed.CursorPos.Ln
		if ed.isWordEnd(ed.CursorPos) {
			ch = ed.CursorPos.Ch
		} else {
			ch = ed.wordAt().End.Ch
		}
	}
	ed.Buffer.Complete.SrcLn = ln
	ed.Buffer.Complete.SrcCh = ch
	st := lexer.Pos{ed.CursorPos.Ln, 0}
	en := lexer.Pos{ed.CursorPos.Ln, ch}

	tbe := ed.Buffer.Region(st, en)
	var s string
	if tbe != nil {
		s = string(tbe.ToBytes())
		s = strings.TrimLeft(s, " \t") // trim ' ' and '\t'
	}

	//	count := ed.Buf.ByteOffs[ed.CursorPos.Ln] + ed.CursorPos.Ch
	cpos := ed.charStartPos(ed.CursorPos).ToPoint() // physical location
	cpos.X += 5
	cpos.Y += 10
	// ed.Buffer.setByteOffs() // make sure the pos offset is updated!!
	// todo: why?
	ed.Buffer.currentEditor = ed
	ed.Buffer.Complete.Lookup(s, ed.CursorPos.Ln, ed.CursorPos.Ch, ed.Scene, cpos)
}

// iSpellKeyInput locates the word to spell check based on cursor position and
// the key input, then passes the text region to SpellCheck
func (ed *Editor) iSpellKeyInput(kt events.Event) {
	if !ed.Buffer.isSpellEnabled(ed.CursorPos) {
		return
	}

	isDoc := ed.Buffer.Info.Cat == fileinfo.Doc
	tp := ed.CursorPos

	kf := keymap.Of(kt.KeyChord())
	switch kf {
	case keymap.MoveUp:
		if isDoc {
			ed.Buffer.spellCheckLineTag(tp.Ln)
		}
	case keymap.MoveDown:
		if isDoc {
			ed.Buffer.spellCheckLineTag(tp.Ln)
		}
	case keymap.MoveRight:
		if ed.isWordEnd(tp) {
			reg := ed.wordBefore(tp)
			ed.spellCheck(reg)
			break
		}
		if tp.Ch == 0 { // end of line
			tp.Ln--
			if isDoc {
				ed.Buffer.spellCheckLineTag(tp.Ln) // redo prior line
			}
			tp.Ch = ed.Buffer.LineLen(tp.Ln)
			reg := ed.wordBefore(tp)
			ed.spellCheck(reg)
			break
		}
		txt := ed.Buffer.Line(tp.Ln)
		var r rune
		atend := false
		if tp.Ch >= len(txt) {
			atend = true
			tp.Ch++
		} else {
			r = txt[tp.Ch]
		}
		if atend || core.IsWordBreak(r, rune(-1)) {
			tp.Ch-- // we are one past the end of word
			reg := ed.wordBefore(tp)
			ed.spellCheck(reg)
		}
	case keymap.Enter:
		tp.Ln--
		if isDoc {
			ed.Buffer.spellCheckLineTag(tp.Ln) // redo prior line
		}
		tp.Ch = ed.Buffer.LineLen(tp.Ln)
		reg := ed.wordBefore(tp)
		ed.spellCheck(reg)
	case keymap.FocusNext:
		tp.Ch-- // we are one past the end of word
		reg := ed.wordBefore(tp)
		ed.spellCheck(reg)
	case keymap.Backspace, keymap.Delete:
		if ed.isWordMiddle(ed.CursorPos) {
			reg := ed.wordAt()
			ed.spellCheck(ed.Buffer.Region(reg.Start, reg.End))
		} else {
			reg := ed.wordBefore(tp)
			ed.spellCheck(reg)
		}
	case keymap.None:
		if unicode.IsSpace(kt.KeyRune()) || unicode.IsPunct(kt.KeyRune()) && kt.KeyRune() != '\'' { // contractions!
			tp.Ch-- // we are one past the end of word
			reg := ed.wordBefore(tp)
			ed.spellCheck(reg)
		} else {
			if ed.isWordMiddle(ed.CursorPos) {
				reg := ed.wordAt()
				ed.spellCheck(ed.Buffer.Region(reg.Start, reg.End))
			}
		}
	}
}

// spellCheck offers spelling corrections if we are at a word break or other word termination
// and the word before the break is unknown -- returns true if misspelled word found
func (ed *Editor) spellCheck(reg *text.Edit) bool {
	if ed.Buffer.spell == nil {
		return false
	}
	wb := string(reg.ToBytes())
	lwb := lexer.FirstWordApostrophe(wb) // only lookup words
	if len(lwb) <= 2 {
		return false
	}
	widx := strings.Index(wb, lwb) // adjust region for actual part looking up
	ld := len(wb) - len(lwb)
	reg.Reg.Start.Ch += widx
	reg.Reg.End.Ch += widx - ld

	sugs, knwn := ed.Buffer.spell.checkWord(lwb)
	if knwn {
		ed.Buffer.RemoveTag(reg.Reg.Start, token.TextSpellErr)
		return false
	}
	// fmt.Printf("spell err: %s\n", wb)
	ed.Buffer.spell.setWord(wb, sugs, reg.Reg.Start.Ln, reg.Reg.Start.Ch)
	ed.Buffer.RemoveTag(reg.Reg.Start, token.TextSpellErr)
	ed.Buffer.AddTagEdit(reg, token.TextSpellErr)
	return true
}

// offerCorrect pops up a menu of possible spelling corrections for word at
// current CursorPos. If no misspelling there or not in spellcorrect mode
// returns false
func (ed *Editor) offerCorrect() bool {
	if ed.Buffer.spell == nil || ed.ISearch.On || ed.QReplace.On || ed.IsDisabled() {
		return false
	}
	sel := ed.SelectRegion
	if !ed.selectWord() {
		ed.SelectRegion = sel
		return false
	}
	tbe := ed.Selection()
	if tbe == nil {
		ed.SelectRegion = sel
		return false
	}
	ed.SelectRegion = sel
	wb := string(tbe.ToBytes())
	wbn := strings.TrimLeft(wb, " \t")
	if len(wb) != len(wbn) {
		return false // SelectWord captures leading whitespace - don't offer if there is leading whitespace
	}
	sugs, knwn := ed.Buffer.spell.checkWord(wb)
	if knwn && !ed.Buffer.spell.isLastLearned(wb) {
		return false
	}
	ed.Buffer.spell.setWord(wb, sugs, tbe.Reg.Start.Ln, tbe.Reg.Start.Ch)

	cpos := ed.charStartPos(ed.CursorPos).ToPoint() // physical location
	cpos.X += 5
	cpos.Y += 10
	ed.Buffer.currentEditor = ed
	ed.Buffer.spell.show(wb, ed.Scene, cpos)
	return true
}

// cancelCorrect cancels any pending spell correction.
// Call this when new events have moved beyond any prior correction scenario.
func (ed *Editor) cancelCorrect() {
	if ed.Buffer.spell == nil || ed.ISearch.On || ed.QReplace.On {
		return
	}
	if !ed.Buffer.Options.SpellCorrect {
		return
	}
	ed.Buffer.currentEditor = nil
	ed.Buffer.spell.cancel()
}
