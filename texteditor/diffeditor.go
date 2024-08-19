// Copyright (c) 2020, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package texteditor

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/fileinfo/mimedata"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/stringsx"
	"cogentcore.org/core/base/vcs"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/parse/lexer"
	"cogentcore.org/core/parse/token"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/texteditor/textbuf"
	"cogentcore.org/core/tree"
)

// DiffFiles shows the diffs between this file as the A file, and other file as B file,
// in a DiffEditorDialog
func DiffFiles(ctx core.Widget, afile, bfile string) (*DiffEditor, error) {
	ab, err := os.ReadFile(afile)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	bb, err := os.ReadFile(bfile)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	astr := stringsx.SplitLines(string(ab))
	bstr := stringsx.SplitLines(string(bb))
	dlg := DiffEditorDialog(ctx, "Diff File View", astr, bstr, afile, bfile, "", "")
	return dlg, nil
}

// DiffEditorDialogFromRevs opens a dialog for displaying diff between file
// at two different revisions from given repository
// if empty, defaults to: A = current HEAD, B = current WC file.
// -1, -2 etc also work as universal ways of specifying prior revisions.
func DiffEditorDialogFromRevs(ctx core.Widget, repo vcs.Repo, file string, fbuf *Buffer, rev_a, rev_b string) (*DiffEditor, error) {
	var astr, bstr []string
	if rev_b == "" { // default to current file
		if fbuf != nil {
			bstr = fbuf.Strings(false)
		} else {
			fb, err := textbuf.FileBytes(file)
			if err != nil {
				core.ErrorDialog(ctx, err)
				return nil, err
			}
			bstr = textbuf.BytesToLineStrings(fb, false) // don't add new lines
		}
	} else {
		fb, err := repo.FileContents(file, rev_b)
		if err != nil {
			core.ErrorDialog(ctx, err)
			return nil, err
		}
		bstr = textbuf.BytesToLineStrings(fb, false) // don't add new lines
	}
	fb, err := repo.FileContents(file, rev_a)
	if err != nil {
		core.ErrorDialog(ctx, err)
		return nil, err
	}
	astr = textbuf.BytesToLineStrings(fb, false) // don't add new lines
	if rev_a == "" {
		rev_a = "HEAD"
	}
	return DiffEditorDialog(ctx, "DiffVcs: "+fsx.DirAndFile(file), astr, bstr, file, file, rev_a, rev_b), nil
}

// DiffEditorDialog opens a dialog for displaying diff between two files as line-strings
func DiffEditorDialog(ctx core.Widget, title string, astr, bstr []string, afile, bfile, arev, brev string) *DiffEditor {
	d := core.NewBody("Diff editor")
	d.SetTitle(title)

	dv := NewDiffEditor(d)
	dv.SetFileA(afile).SetFileB(bfile).SetRevA(arev).SetRevB(brev)
	dv.DiffStrings(astr, bstr)
	d.AddAppBar(dv.MakeToolbar)
	d.NewWindow().SetContext(ctx).SetNewWindow(true).Run()
	return dv
}

// TextDialog opens a dialog for displaying text string
func TextDialog(ctx core.Widget, title, text string) *Editor {
	d := core.NewBody().AddTitle(title)
	buf := NewBuffer()
	ed := NewEditor(d).SetBuffer(buf)
	buf.SetText([]byte(text))
	d.AddBottomBar(func(parent core.Widget) {
		core.NewButton(parent).SetText("Copy to clipboard").SetIcon(icons.ContentCopy).
			OnClick(func(e events.Event) {
				d.Clipboard().Write(mimedata.NewText(text))
			})
		d.AddOK(parent)
	})
	d.RunWindowDialog(ctx)
	return ed
}

///////////////////////////////////////////////////////////////////
// DiffEditor

// DiffEditor presents two side-by-side [Editor]s showing the differences
// between two files (represented as lines of strings).
type DiffEditor struct {
	core.Frame

	// first file name being compared
	FileA string

	// second file name being compared
	FileB string

	// revision for first file, if relevant
	RevA string

	// revision for second file, if relevant
	RevB string

	// textbuf for A showing the aligned edit view
	BufA *Buffer `json:"-" xml:"-" set:"-"`

	// textbuf for B showing the aligned edit view
	BufB *Buffer `json:"-" xml:"-" set:"-"`

	// aligned diffs records diff for aligned lines
	AlignD textbuf.Diffs `json:"-" xml:"-" set:"-"`

	// Diffs applied
	Diffs textbuf.DiffSelected

	inInputEvent bool
}

