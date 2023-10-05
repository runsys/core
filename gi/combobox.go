// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"unicode/utf8"

	"goki.dev/colors"
	"goki.dev/cursors"
	"goki.dev/girl/states"
	"goki.dev/girl/styles"
	"goki.dev/girl/units"
	"goki.dev/goosi/events"
	"goki.dev/icons"
	"goki.dev/ki/v2"
	"goki.dev/laser"
	"goki.dev/pi/v2/complete"
)

// ComboBox is for selecting items from a dropdown list, with an optional
// edit TextField for typing directly.
// The items can be of any type, including enum values -- they are converted
// to strings for the display.  If the items are [icons.Icon] type, then they
// are displayed using icons instead.
type ComboBox struct {
	ButtonBase

	// provide a text field for editing the value, or just a button for selecting items?  Set the editable property
	Editable bool `xml:"editable" desc:"provide a text field for editing the value, or just a button for selecting items?  Set the editable property"`

	// current selected value
	CurVal any `json:"-" xml:"-" desc:"current selected value"`

	// current index in list of possible items
	CurIndex int `json:"-" xml:"-" desc:"current index in list of possible items"`

	// items available for selection
	Items []any `json:"-" xml:"-" desc:"items available for selection"`

	// an optional list of tooltips displayed on hover for combobox items; the indices for tooltips correspond to those for items
	Tooltips []string `json:"-" xml:"-" desc:"an optional list of tooltips displayed on hover for combobox items; the indices for tooltips correspond to those for items"`

	// if Editable is set to true, text that is displayed in the text field when it is empty, in a lower-contrast manner
	Placeholder string `desc:"if Editable is set to true, text that is displayed in the text field when it is empty, in a lower-contrast manner"`

	// the type of combo box
	Type ComboBoxTypes `desc:"the type of combo box"`

	// [view: -] signal for combo box, when a new value has been selected -- the signal type is the index of the selected item, and the data is the value
	// ComboSig ki.Signal `copy:"-" json:"-" xml:"-" view:"-" desc:"signal for combo box, when a new value has been selected -- the signal type is the index of the selected item, and the data is the value"`

	// maximum label length (in runes)
	MaxLength int `desc:"maximum label length (in runes)"`
}

func (cb *ComboBox) CopyFieldsFrom(frm any) {
	fr := frm.(*ComboBox)
	cb.ButtonBase.CopyFieldsFrom(&fr.ButtonBase)
	cb.Editable = fr.Editable
	cb.CurVal = fr.CurVal
	cb.CurIndex = fr.CurIndex
	cb.Items = fr.Items
	cb.MaxLength = fr.MaxLength
}

// ComboBoxTypes is an enum containing the
// different possible types of combo boxes
type ComboBoxTypes int //enums:enum

const (
	// ComboBoxFilled represents a filled
	// ComboBox with a background color
	// and a bottom border
	ComboBoxFilled ComboBoxTypes = iota
	// ComboBoxOutlined represents an outlined
	// ComboBox with a border on all sides
	// and no background color
	ComboBoxOutlined
)

func (cb *ComboBox) OnInit() {
	cb.ComboBoxHandlers()
	cb.ComboBoxStyles()
}

func (cb *ComboBox) ComboBoxHandlers() {
	cb.ButtonBaseHandlers()
	cb.ComboBoxKeys()
}

