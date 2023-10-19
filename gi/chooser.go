// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"unicode/utf8"

	"goki.dev/colors"
	"goki.dev/cursors"
	"goki.dev/enums"
	"goki.dev/girl/abilities"
	"goki.dev/girl/states"
	"goki.dev/girl/styles"
	"goki.dev/girl/units"
	"goki.dev/goosi/events"
	"goki.dev/gti"
	"goki.dev/icons"
	"goki.dev/ki/v2"
	"goki.dev/pi/v2/complete"
)

// Chooser is for selecting items from a dropdown list, with an optional
// edit TextField for typing directly.
// The items can be of any type, including enum values -- they are converted
// to strings for the display.  If the items are of type [icons.Icon], then they
// are displayed using icons instead.
type Chooser struct {
	WidgetBase

	// the type of combo box
	Type ChooserTypes

	// optional icon
	Icon icons.Icon `view:"show-name"`

	// name of the indicator icon to present.
	Indicator icons.Icon `view:"show-name"`

	// provide a text field for editing the value, or just a button for selecting items?  Set the editable property
	Editable bool

	// TODO(kai): implement AllowNew button

	// whether to allow the user to add new items to the combo box through the editable textfield (if Editable is set to true) and a button at the end of the combo box menu
	AllowNew bool

	// CurValLabel is the string label for the current value
	CurValLabel string

	// current selected value
	CurVal any `json:"-" xml:"-"`

	// current index in list of possible items
	CurIndex int `json:"-" xml:"-"`

	// items available for selection
	Items []any `json:"-" xml:"-"`

	// an optional list of tooltips displayed on hover for Chooser items; the indices for tooltips correspond to those for items
	Tooltips []string `json:"-" xml:"-"`

	// if Editable is set to true, text that is displayed in the text field when it is empty, in a lower-contrast manner
	Placeholder string

	// maximum label length (in runes)
	MaxLength int
}

func (ch *Chooser) CopyFieldsFrom(frm any) {
	fr := frm.(*Chooser)
	ch.WidgetBase.CopyFieldsFrom(&fr.WidgetBase)
	ch.Editable = fr.Editable
	ch.CurVal = fr.CurVal
	ch.CurIndex = fr.CurIndex
	ch.Items = fr.Items
	ch.MaxLength = fr.MaxLength
}

// ChooserTypes is an enum containing the
// different possible types of combo boxes
type ChooserTypes int32 //enums:enum -trim-prefix Chooser

const (
	// ChooserFilled represents a filled
	// Chooser with a background color
	// and a bottom border
	ChooserFilled ChooserTypes = iota
	// ChooserOutlined represents an outlined
	// Chooser with a border on all sides
	// and no background color
	ChooserOutlined
)

func (ch *Chooser) OnInit() {
	ch.HandleChooserEvents()
	ch.ChooserStyles()
}