func (dv *DiffEditor) Init() {
	dv.Frame.Init()
	dv.BufA = NewBuffer().SetFilename(dv.FileA)
	dv.BufB = NewBuffer().SetFilename(dv.FileB)
	dv.BufA.Options.LineNumbers = true
	dv.BufA.Stat() // update markup
	dv.BufB.Options.LineNumbers = true
	dv.BufB.Stat() // update markup

	dv.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
	})

	f := func(name string, buf *Buffer) {
		tree.AddChildAt(dv, name, func(w *DiffTextEditor) {
			w.SetBuffer(buf)
			w.SetReadOnly(true)
			w.Styler(func(s *styles.Style) {
				s.Min.X.Ch(80)
				s.Min.Y.Em(40)
			})
			w.On(events.Scroll, func(e events.Event) {
				dv.SyncViews(events.Scroll, e, name)
			})
			w.On(events.Input, func(e events.Event) {
				dv.SyncViews(events.Input, e, name)
			})
		})
	}
	f("text-a", dv.BufA)
	f("text-b", dv.BufB)
}

// SyncViews synchronizes the text view scrolling and cursor positions
func (dv *DiffEditor) SyncViews(typ events.Types, e events.Event, name string) {
	tva, tvb := dv.TextEditors()
	me, other := tva, tvb
	if name == "text-b" {
		me, other = tvb, tva
	}
	switch typ {
	case events.Scroll:
		other.Geom.Scroll.Y = me.Geom.Scroll.Y
		other.ScrollUpdateFromGeom(math32.Y)
	case events.Input:
		if dv.inInputEvent {
			return
		}
		dv.inInputEvent = true
		other.SetCursorShow(me.CursorPos)
		dv.inInputEvent = false
	}
}

// NextDiff moves to next diff region
func (dv *DiffEditor) NextDiff(ab int) bool {
	tva, tvb := dv.TextEditors()
	tv := tva
	if ab == 1 {
		tv = tvb
	}
	nd := len(dv.AlignD)
	curLn := tv.CursorPos.Ln
	di, df := dv.AlignD.DiffForLine(curLn)
	if di < 0 {
		return false
	}
	for {
		di++
		if di >= nd {
			return false
		}
		df = dv.AlignD[di]
		if df.Tag != 'e' {
			break
		}
	}
	tva.SetCursorTarget(lexer.Pos{Ln: df.I1})
	return true
}

// PrevDiff moves to previous diff region
func (dv *DiffEditor) PrevDiff(ab int) bool {
	tva, tvb := dv.TextEditors()
	tv := tva
	if ab == 1 {
		tv = tvb
	}
	curLn := tv.CursorPos.Ln
	di, df := dv.AlignD.DiffForLine(curLn)
	if di < 0 {
		return false
	}
	for {
		di--
		if di < 0 {
			return false
		}
		df = dv.AlignD[di]
		if df.Tag != 'e' {
			break
		}
	}
	tva.SetCursorTarget(lexer.Pos{Ln: df.I1})
	return true
}

// SaveAs saves A or B edits into given file.
// It checks for an existing file, prompts to overwrite or not.
func (dv *DiffEditor) SaveAs(ab bool, filename core.Filename) {
	if !errors.Log1(fsx.FileExists(string(filename))) {
		dv.SaveFile(ab, filename)
	} else {
		d := core.NewBody().AddTitle("File Exists, Overwrite?").
			AddText(fmt.Sprintf("File already exists, overwrite?  File: %v", filename))
		d.AddBottomBar(func(parent core.Widget) {
			d.AddCancel(parent)
			d.AddOK(parent).OnClick(func(e events.Event) {
				dv.SaveFile(ab, filename)
			})
		})
		d.RunDialog(dv)
	}
}

// SaveFile writes A or B edits to file, with no prompting, etc
func (dv *DiffEditor) SaveFile(ab bool, filename core.Filename) error {
	var txt string
	if ab {
		txt = strings.Join(dv.Diffs.B.Edit, "\n")
	} else {
		txt = strings.Join(dv.Diffs.A.Edit, "\n")
	}
	err := os.WriteFile(string(filename), []byte(txt), 0644)
	if err != nil {
		core.ErrorSnackbar(dv, err)
		slog.Error(err.Error())
	}
	return err
}

