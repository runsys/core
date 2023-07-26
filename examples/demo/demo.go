// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/icons"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
)

func main() {
	gimain.Main(mainrun)
}

func mainrun() {
	gi.SetAppName("gogi-demo")
	gi.SetAppAbout("The GoGi Demo demonstrates the various features of the GoGi 2D and 3D Go GUI framework.")

	rec := ki.Node{}          // receiver for events
	rec.InitName(&rec, "rec") // this is essential for root objects not owned by other Ki tree nodes

	win := gi.NewMainWindow("gogi-demo", "The GoGi Demo", 1024, 768)
	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tv := gi.AddNewTabView(mfr, "tv")

	_, home := tv.AddNewTabFrame(gi.KiT_Frame, "Home")
	home.Lay = gi.LayoutVert
	home.AddStyleFunc(gi.StyleFuncFinal, func() {
		home.Spacing.SetEx(1)
		home.Style.Padding.Set(units.Px(8))
		home.Style.MaxWidth.SetPx(-1)
		home.Style.MaxHeight.SetPx(-1)
	})

	title := gi.AddNewLabel(home, "title", "The GoGi Demo")
	title.Type = gi.LabelH1

	desc := gi.AddNewLabel(home, "desc", "A demonstration of the <i>various</i> features of the <u>GoGi</u> 2D and 3D Go GUI <b>framework.</b>")
	desc.Type = gi.LabelP

	bdesc := gi.AddNewLabel(home, "bdesc", "Buttons")
	bdesc.Type = gi.LabelH3

	brow := gi.AddNewLayout(home, "brow", gi.LayoutHorizFlow)
	brow.AddStyleFunc(gi.StyleFuncFinal, func() {
		brow.Spacing.SetEx(1)
		brow.Style.MaxWidth.SetPx(-1)
	})

	bpri := gi.AddNewButton(brow, "buttonPrimary")
	bpri.Text = "Primary Button"
	bpri.Type = gi.ButtonPrimary
	bpri.Icon = icons.PlayArrow

	bsec := gi.AddNewButton(brow, "buttonSecondary")
	bsec.Text = "Secondary Button"
	bsec.Type = gi.ButtonSecondary
	// bsec.Icon = icons.Settings

	bdef := gi.AddNewButton(brow, "buttonDefault")
	bdef.Text = "Default Button"
	bdef.Icon = icons.Reviews

	browi := gi.AddNewLayout(home, "browi", gi.LayoutHorizFlow)
	browi.AddStyleFunc(gi.StyleFuncFinal, func() {
		browi.Spacing.SetEx(1)
		browi.Style.MaxWidth.SetPx(-1)
	})

	bprii := gi.AddNewButton(browi, "buttonPrimaryInactive")
	bprii.Text = "Inactive Primary Button"
	bprii.Type = gi.ButtonPrimary
	bprii.Icon = icons.OpenInNew
	bprii.SetInactive()

	bseci := gi.AddNewButton(browi, "buttonSecondaryInactive")
	bseci.Text = "Inactive Secondary Button"
	bseci.Type = gi.ButtonSecondary
	bseci.Icon = icons.Settings
	bseci.SetInactive()

	bdefi := gi.AddNewButton(browi, "buttonDefaultInactive")
	bdefi.Text = "Inactive Default Button"
	bdefi.SetInactive()

	idesc := gi.AddNewLabel(home, "idesc", "Inputs")
	idesc.Type = gi.LabelH3

	irow := gi.AddNewLayout(home, "irow", gi.LayoutVert)
	irow.AddStyleFunc(gi.StyleFuncFinal, func() {
		irow.Spacing.SetEx(1)
		irow.Style.MaxWidth.SetPx(-1)
		irow.Style.MaxHeight.SetPx(500)
	})

	check := gi.AddNewCheckBox(irow, "check")
	check.Text = "Checkbox"

	tfield := gi.AddNewTextField(irow, "tfield")
	tfield.Placeholder = "Text Field"
	// tfield.AddStyleFunc(gi.StyleFuncFinal, func() {
	// 	tfield.Style.BackgroundColor.SetColor(colors.Green)
	// })

	sbox := gi.AddNewSpinBox(irow, "sbox")
	sbox.Value = 0.5

	cbox := gi.AddNewComboBox(irow, "cbox")
	cbox.Text = "Select an option"
	cbox.Items = []any{"Option 1", "Option 2", "Option 3"}
	cbox.Tooltips = []string{"A description for Option 1", "A description for Option 2", "A description for Option 3"}

	cboxe := gi.AddNewComboBox(irow, "cboxe")
	cboxe.Editable = true
	cboxe.Text = "Select or type an option"
	cboxe.Items = []any{"Option 1", "Option 2", "Option 3"}
	cboxe.Tooltips = []string{"A description for Option 1", "A description for Option 2", "A description for Option 3"}

	bbox := gi.AddNewButtonBox(irow, "bbox")
	bbox.Items = []string{"Checkbox 1", "Checkbox 2", "Checkbox 3"}
	bbox.Tooltips = []string{"A description for Checkbox 1", "A description for Checkbox 2", "A description for Checkbox 3"}

	bboxr := gi.AddNewButtonBox(irow, "bboxr")
	bboxr.Items = []string{"Radio Button 1", "Radio Button 2", "Radio Button 3"}
	bboxr.Tooltips = []string{"A description for Radio Button 1", "A description for Radio Button 2", "A description for Radio Button 3"}
	bboxr.Mutex = true

	tbuf := &giv.TextBuf{}
	tbuf.InitName(tbuf, "tbuf")
	tbuf.SetText([]byte("A keyboard-navigable, multi-line\ntext editor with support for\ncompletion and syntax highlighting"))

	tview := giv.AddNewTextView(home, "tview")
	tview.SetBuf(tbuf)
	tview.AddStyleFunc(gi.StyleFuncFinal, func() {
		tview.Style.MaxWidth.SetPx(500)
		tview.Style.MaxHeight.SetPx(300)
	})

	bmap := gi.AddNewBitmap(home, "bmap")
	err := bmap.OpenImage("gopher.png", 300, 300)
	if err != nil {
		fmt.Println("error loading gopher image:", err)
	}

	// Main Menu

	appnm := gi.AppName()
	mmen := win.MainMenu
	mmen.ConfigMenus([]string{appnm, "File", "Edit", "Window"})

	amen := win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
	amen.Menu.AddAppMenu(win)

	fmen := win.MainMenu.ChildByName("File", 0).(*gi.Action)
	fmen.Menu.AddAction(gi.ActOpts{Label: "New", ShortcutKey: gi.KeyFunMenuNew},
		rec.This(), func(recv, send ki.Ki, sig int64, data any) {
			fmt.Printf("File:New menu action triggered\n")
		})
	fmen.Menu.AddAction(gi.ActOpts{Label: "Open", ShortcutKey: gi.KeyFunMenuOpen},
		rec.This(), func(recv, send ki.Ki, sig int64, data any) {
			fmt.Printf("File:Open menu action triggered\n")
		})
	fmen.Menu.AddAction(gi.ActOpts{Label: "Save", ShortcutKey: gi.KeyFunMenuSave},
		rec.This(), func(recv, send ki.Ki, sig int64, data any) {
			fmt.Printf("File:Save menu action triggered\n")
		})
	fmen.Menu.AddAction(gi.ActOpts{Label: "Save As..", ShortcutKey: gi.KeyFunMenuSaveAs},
		rec.This(), func(recv, send ki.Ki, sig int64, data any) {
			fmt.Printf("File:SaveAs menu action triggered\n")
		})
	fmen.Menu.AddSeparator("csep")
	fmen.Menu.AddAction(gi.ActOpts{Label: "Close Window", ShortcutKey: gi.KeyFunWinClose},
		win.This(), func(recv, send ki.Ki, sig int64, data any) {
			win.CloseReq()
		})

	emen := win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
	emen.Menu.AddCopyCutPaste(win)

	inQuitPrompt := false
	gi.SetQuitReqFunc(func() {
		if inQuitPrompt {
			return
		}
		inQuitPrompt = true
		gi.PromptDialog(vp, gi.DlgOpts{Title: "Really Quit?",
			Prompt: "Are you <i>sure</i> you want to quit?"}, gi.AddOk, gi.AddCancel,
			win.This(), func(recv, send ki.Ki, sig int64, data any) {
				if sig == int64(gi.DialogAccepted) {
					gi.Quit()
				} else {
					inQuitPrompt = false
				}
			})
	})

	gi.SetQuitCleanFunc(func() {
		fmt.Printf("Doing final Quit cleanup here..\n")
	})

	inClosePrompt := false
	win.SetCloseReqFunc(func(w *gi.Window) {
		if inClosePrompt {
			return
		}
		inClosePrompt = true
		gi.PromptDialog(vp, gi.DlgOpts{Title: "Really Close Window?",
			Prompt: "Are you <i>sure</i> you want to close the window?  This will Quit the App as well."}, gi.AddOk, gi.AddCancel,
			win.This(), func(recv, send ki.Ki, sig int64, data any) {
				if sig == int64(gi.DialogAccepted) {
					gi.Quit()
				} else {
					inClosePrompt = false
				}
			})
	})

	win.SetCloseCleanFunc(func(w *gi.Window) {
		fmt.Printf("Doing final Close cleanup here..\n")
	})

	win.MainMenuUpdated()
	vp.UpdateEndNoSig(updt)
	win.StartEventLoop()
}