func (cb *ComboBox) ComboBoxStyles() {
	cb.AddStyles(func(s *styles.Style) {
		s.Cursor = cursors.Pointer
		s.Text.Align = styles.AlignCenter
		if cb.Editable {
			s.Padding.Set()
			s.Padding.Right.SetDp(16 * Prefs.DensityMul())
		} else {
			s.Border.Radius = styles.BorderRadiusExtraSmall
			s.Padding.Set(units.Dp(8*Prefs.DensityMul()), units.Dp(16*Prefs.DensityMul()))
		}
		s.Color = colors.Scheme.OnSurface
		switch cb.Type {
		case ComboBoxFilled:
			s.BackgroundColor.SetSolid(colors.Scheme.SurfaceContainerHighest)
			if cb.Editable {
				s.Border.Style.Set(styles.BorderNone)
				s.Border.Style.Bottom = styles.BorderSolid
				s.Border.Width.Set()
				s.Border.Width.Bottom = units.Dp(1)
				s.Border.Color.Set()
				s.Border.Color.Bottom = colors.Scheme.OnSurfaceVariant
				s.Border.Radius = styles.BorderRadiusExtraSmallTop
				// if cb.HasFlagWithin(CanFocus) {
				// todo:
				s.Border.Width.Bottom = units.Dp(2)
				s.Border.Color.Bottom = colors.Scheme.Primary.Base
				// }

			}
		case ComboBoxOutlined:
			s.Border.Style.Set(styles.BorderSolid)
			s.Border.Width.Set(units.Dp(1))
			s.Border.Color.Set(colors.Scheme.OnSurfaceVariant)
			if cb.Editable {
				s.Border.Radius = styles.BorderRadiusExtraSmall
				// if cb.HasFlagWithin(CanFocus) {
				// todo:
				s.Border.Width.Set(units.Dp(2))
				s.Border.Color.Set(colors.Scheme.Primary.Base)
				// }
			}
		}
	})
}

func (cb *ComboBox) OnChildAdded(child ki.Ki) {
	if _, wb := AsWidget(child); wb != nil {
		switch wb.Name() {
		case "icon":
			wb.AddStyles(func(s *styles.Style) {
				s.Margin.Set()
				s.Padding.Set()
			})
		case "label":
			wb.AddStyles(func(s *styles.Style) {
				s.Margin.Set()
				s.Padding.Set()
				s.AlignV = styles.AlignMiddle
			})
		case "text":
			text := child.(*TextField)
			text.Placeholder = cb.Placeholder
			if cb.Type == ComboBoxFilled {
				text.Type = TextFieldFilled
			} else {
				text.Type = TextFieldOutlined
			}
			text.AddStyles(func(s *styles.Style) {
				s.Border.Style.Set(styles.BorderNone)
				s.Border.Width.Set()
				if cb.MaxLength > 0 {
					s.SetMinPrefWidth(units.Ch(float32(cb.MaxLength)))
				}
			})
		case "ind-stretch":
			wb.AddStyles(func(s *styles.Style) {
				if cb.Editable {
					s.Width.SetDp(0)
				} else {
					s.Width.SetDp(16 * Prefs.DensityMul())
				}
			})
		case "indicator":
			wb.AddStyles(func(s *styles.Style) {
				s.Font.Size.SetEm(1.5)
				s.AlignV = styles.AlignMiddle
			})
		}
	}
}

// ConfigPartsIconText returns a standard config for creating parts, of icon
// and text left-to right in a row -- always makes text
func (cb *ComboBox) ConfigPartsIconText(config *ki.Config, icnm icons.Icon) (icIdx, txIdx int) {
	// todo: add some styles for button layout
	icIdx = -1
	txIdx = -1
	if icnm.IsValid() {
		icIdx = len(*config)
		config.Add(IconType, "icon")
		config.Add(SpaceType, "space")
	}
	txIdx = len(*config)
	config.Add(TextFieldType, "text")
	return
}

// ConfigPartsSetText sets part style props, using given props if not set in
// object props
func (cb *ComboBox) ConfigPartsSetText(txt string, txIdx, icIdx, indIdx int) {
	if txIdx >= 0 {
		tx := cb.Parts.Child(txIdx).(*TextField)
		tx.SetText(txt)
		tx.SetCompleter(tx, cb.CompleteMatch, cb.CompleteEdit)
	}
}

// ConfigPartsAddIndicatorSpace adds indicator with a space instead of a stretch
// for editable combobox, where textfield then takes up the rest of the space
func (bb *ButtonBase) ConfigPartsAddIndicatorSpace(config *ki.Config, defOn bool) int {
	needInd := (bb.HasMenu() || defOn) && bb.Indicator != icons.None
	if !needInd {
		return -1
	}
	indIdx := -1
	config.Add(SpaceType, "ind-stretch")
	indIdx = len(*config)
	config.Add(IconType, "indicator")
	return indIdx
}