// SaveFileA saves the current state of file A to given filename
func (dv *DiffEditor) SaveFileA(fname core.Filename) { //types:add
	dv.SaveAs(false, fname)
	// dv.UpdateToolbar()
}

// SaveFileB saves the current state of file B to given filename
func (dv *DiffEditor) SaveFileB(fname core.Filename) { //types:add
	dv.SaveAs(true, fname)
	// dv.UpdateToolbar()
}

// DiffStrings computes differences between two lines-of-strings and displays in
// DiffEditor.
func (dv *DiffEditor) DiffStrings(astr, bstr []string) {
	dv.Diffs.SetStringLines(astr, bstr)

	dv.BufA.LineColors = nil
	dv.BufB.LineColors = nil
	del := colors.Scheme.Error.Base
	ins := colors.Scheme.Success.Base
	chg := colors.Scheme.Primary.Base

	nd := len(dv.Diffs.Diffs)
	dv.AlignD = make(textbuf.Diffs, nd)
	var ab, bb [][]byte
	absln := 0
	bspc := []byte(" ")
	for i, df := range dv.Diffs.Diffs {
		switch df.Tag {
		case 'r':
			di := df.I2 - df.I1
			dj := df.J2 - df.J1
			mx := max(di, dj)
			ad := df
			ad.I1 = absln
			ad.I2 = absln + di
			ad.J1 = absln
			ad.J2 = absln + dj
			dv.AlignD[i] = ad
			for i := 0; i < mx; i++ {
				dv.BufA.SetLineColor(absln+i, chg)
				dv.BufB.SetLineColor(absln+i, chg)
				blen := 0
				alen := 0
				if i < di {
					aln := []byte(astr[df.I1+i])
					alen = len(aln)
					ab = append(ab, aln)
				}
				if i < dj {
					bln := []byte(bstr[df.J1+i])
					blen = len(bln)
					bb = append(bb, bln)
				} else {
					bb = append(bb, bytes.Repeat(bspc, alen))
				}
				if i >= di {
					ab = append(ab, bytes.Repeat(bspc, blen))
				}
			}
			absln += mx
		case 'd':
			di := df.I2 - df.I1
			ad := df
			ad.I1 = absln
			ad.I2 = absln + di
			ad.J1 = absln
			ad.J2 = absln + di
			dv.AlignD[i] = ad
			for i := 0; i < di; i++ {
				dv.BufA.SetLineColor(absln+i, ins)
				dv.BufB.SetLineColor(absln+i, del)
				aln := []byte(astr[df.I1+i])
				alen := len(aln)
				ab = append(ab, aln)
				bb = append(bb, bytes.Repeat(bspc, alen))
			}
			absln += di
		case 'i':
			dj := df.J2 - df.J1
			ad := df
			ad.I1 = absln
			ad.I2 = absln + dj
			ad.J1 = absln
			ad.J2 = absln + dj
			dv.AlignD[i] = ad
			for i := 0; i < dj; i++ {
				dv.BufA.SetLineColor(absln+i, del)
				dv.BufB.SetLineColor(absln+i, ins)
				bln := []byte(bstr[df.J1+i])
				blen := len(bln)
				bb = append(bb, bln)
				ab = append(ab, bytes.Repeat(bspc, blen))
			}
			absln += dj
		case 'e':
			di := df.I2 - df.I1
			ad := df
			ad.I1 = absln
			ad.I2 = absln + di
			ad.J1 = absln
			ad.J2 = absln + di
			dv.AlignD[i] = ad
			for i := 0; i < di; i++ {
				ab = append(ab, []byte(astr[df.I1+i]))
				bb = append(bb, []byte(bstr[df.J1+i]))
			}
			absln += di
		}
	}
	dv.BufA.SetTextLines(ab, false) // don't copy
	dv.BufB.SetTextLines(bb, false) // don't copy
	dv.TagWordDiffs()
	dv.BufA.ReMarkup()
	dv.BufB.ReMarkup()
}