func (ch *Chooser) ChooserStyles() {
	ch.Icon = icons.None
	ch.Indicator = icons.KeyboardArrowRight
	ch.Style(func(s *styles.Style) {
		s.SetAbilities(true, abilities.Activatable, abilities.Focusable, abilities.FocusWithinable, abilities.Hoverable, abilities.LongHoverable)
		s.Cursor = cursors.Pointer
		s.Text.Align = styles.AlignCenter
		if ch.Editable {
			s.Padding.Set()
			s.Padding.Right.SetDp(16)
		} else {
			s.Border.Radius = styles.BorderRadiusExtraSmall
			s.Padding.Set(units.Dp(8), units.Dp(16))
		}
		s.Color = colors.Scheme.OnSurface
		switch ch.Type {
		case ChooserFilled:
			s.StateLayer += 0.06
			if ch.Editable {
				s.Border.Style.Set(styles.BorderNone)
				s.Border.Style.Bottom = styles.BorderSolid
				s.Border.Width.Set()
				s.Border.Width.Bottom = units.Dp(1)
				s.Border.Color.Set()
				s.Border.Color.Bottom = colors.Scheme.OnSurfaceVariant
				s.Border.Radius = styles.BorderRadiusExtraSmallTop
				if s.Is(states.FocusedWithin) {
					s.Border.Width.Bottom = units.Dp(2)
					s.Border.Color.Bottom = colors.Scheme.Primary.Base
				}
			}
		case ChooserOutlined:
			s.Border.Style.Set(styles.BorderSolid)
			s.Border.Width.Set(units.Dp(1))
			s.Border.Color.Set(colors.Scheme.OnSurfaceVariant)
			if ch.Editable {
				s.Border.Radius = styles.BorderRadiusExtraSmall
				if s.Is(states.FocusedWithin) {
					s.Border.Width.Set(units.Dp(2))
					s.Border.Color.Set(colors.Scheme.Primary.Base)
				}
			}
		}
		if s.Is(states.Selected) {
			s.BackgroundColor.SetSolid(colors.Scheme.Select.Container)
		}
		if s.Is(states.Disabled) {
			s.Cursor = cursors.NotAllowed
		}
	})
	ch.OnWidgetAdded(func(w Widget) {
		switch w.PathFrom(ch.This()) {
		case "parts/icon":
			w.Style(func(s *styles.Style) {
				s.Margin.Set()
				s.Padding.Set()
			})
		case "parts/label":
			w.Style(func(s *styles.Style) {
				s.SetAbilities(false, abilities.Selectable, abilities.DoubleClickable)
				s.Cursor = cursors.None
				s.Margin.Set()
				s.Padding.Set()
				s.AlignV = styles.AlignMiddle
				if ch.MaxLength > 0 {
					s.SetMinPrefWidth(units.Ch(float32(ch.MaxLength)))
				}
			})
		case "parts/text":
			text := w.(*TextField)
			text.Placeholder = ch.Placeholder
			if ch.Type == ChooserFilled {
				text.Type = TextFieldFilled
			} else {
				text.Type = TextFieldOutlined
			}
			ch.TextFieldHandlers(text)
			text.Style(func(s *styles.Style) {
				s.Border.Style.Set(styles.BorderNone)
				s.Border.Width.Set()
				if ch.MaxLength > 0 {
					s.SetMinPrefWidth(units.Ch(float32(ch.MaxLength)))
				}
			})
		case "parts/ind-stretch":
			w.Style(func(s *styles.Style) {
				if ch.Editable {
					s.Width.SetDp(0)
				} else {
					s.Width.SetDp(16)
				}
			})
		case "parts/indicator":
			w.Style(func(s *styles.Style) {
				s.Font.Size.SetDp(16)
				s.AlignV = styles.AlignMiddle
			})
		}
	})
}

// SetType sets the styling type of the combo box
func (ch *Chooser) SetType(typ ChooserTypes) *Chooser {
	updt := ch.UpdateStart()
	ch.Type = typ
	ch.UpdateEndLayout(updt)
	return ch
}

// SetType sets whether the combo box is editable
func (ch *Chooser) SetEditable(editable bool) *Chooser {
	updt := ch.UpdateStart()
	ch.Editable = editable
	ch.UpdateEndLayout(updt)
	return ch
}

// SetAllowNew sets whether to allow the user to add new values
func (ch *Chooser) SetAllowNew(allowNew bool) *Chooser {
	updt := ch.UpdateStart()
	ch.AllowNew = allowNew
	ch.UpdateEndLayout(updt)
	return ch
}

func (ch *Chooser) ConfigWidget(sc *Scene) {
	ch.ConfigParts(sc)
}

func (ch *Chooser) ConfigParts(sc *Scene) {
	parts := ch.NewParts(LayoutHoriz)
	config := ki.Config{}

	icIdx := -1
	var lbIdx, txIdx, indIdx int
	if ch.Icon.IsValid() {
		config.Add(IconType, "icon")
		config.Add(SpaceType, "space")
		icIdx = 0
	}
	if ch.Editable {
		lbIdx = -1
		txIdx = len(config)
		config.Add(TextFieldType, "text")
	} else {
		txIdx = -1
		lbIdx = len(config)
		config.Add(LabelType, "label")
	}
	if !ch.Indicator.IsValid() {
		ch.Indicator = icons.KeyboardArrowRight
	}
	indIdx = len(config)
	config.Add(IconType, "indicator")

	mods, updt := parts.ConfigChildren(config)

	if icIdx >= 0 {
		ic := ch.Parts.Child(icIdx).(*Icon)
		ic.SetIcon(ch.Icon)
	}
	if ch.Editable {
		tx := ch.Parts.Child(txIdx).(*TextField)
		tx.SetText(ch.CurValLabel)
		tx.SetCompleter(tx, ch.CompleteMatch, ch.CompleteEdit)
	} else {
		lbl := ch.Parts.Child(lbIdx).(*Label)
		lbl.SetText(ch.CurValLabel)
		lbl.Config(ch.Sc) // this is essential
	}
	{ // indicator
		ic := ch.Parts.Child(indIdx).(*Icon)
		ic.SetIcon(ch.Indicator)
	}
	if mods {
		parts.UpdateEnd(updt)
		ch.SetNeedsLayoutUpdate(sc, updt)
	}
}

