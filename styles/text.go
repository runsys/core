// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package styles

import (
	"cogentcore.org/core/styles/units"
)

// IMPORTANT: any changes here must be updated in style_properties.go StyleTextFuncs

// Text is used for layout-level (widget, html-style) text styling --
// FontStyle contains all the lower-level text rendering info used in SVG --
// most of these are inherited
type Text struct { //types:add

	// how to align text, horizontally (inherited).
	// This *only* applies to the text within its containing element,
	// and is typically relevant only for multi-line text:
	// for single-line text, if element does not have a specified size
	// that is different from the text size, then this has *no effect*.
	Align Aligns

	// vertical alignment of text (inherited).
	// This *only* applies to the text within its containing element:
	// if that element does not have a specified size
	// that is different from the text size, then this has *no effect*.
	AlignV Aligns

	// for svg rendering only (inherited):
	// determines the alignment relative to text position coordinate.
	// For RTL start is right, not left, and start is top for TB
	Anchor TextAnchors

	// spacing between characters and lines
	LetterSpacing units.Value

	// extra space to add between words (inherited)
	WordSpacing units.Value

	// specified height of a line of text (inherited); text is centered within the overall lineheight;
	// the standard way to specify line height is in terms of em
	LineHeight units.Value

	// WhiteSpace (not inherited) specifies how white space is processed,
	// and how lines are wrapped.  If set to WhiteSpaceNormal (default) lines are wrapped.
	// See info about interactions with Grow.X setting for this and the NoWrap case.
	WhiteSpace WhiteSpaces

	// determines how to treat unicode bidirectional information (inherited)
	UnicodeBidi UnicodeBidi

	// bidi-override or embed -- applies to all text elements (inherited)
	Direction TextDirections

	// overall writing mode -- only for text elements, not span (inherited)
	WritingMode TextDirections

	// for TBRL writing mode (only), determines orientation of alphabetic characters (inherited);
	// 90 is default (rotated); 0 means keep upright
	OrientationVert float32

	// for horizontal LR/RL writing mode (only), determines orientation of all characters (inherited);
	// 0 is default (upright)
	OrientationHoriz float32

	// how much to indent the first line in a paragraph (inherited)
	Indent units.Value

	// extra spacing between paragraphs (inherited); copied from [Style.Margin] per CSS spec
	// if that is non-zero, else can be set directly with para-spacing
	ParaSpacing units.Value

	// tab size, in number of characters (inherited)
	TabSize int
}

// LineHeightNormal represents a normal line height,
// equal to the default height of the font being used.
var LineHeightNormal = units.Dp(-1)

func (ts *Text) Defaults() {
	ts.LineHeight = LineHeightNormal
	ts.Align = Start
	ts.AlignV = Baseline
	ts.Direction = LTR
	ts.OrientationVert = 90
	ts.TabSize = 4
}

// ToDots runs ToDots on unit values, to compile down to raw pixels
func (ts *Text) ToDots(uc *units.Context) {
	ts.LetterSpacing.ToDots(uc)
	ts.WordSpacing.ToDots(uc)
	ts.LineHeight.ToDots(uc)
	ts.Indent.ToDots(uc)
	ts.ParaSpacing.ToDots(uc)
}

// InheritFields from parent
func (ts *Text) InheritFields(parent *Text) {
	ts.Align = parent.Align
	ts.AlignV = parent.AlignV
	ts.Anchor = parent.Anchor
	ts.WordSpacing = parent.WordSpacing
	ts.LineHeight = parent.LineHeight
	// ts.WhiteSpace = par.WhiteSpace // todo: we can't inherit this b/c label base default then gets overwritten
	ts.UnicodeBidi = parent.UnicodeBidi
	ts.Direction = parent.Direction
	ts.WritingMode = parent.WritingMode
	ts.OrientationVert = parent.OrientationVert
	ts.OrientationHoriz = parent.OrientationHoriz
	ts.Indent = parent.Indent
	ts.ParaSpacing = parent.ParaSpacing
	ts.TabSize = parent.TabSize
}