// TagWordDiffs goes through replace diffs and tags differences at the
// word level between the two regions.
func (dv *DiffEditor) TagWordDiffs() {
	for _, df := range dv.AlignD {
		if df.Tag != 'r' {
			continue
		}
		di := df.I2 - df.I1
		dj := df.J2 - df.J1
		mx := max(di, dj)
		stln := df.I1
		for i := 0; i < mx; i++ {
			ln := stln + i
			ra := dv.BufA.Lines[ln]
			rb := dv.BufB.Lines[ln]
			lna := lexer.RuneFields(ra)
			lnb := lexer.RuneFields(rb)
			fla := lna.RuneStrings(ra)
			flb := lnb.RuneStrings(rb)
			nab := max(len(fla), len(flb))
			ldif := textbuf.DiffLines(fla, flb)
			ndif := len(ldif)
			if nab > 25 && ndif > nab/2 { // more than half of big diff -- skip
				continue
			}
			for _, ld := range ldif {
				switch ld.Tag {
				case 'r':
					sla := lna[ld.I1]
					ela := lna[ld.I2-1]
					dv.BufA.AddTag(ln, sla.St, ela.Ed, token.TextStyleError)
					slb := lnb[ld.J1]
					elb := lnb[ld.J2-1]
					dv.BufB.AddTag(ln, slb.St, elb.Ed, token.TextStyleError)
				case 'd':
					sla := lna[ld.I1]
					ela := lna[ld.I2-1]
					dv.BufA.AddTag(ln, sla.St, ela.Ed, token.TextStyleDeleted)
				case 'i':
					slb := lnb[ld.J1]
					elb := lnb[ld.J2-1]
					dv.BufB.AddTag(ln, slb.St, elb.Ed, token.TextStyleDeleted)
				}
			}
		}
	}
}

// ApplyDiff applies change from the other buffer to the buffer for given file
// name, from diff that includes given line.
func (dv *DiffEditor) ApplyDiff(ab int, line int) bool {
	tva, tvb := dv.TextEditors()
	tv := tva
	if ab == 1 {
		tv = tvb
	}
	if line < 0 {
		line = tv.CursorPos.Ln
	}
	di, df := dv.AlignD.DiffForLine(line)
	if di < 0 || df.Tag == 'e' {
		return false
	}

	if ab == 0 {
		dv.BufA.Undos.Off = false
		// srcLen := len(dv.BufB.Lines[df.J2])
		spos := lexer.Pos{Ln: df.I1, Ch: 0}
		epos := lexer.Pos{Ln: df.I2, Ch: 0}
		src := dv.BufB.Region(spos, epos)
		dv.BufA.DeleteText(spos, epos, true)
		dv.BufA.InsertText(spos, src.ToBytes(), true) // we always just copy, is blank for delete..
		dv.Diffs.BtoA(di)
	} else {
		dv.BufB.Undos.Off = false
		spos := lexer.Pos{Ln: df.J1, Ch: 0}
		epos := lexer.Pos{Ln: df.J2, Ch: 0}
		src := dv.BufA.Region(spos, epos)
		dv.BufB.DeleteText(spos, epos, true)
		dv.BufB.InsertText(spos, src.ToBytes(), true)
		dv.Diffs.AtoB(di)
	}
	// dv.UpdateToolbar()
	return true
}

// UndoDiff undoes last applied change, if any.
func (dv *DiffEditor) UndoDiff(ab int) error {
	tva, tvb := dv.TextEditors()
	if ab == 1 {
		if !dv.Diffs.B.Undo() {
			err := errors.New("No more edits to undo")
			core.ErrorSnackbar(dv, err)
			return err
		}
		tvb.Undo()
	} else {
		if !dv.Diffs.A.Undo() {
			err := errors.New("No more edits to undo")
			core.ErrorSnackbar(dv, err)
			return err
		}
		tva.Undo()
	}
	return nil
}