// LabelWidget returns the label widget if present
func (ch *Chooser) LabelWidget() *Label {
	if ch.Parts == nil {
		return nil
	}
	lbi := ch.Parts.ChildByName("label")
	if lbi == nil {
		return nil
	}
	return lbi.(*Label)
}

// IconWidget returns the icon widget if present
func (ch *Chooser) IconWidget() *Icon {
	if ch.Parts == nil {
		return nil
	}
	ici := ch.Parts.ChildByName("icon")
	if ici == nil {
		return nil
	}
	return ici.(*Icon)
}

// TextField returns the text field of an editable Chooser
// if present
func (ch *Chooser) TextField() (*TextField, bool) {
	if ch.Parts == nil {
		return nil, false
	}
	tf := ch.Parts.ChildByName("text", 2)
	if tf == nil {
		return nil, false
	}
	return tf.(*TextField), true
}

// MakeItems makes sure the Items list is made, and if not, or reset is true,
// creates one with the given capacity
func (ch *Chooser) MakeItems(reset bool, capacity int) {
	if ch.Items == nil || reset {
		ch.Items = make([]any, 0, capacity)
	}
}

// SortItems sorts the items according to their labels
func (ch *Chooser) SortItems(ascending bool) {
	sort.Slice(ch.Items, func(i, j int) bool {
		if ascending {
			return ToLabel(ch.Items[i]) < ToLabel(ch.Items[j])
		} else {
			return ToLabel(ch.Items[i]) > ToLabel(ch.Items[j])
		}
	})
}

// SetToMaxLength gets the maximum label length so that the width of the
// button label is automatically set according to the max length of all items
// in the list -- if maxLen > 0 then it is used as an upper do-not-exceed
// length
func (ch *Chooser) SetToMaxLength(maxLen int) {
	ml := 0
	for _, it := range ch.Items {
		ml = max(ml, utf8.RuneCountInString(ToLabel(it)))
	}
	if maxLen > 0 {
		ml = min(ml, maxLen)
	}
	ch.MaxLength = ml
}

// ItemsFromTypes sets the Items list from a list of types, e.g., from gti.AllEmbedersOf.
// If setFirst then set current item to the first item in the list,
// and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit.
func (ch *Chooser) ItemsFromTypes(tl []*gti.Type, setFirst, sort bool, maxLen int) *Chooser {
	n := len(tl)
	if n == 0 {
		return ch
	}
	ch.Items = make([]any, n)
	for i, typ := range tl {
		ch.Items[i] = typ
	}
	if sort {
		ch.SortItems(true)
	}
	if maxLen > 0 {
		ch.SetToMaxLength(maxLen)
	}
	if setFirst {
		ch.SetCurIndex(0)
	}
	return ch
}

// ItemsFromStringList sets the Items list from a list of string values.
// If setFirst then set current item to the first item in the list,
// and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit.
func (ch *Chooser) ItemsFromStringList(el []string, setFirst bool, maxLen int) *Chooser {
	n := len(el)
	if n == 0 {
		return ch
	}
	ch.Items = make([]any, n)
	for i, str := range el {
		ch.Items[i] = str
	}
	if maxLen > 0 {
		ch.SetToMaxLength(maxLen)
	}
	if setFirst {
		ch.SetCurIndex(0)
	}
	return ch
}

// ItemsFromIconList sets the Items list from a list of icons.Icon values.
// If setFirst then set current item to the first item in the list,
// and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit.
func (ch *Chooser) ItemsFromIconList(el []icons.Icon, setFirst bool, maxLen int) *Chooser {
	n := len(el)
	if n == 0 {
		return ch
	}
	ch.Items = make([]any, n)
	for i, str := range el {
		ch.Items[i] = str
	}
	if maxLen > 0 {
		ch.SetToMaxLength(maxLen)
	}
	if setFirst {
		ch.SetCurIndex(0)
	}
	return ch
}

// ItemsFromEnumListsets the Items list from a list of enums.Enum values.
// If setFirst then set current item to the first item in the list,
// and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit.
func (ch *Chooser) ItemsFromEnums(el []enums.Enum, setFirst bool, maxLen int) *Chooser {
	n := len(el)
	if n == 0 {
		return ch
	}
	ch.Items = make([]any, n)
	ch.Tooltips = make([]string, n)
	for i, enum := range el {
		ch.Items[i] = enum
		ch.Tooltips[i] = enum.Desc()
	}
	if maxLen > 0 {
		ch.SetToMaxLength(maxLen)
	}
	if setFirst {
		ch.SetCurIndex(0)
	}
	return ch
}