// EffLineHeight returns the effective line height for the given
// font height, handling the [LineHeightNormal] special case.
func (ts *Text) EffLineHeight(fontHeight float32) float32 {
	if ts.LineHeight.Value < 0 {
		return fontHeight
	}
	return ts.LineHeight.Dots
}

// AlignFactors gets basic text alignment factors
func (ts *Text) AlignFactors() (ax, ay float32) {
	ax = 0.0
	ay = 0.0
	hal := ts.Align
	switch hal {
	case Center:
		ax = 0.5 // todo: determine if font is horiz or vert..
	case End:
		ax = 1.0
	}
	val := ts.AlignV
	switch val {
	case Start:
		ay = 0.9 // todo: need to find out actual baseline
	case Center:
		ay = 0.45 // todo: determine if font is horiz or vert..
	case End:
		ay = -0.1 // todo: need actual baseline
	}
	return
}

// UnicodeBidi determines the type of bidirectional text support.
// See https://pkg.go.dev/golang.org/x/text/unicode/bidi.
type UnicodeBidi int32 //enums:enum -trim-prefix Bidi -transform kebab

const (
	BidiNormal UnicodeBidi = iota
	BidiEmbed
	BidiBidiOverride
)

// TextDirections are for direction of text writing, used in direction and writing-mode styles
type TextDirections int32 //enums:enum -transform lower

const (
	LRTB TextDirections = iota
	RLTB
	TBRL
	LR
	RL
	TB
	LTR
	RTL
)

// TextAnchors are for direction of text writing, used in direction and writing-mode styles
type TextAnchors int32 //enums:enum -trim-prefix Anchor -transform kebab

const (
	AnchorStart TextAnchors = iota
	AnchorMiddle
	AnchorEnd
)

// WhiteSpaces determine how white space is processed
type WhiteSpaces int32 //enums:enum -trim-prefix WhiteSpace

const (
	// WhiteSpaceNormal means that all white space is collapsed to a single
	// space, and text wraps when necessary.  To get full word wrapping to
	// expand to all available space, you also need to set GrowWrap = 1.
	// Use the SetTextWrap convenience method to set both.
	WhiteSpaceNormal WhiteSpaces = iota

	// WhiteSpaceNowrap means that sequences of whitespace will collapse into
	// a single whitespace. Text will never wrap to the next line except
	// if there is an explicit line break via a <br> tag.  In general you
	// also don't want simple non-wrapping text labels to Grow (GrowWrap = 0).
	// Use the SetTextWrap method to set both.
	WhiteSpaceNowrap

	// WhiteSpacePre means that whitespace is preserved. Text
	// will only wrap on line breaks. Acts like the <pre> tag in HTML.  This
	// invokes a different hand-written parser because the default Go
	// parser automatically throws away whitespace.
	WhiteSpacePre

	// WhiteSpacePreLine means that sequences of whitespace will collapse
	// into a single whitespace. Text will wrap when necessary, and on line
	// breaks
	WhiteSpacePreLine

	// WhiteSpacePreWrap means that whitespace is preserved.
	// Text will wrap when necessary, and on line breaks
	WhiteSpacePreWrap
)

// HasWordWrap returns true if current white space option supports word wrap
func (ts *Text) HasWordWrap() bool {
	switch ts.WhiteSpace {
	case WhiteSpaceNormal, WhiteSpacePreLine, WhiteSpacePreWrap:
		return true
	default:
		return false
	}
}

// HasPre returns true if current white space option preserves existing
// whitespace (or at least requires that parser in case of PreLine, which is
// intermediate)
func (ts *Text) HasPre() bool {
	switch ts.WhiteSpace {
	case WhiteSpaceNormal, WhiteSpaceNowrap:
		return false
	default:
		return true
	}
}