func (cb *ComboBox) ConfigParts(sc *Scene) {
	cb.MakeMenuFunc = cb.MakeItemsMenu
	parts := cb.NewParts(LayoutHoriz)
	if eb, err := cb.PropTry("editable"); err == nil {
		cb.Editable, _ = laser.ToBool(eb)
	}
	config := ki.Config{}
	var icIdx, lbIdx, txIdx, indIdx int
	if cb.Editable {
		lbIdx = -1
		icIdx, txIdx = cb.ConfigPartsIconText(&config, cb.Icon)
		cb.SetProp("no-focus", true)
		indIdx = cb.ConfigPartsAddIndicatorSpace(&config, true) // use space instead of stretch
	} else {
		txIdx = -1
		icIdx, lbIdx = cb.ConfigPartsIconLabel(&config, cb.Icon, cb.Text)
		indIdx = cb.ConfigPartsAddIndicator(&config, true) // default on
	}
	mods, updt := parts.ConfigChildren(config)
	cb.ConfigPartsSetIconLabel(cb.Icon, cb.Text, icIdx, lbIdx)
	cb.ConfigPartsIndicator(indIdx)
	if txIdx >= 0 {
		cb.ConfigPartsSetText(cb.Text, txIdx, icIdx, indIdx)
	}
	if cb.MaxLength > 0 && lbIdx >= 0 {
		lbl := parts.Child(lbIdx).(*Label)
		lbl.SetMinPrefWidth(units.Ch(float32(cb.MaxLength)))
	}
	if mods {
		cb.UpdateEnd(updt)
	}
}

// TextField returns the text field of an editable combobox, and false if not made
func (cb *ComboBox) TextField() (*TextField, bool) {
	tff := cb.Parts.ChildByName("text", 2)
	if tff == nil {
		return nil, false
	}
	return tff.(*TextField), true
}

// MakeItems makes sure the Items list is made, and if not, or reset is true,
// creates one with the given capacity
func (cb *ComboBox) MakeItems(reset bool, capacity int) {
	if cb.Items == nil || reset {
		cb.Items = make([]any, 0, capacity)
	}
}

// SortItems sorts the items according to their labels
func (cb *ComboBox) SortItems(ascending bool) {
	sort.Slice(cb.Items, func(i, j int) bool {
		if ascending {
			return ToLabel(cb.Items[i]) < ToLabel(cb.Items[j])
		} else {
			return ToLabel(cb.Items[i]) > ToLabel(cb.Items[j])
		}
	})
}

// SetToMaxLength gets the maximum label length so that the width of the
// button label is automatically set according to the max length of all items
// in the list -- if maxLen > 0 then it is used as an upper do-not-exceed
// length
func (cb *ComboBox) SetToMaxLength(maxLen int) {
	ml := 0
	for _, it := range cb.Items {
		ml = max(ml, utf8.RuneCountInString(ToLabel(it)))
	}
	if maxLen > 0 {
		ml = min(ml, maxLen)
	}
	cb.MaxLength = ml
}

// ItemsFromTypes sets the Items list from a list of types -- see e.g.,
// AllImplementersOf or AllEmbedsOf in kit.TypeRegistry -- if setFirst then
// set current item to the first item in the list, sort sorts the list in
// ascending order, and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit
func (cb *ComboBox) ItemsFromTypes(tl []reflect.Type, setFirst, sort bool, maxLen int) {
	sz := len(tl)
	if sz == 0 {
		return
	}
	cb.Items = make([]any, sz)
	for i, typ := range tl {
		cb.Items[i] = typ
	}
	if sort {
		cb.SortItems(true)
	}
	if maxLen > 0 {
		cb.SetToMaxLength(maxLen)
	}
	if setFirst {
		cb.SetCurIndex(0)
	}
}

// ItemsFromStringList sets the Items list from a list of string values -- if
// setFirst then set current item to the first item in the list, and maxLen if
// > 0 auto-sets the width of the button to the contents, with the given upper
// limit
func (cb *ComboBox) ItemsFromStringList(el []string, setFirst bool, maxLen int) {
	sz := len(el)
	if sz == 0 {
		return
	}
	cb.Items = make([]any, sz)
	for i, str := range el {
		cb.Items[i] = str
	}
	if maxLen > 0 {
		cb.SetToMaxLength(maxLen)
	}
	if setFirst {
		cb.SetCurIndex(0)
	}
}