// ItemsFromEnum sets the Items list from given enums.Enum Values().
// If setFirst then set current item to the first item in the list,
// and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit.
func (ch *Chooser) ItemsFromEnum(enum enums.Enum, setFirst bool, maxLen int) *Chooser {
	return ch.ItemsFromEnums(enum.Values(), setFirst, maxLen)
}

// FindItem finds an item on list of items and returns its index
func (ch *Chooser) FindItem(it any) int {
	if ch.Items == nil {
		return -1
	}
	for i, v := range ch.Items {
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
func (ch *Chooser) SetCurVal(it any) int {
	ch.CurVal = it
	ch.CurIndex = ch.FindItem(it)
	if ch.CurIndex < 0 { // add to list if not found..
		ch.CurIndex = len(ch.Items)
		ch.Items = append(ch.Items, it)
	}
	ch.ShowCurVal(ToLabel(ch.CurVal))
	return ch.CurIndex
}

// SetCurIndex sets the current index (CurIndex) and the corresponding CurVal
// for that item on the current Items list (-1 if not found) -- returns value
// -- and sets the text to the string value of that value (using standard
// Stringer string conversion)
func (ch *Chooser) SetCurIndex(idx int) any {
	ch.CurIndex = idx
	if idx < 0 || idx >= len(ch.Items) {
		ch.CurVal = nil
		ch.ShowCurVal(fmt.Sprintf("idx %v > len", idx))
	} else {
		ch.CurVal = ch.Items[idx]
		ch.ShowCurVal(ToLabel(ch.CurVal))
	}
	return ch.CurVal
}

// SetIcon sets the Icon to given icon name (could be empty or 'none') and
// updates the button.
// Use this for optimized auto-updating based on nature of changes made.
// Otherwise, can set Icon directly followed by ReConfig()
func (ch *Chooser) SetIcon(iconName icons.Icon) *Chooser {
	if ch.Icon == iconName {
		return ch
	}
	updt := ch.UpdateStart()
	recfg := (ch.Icon == "" && iconName != "") || (ch.Icon != "" && iconName == "")
	ch.Icon = iconName
	if recfg {
		ch.ConfigParts(ch.Sc)
	} else {
		ic := ch.IconWidget()
		if ic != nil {
			ic.SetIcon(ch.Icon)
		}
	}
	ch.UpdateEndLayout(updt)
	return ch
}

// ShowCurVal updates the display to present the
// currently-selected value (CurVal)
func (ch *Chooser) ShowCurVal(label string) {
	updt := ch.UpdateStart()
	defer ch.UpdateEndRender(updt)

	ch.CurValLabel = label
	if ch.Editable {
		tf, ok := ch.TextField()
		if ok {
			tf.SetText(ch.CurValLabel)
		}
	} else {
		if icnm, isic := ch.CurVal.(icons.Icon); isic {
			ch.SetIcon(icnm)
		} else {
			lbl := ch.LabelWidget()
			if lbl != nil {
				lbl.SetText(ch.CurValLabel)
			}
		}
	}
}

// SelectItem selects a given item and updates the display to it
func (ch *Chooser) SelectItem(idx int) *Chooser {
	if ch.This() == nil {
		return ch
	}
	updt := ch.UpdateStart()
	ch.SetCurIndex(idx)
	ch.UpdateEndLayout(updt)
	return ch
}

// SelectItemAction selects a given item and updates the display to it
// and sends a Changed event to indicate that the value has changed.
func (ch *Chooser) SelectItemAction(idx int) {
	if ch.This() == nil {
		return
	}
	ch.SelectItem(idx)
	ch.SendChange()
}

// MakeItemsMenu constructs a menu of all the items.
// It is automatically set as the [Button.Menu] for the Chooser.
func (ch *Chooser) MakeItemsMenu(m *Scene) {
	if len(ch.Items) == 0 {
		return
	}
	_, ics := ch.Items[0].(icons.Icon) // if true, we render as icons
	for i, it := range ch.Items {
		nm := "item-" + strconv.Itoa(i)
		bt := NewButton(m, nm).SetType(ButtonMenu)
		if ics {
			bt.Icon = it.(icons.Icon)
			bt.Tooltip = string(bt.Icon)
		} else {
			ch.CurValLabel = ToLabel(it)
			if len(ch.Tooltips) > i {
				bt.Tooltip = ch.Tooltips[i]
			}
		}
		bt.Data = i // index is the data
		bt.SetSelected(i == ch.CurIndex)
		idx := i
		bt.OnClick(func(e events.Event) {
			ch.SelectItemAction(idx)
		})
	}
}

func (ch *Chooser) HandleChooserEvents() {
	ch.HandleWidgetEvents()
	ch.HandleClickMenu()
	ch.HandleChooserKeys()
}

func (ch *Chooser) HandleClickMenu() {
	ch.OnClick(func(e events.Event) {
		if ch.OpenMenu(e) {
			e.SetHandled()
		}
	})
}

// OpenMenu will open any menu associated with this element.
// Returns true if menu opened, false if not.
func (ch *Chooser) OpenMenu(e events.Event) bool {
	pos := ch.ContextMenuPos(e)
	if ch.Parts != nil {
		if indic := ch.Parts.ChildByName("indicator", 3); indic != nil {
			pos = indic.(Widget).ContextMenuPos(nil) // use the pos
		}
	}
	NewMenu(ch.MakeItemsMenu, ch.This().(Widget), pos).Run()
	return true
}

func (ch *Chooser) HandleChooserKeys() {
	ch.OnKeyChord(func(e events.Event) {
		if KeyEventTrace {
			fmt.Printf("Chooser KeyChordEvent: %v\n", ch.Path())
		}
		kf := KeyFun(e.KeyChord())
		switch {
		case kf == KeyFunMoveUp:
			e.SetHandled()
			if len(ch.Items) > 0 {
				idx := ch.CurIndex - 1
				if idx < 0 {
					idx += len(ch.Items)
				}
				ch.SelectItemAction(idx)
			}
		case kf == KeyFunMoveDown:
			e.SetHandled()
			if len(ch.Items) > 0 {
				idx := ch.CurIndex + 1
				if idx >= len(ch.Items) {
					idx -= len(ch.Items)
				}
				ch.SelectItemAction(idx)
			}
		case kf == KeyFunPageUp:
			e.SetHandled()
			if len(ch.Items) > 10 {
				idx := ch.CurIndex - 10
				for idx < 0 {
					idx += len(ch.Items)
				}
				ch.SelectItemAction(idx)
			}
		case kf == KeyFunPageDown:
			e.SetHandled()
			if len(ch.Items) > 10 {
				idx := ch.CurIndex + 10
				for idx >= len(ch.Items) {
					idx -= len(ch.Items)
				}
				ch.SelectItemAction(idx)
			}
		case kf == KeyFunEnter || (!ch.Editable && e.KeyRune() == ' '):
			// if !(kt.Rune == ' ' && chb.Sc.Type == ScCompleter) {
			e.SetHandled()
			ch.Send(events.Click, e)
		// }
		default:
			tf, ok := ch.TextField()
			if !ok {
				break
			}
			// if we start typing, we focus the text field
			if !tf.StateIs(states.Focused) {
				tf.GrabFocus()
				tf.Send(events.Focus, e)
				tf.Send(events.KeyChord, e)
			}
		}
	})
}

func (ch *Chooser) TextFieldHandlers(tf *TextField) {
	tf.OnChange(func(e events.Event) {
		text := tf.Text()
		for idx, item := range ch.Items {
			if text == ToLabel(item) {
				ch.SetCurIndex(idx)
				ch.SendChange(e)
				return
			}
		}
		if !ch.AllowNew {
			// TODO: use validation
			slog.Error("invalid Chooser value", "value", text)
			return
		}
		ch.Items = append(ch.Items, text)
		ch.SetCurIndex(len(ch.Items) - 1)
		ch.SendChange(e)
	})
}

// CompleteMatch is the [complete.MatchFunc] used for the
// editable textfield part of the Chooser (if it exists).
func (ch *Chooser) CompleteMatch(data any, text string, posLn, posCh int) (md complete.Matches) {
	md.Seed = text
	comps := make(complete.Completions, len(ch.Items))
	for idx, item := range ch.Items {
		tooltip := ""
		if len(ch.Tooltips) > idx {
			tooltip = ch.Tooltips[idx]
		}
		comps[idx] = complete.Completion{
			Text: ToLabel(item),
			Desc: tooltip,
		}
	}
	md.Matches = complete.MatchSeedCompletion(comps, md.Seed)
	return md
}

// CompleteEdit is the [complete.EditFunc] used for the
// editable textfield part of the Chooser (if it exists).
func (ch *Chooser) CompleteEdit(data any, text string, cursorPos int, completion complete.Completion, seed string) (ed complete.Edit) {
	return complete.Edit{
		NewText:       completion.Text,
		ForwardDelete: len([]rune(text)),
	}
}