func (dv *DiffEditor) MakeToolbar(p *tree.Plan) {
	txta := "A: " + fsx.DirAndFile(dv.FileA)
	if dv.RevA != "" {
		txta += ": " + dv.RevA
	}
	tree.Add(p, func(w *core.Text) {
		w.SetText(txta)
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Next").SetIcon(icons.KeyboardArrowDown).SetTooltip("move down to next diff region")
		w.OnClick(func(e events.Event) {
			dv.NextDiff(0)
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(len(dv.AlignD) <= 1, states.Disabled)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Prev").SetIcon(icons.KeyboardArrowUp).SetTooltip("move up to previous diff region")
		w.OnClick(func(e events.Event) {
			dv.PrevDiff(0)
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(len(dv.AlignD) <= 1, states.Disabled)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("A &lt;- B").SetIcon(icons.ContentCopy).SetTooltip("for current diff region, apply change from corresponding version in B, and move to next diff")
		w.OnClick(func(e events.Event) {
			dv.ApplyDiff(0, -1)
			dv.NextDiff(0)
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(len(dv.AlignD) <= 1, states.Disabled)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Undo").SetIcon(icons.Undo).SetTooltip("undo last diff apply action (A &lt;- B)")
		w.OnClick(func(e events.Event) {
			dv.UndoDiff(0)
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(!dv.BufA.Changed, states.Disabled)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Save").SetIcon(icons.Save).SetTooltip("save edited version of file with the given; prompts for filename")
		w.OnClick(func(e events.Event) {
			fb := core.NewSoloFuncButton(w).SetFunc(dv.SaveFileA)
			fb.Args[0].SetValue(dv.FileA)
			fb.CallFunc()
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(!dv.BufA.Changed, states.Disabled)
		})
	})

	tree.Add(p, func(w *core.Separator) {})

	txtb := "B: " + fsx.DirAndFile(dv.FileB)
	if dv.RevB != "" {
		txtb += ": " + dv.RevB
	}
	tree.Add(p, func(w *core.Text) {
		w.SetText(txtb)
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Next").SetIcon(icons.KeyboardArrowDown).SetTooltip("move down to next diff region")
		w.OnClick(func(e events.Event) {
			dv.NextDiff(1)
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(len(dv.AlignD) <= 1, states.Disabled)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Prev").SetIcon(icons.KeyboardArrowUp).SetTooltip("move up to previous diff region")
		w.OnClick(func(e events.Event) {
			dv.PrevDiff(1)
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(len(dv.AlignD) <= 1, states.Disabled)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("A -&gt; B").SetIcon(icons.ContentCopy).SetTooltip("for current diff region, apply change from corresponding version in A, and move to next diff")
		w.OnClick(func(e events.Event) {
			dv.ApplyDiff(1, -1)
			dv.NextDiff(1)
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(len(dv.AlignD) <= 1, states.Disabled)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Undo").SetIcon(icons.Undo).SetTooltip("undo last diff apply action (A -&gt; B)")
		w.OnClick(func(e events.Event) {
			dv.UndoDiff(1)
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(!dv.BufB.Changed, states.Disabled)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Save").SetIcon(icons.Save).SetTooltip("save edited version of file -- prompts for filename -- this will convert file back to its original form (removing side-by-side alignment) and end the diff editing function")
		w.OnClick(func(e events.Event) {
			fb := core.NewSoloFuncButton(w).SetFunc(dv.SaveFileB)
			fb.Args[0].SetValue(dv.FileB)
			fb.CallFunc()
		})
		w.Styler(func(s *styles.Style) {
			s.SetState(!dv.BufB.Changed, states.Disabled)
		})
	})
}

func (dv *DiffEditor) TextEditors() (*DiffTextEditor, *DiffTextEditor) {
	av := dv.Child(0).(*DiffTextEditor)
	bv := dv.Child(1).(*DiffTextEditor)
	return av, bv
}

////////////////////////////////////////////////////////////////////////////////
//   DiffTextEditor

// DiffTextEditor supports double-click based application of edits from one
// buffer to the other.
type DiffTextEditor struct {
	Editor
}

func (tv *DiffTextEditor) Init() {
	tv.Editor.Init()
	tv.HandleDoubleClick()
}

func (tv *DiffTextEditor) DiffEditor() *DiffEditor {
	return tree.ParentByType[*DiffEditor](tv)
}

func (tv *DiffTextEditor) HandleDoubleClick() {
	tv.On(events.DoubleClick, func(e events.Event) {
		pt := tv.PointToRelPos(e.Pos())
		if pt.X >= 0 && pt.X < int(tv.LineNumberOffset) {
			newPos := tv.PixelToCursor(pt)
			ln := newPos.Ln
			dv := tv.DiffEditor()
			if dv != nil && tv.Buffer != nil {
				if tv.Name == "text-a" {
					dv.ApplyDiff(0, ln)
				} else {
					dv.ApplyDiff(1, ln)
				}
			}
			e.SetHandled()
			return
		}
	})
}