// ItemsFromIconList sets the Items list from a list of icons.Icon values -- if
// setFirst then set current item to the first item in the list, and maxLen if
// > 0 auto-sets the width of the button to the contents, with the given upper
// limit
func (cb *ComboBox) ItemsFromIconList(el []icons.Icon, setFirst bool, maxLen int) {
	sz := len(el)
	if sz == 0 {
		return
	}
	cb.Items = make([]any, sz)
	for i, str := range el {
		cb.Items[i] = str
	}
	if maxLen > 0 {
		cb.SetToMaxLength(maxLen)
	}
	if setFirst {
		cb.SetCurIndex(0)
	}
}

/*

// ItemsFromEnumList sets the Items list from a list of enum values (see
// kit.EnumRegistry) -- if setFirst then set current item to the first item in
// the list, and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit
func (cb *ComboBox) ItemsFromEnumList(el []kit.EnumValue, setFirst bool, maxLen int) {
	sz := len(el)
	if sz == 0 {
		return
	}
	cb.Items = make([]any, sz)
	cb.Tooltips = make([]string, sz)
	for i, enum := range el {
		cb.Items[i] = enum
		cb.Tooltips[i] = enum.Desc
	}
	if maxLen > 0 {
		cb.SetToMaxLength(maxLen)
	}
	if setFirst {
		cb.SetCurIndex(0)
	}
}

// ItemsFromEnum sets the Items list from an enum type, which must be
// registered on kit.EnumRegistry -- if setFirst then set current item to the
// first item in the list, and maxLen if > 0 auto-sets the width of the button
// to the contents, with the given upper limit -- see kit.EnumRegistry, and
// maxLen if > 0 auto-sets the width of the button to the contents, with the
// given upper limit
func (cb *ComboBox) ItemsFromEnum(enumtyp reflect.Type, setFirst bool, maxLen int) {
	cb.ItemsFromEnumList(kit.Enums.TypeValues(enumtyp, true), setFirst, maxLen)
}
*/

// FindItem finds an item on list of items and returns its index
func (cb *ComboBox) FindItem(it any) int {
	if cb.Items == nil {
		return -1
	}
	for i, v := range cb.Items {
		if v == it {
			return i
		}
	}
	return -1
}

// SetCurVal sets the current value (CurVal) and the corresponding CurIndex
// for that item on the current Items list (adds to items list if not found)
// -- returns that index -- and sets the text to the string value of that
// value (using standard Stringer string conversion)
func (cb *ComboBox) SetCurVal(it any) int {
	cb.CurVal = it
	cb.CurIndex = cb.FindItem(it)
	if cb.CurIndex < 0 { // add to list if not found..
		cb.CurIndex = len(cb.Items)
		cb.Items = append(cb.Items, it)
	}
	cb.ShowCurVal()
	return cb.CurIndex
}

// SetCurIndex sets the current index (CurIndex) and the corresponding CurVal
// for that item on the current Items list (-1 if not found) -- returns value
// -- and sets the text to the string value of that value (using standard
// Stringer string conversion)
func (cb *ComboBox) SetCurIndex(idx int) any {
	cb.CurIndex = idx
	if idx < 0 || idx >= len(cb.Items) {
		cb.CurVal = nil
		cb.SetText(fmt.Sprintf("idx %v > len", idx))
	} else {
		cb.CurVal = cb.Items[idx]
		cb.ShowCurVal()
	}
	return cb.CurVal
}

// ShowCurVal updates the display to present the
// currently-selected value (CurVal)
func (cb *ComboBox) ShowCurVal() {
	if icnm, isic := cb.CurVal.(icons.Icon); isic {
		cb.SetIcon(icnm)
	} else {
		cb.SetText(ToLabel(cb.CurVal))
	}
}

// SelectItem selects a given item and updates the display to it
func (cb *ComboBox) SelectItem(idx int) {
	if cb.This() == nil {
		return
	}
	updt := cb.UpdateStart()
	cb.SetCurIndex(idx)
	cb.UpdateEnd(updt)
}

// SelectItemAction selects a given item and emits the index as the ComboSig signal
// and the selected item as the data.
func (cb *ComboBox) SelectItemAction(idx int) {
	if cb.This() == nil {
		return
	}
	updt := cb.UpdateStart()
	cb.SelectItem(idx)
	// cb.ComboSig.Emit(cb.This(), int64(cb.CurIndex), cb.CurVal)
	cb.UpdateEnd(updt)
}

// MakeItemsMenu makes menu of all the items.  It is set as the
// MakeMenuFunc for this combobox.
func (cb *ComboBox) MakeItemsMenu(obj Widget, menu *MenuActions) {
	nitm := len(cb.Items)
	if cb.Menu == nil {
		cb.Menu = make(MenuActions, 0, nitm)
	}
	sz := len(cb.Menu)
	if nitm < sz {
		cb.Menu = cb.Menu[0:nitm]
	}
	if nitm == 0 {
		return
	}
	_, ics := cb.Items[0].(icons.Icon) // if true, we render as icons
	for i, it := range cb.Items {
		var ac *Action
		if sz > i {
			ac = cb.Menu[i].(*Action)
		} else {
			ac = &Action{}
			ki.InitNode(ac)
			cb.Menu = append(cb.Menu, ac.This().(Widget))
		}
		nm := "Item_" + strconv.Itoa(i)
		ac.SetName(nm)
		if ics {
			ac.Icon = it.(icons.Icon)
			ac.Tooltip = string(ac.Icon)
		} else {
			ac.Text = ToLabel(it)
			if len(cb.Tooltips) > i {
				ac.Tooltip = cb.Tooltips[i]
			}
		}
		ac.Data = i // index is the data
		ac.SetSelected(i == cb.CurIndex)
		ac.SetAsMenu()
		idx := i
		ac.On(events.Click, func(e events.Event) {
			cb.SelectItemAction(idx)
		})
	}
}

func (cb *ComboBox) ComboBoxKeys() {
	cb.On(events.KeyChord, func(e events.Event) {
		if cb.StateIs(states.Disabled) {
			return
		}
		if KeyEventTrace {
			fmt.Printf("ComboBox KeyChordEvent: %v\n", cb.Path())
		}
		kf := KeyFun(e.KeyChord())
		switch {
		case kf == KeyFunMoveUp:
			e.SetHandled()
			if len(cb.Items) > 0 {
				idx := cb.CurIndex - 1
				if idx < 0 {
					idx += len(cb.Items)
				}
				cb.SelectItemAction(idx)
			}
		case kf == KeyFunMoveDown:
			e.SetHandled()
			if len(cb.Items) > 0 {
				idx := cb.CurIndex + 1
				if idx >= len(cb.Items) {
					idx -= len(cb.Items)
				}
				cb.SelectItemAction(idx)
			}
		case kf == KeyFunPageUp:
			e.SetHandled()
			if len(cb.Items) > 10 {
				idx := cb.CurIndex - 10
				for idx < 0 {
					idx += len(cb.Items)
				}
				cb.SelectItemAction(idx)
			}
		case kf == KeyFunPageDown:
			e.SetHandled()
			if len(cb.Items) > 10 {
				idx := cb.CurIndex + 10
				for idx >= len(cb.Items) {
					idx -= len(cb.Items)
				}
				cb.SelectItemAction(idx)
			}
		case kf == KeyFunEnter || (!cb.Editable && e.KeyRune() == ' '):
			// if !(kt.Rune == ' ' && cbb.Sc.Type == ScCompleter) {
			e.SetHandled()
			cb.Send(events.Click, e)
			// }
		}
	})
}

// CompleteMatch is the [complete.MatchFunc] used for the
// editable textfield part of the ComboBox (if it exists).
func (cb *ComboBox) CompleteMatch(data any, text string, posLn, posCh int) (md complete.Matches) {
	md.Seed = text
	comps := make(complete.Completions, len(cb.Items))
	for idx, item := range cb.Items {
		tooltip := ""
		if len(cb.Tooltips) > idx {
			tooltip = cb.Tooltips[idx]
		}
		comps[idx] = complete.Completion{
			Text: laser.ToString(item),
			Desc: tooltip,
		}
	}
	md.Matches = complete.MatchSeedCompletion(comps, md.Seed)
	return md
}

// CompleteEdit is the [complete.EditFunc] used for the
// editable textfield part of the ComboBox (if it exists).
func (cb *ComboBox) CompleteEdit(data any, text string, cursorPos int, completion complete.Completion, seed string) (ed complete.Edit) {
	return complete.Edit{
		NewText:       completion.Text,
		ForwardDelete: len([]rune(text)),
	}
}
