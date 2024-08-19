// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package styles

import (
	"cogentcore.org/core/math32"
	"golang.org/x/image/font"
)

// FontFace is our enhanced Font Face structure which contains the enhanced computed
// metrics in addition to the font.Face face
type FontFace struct { //types:add

	// The full FaceName that the font is accessed by
	Name string

	// The integer font size in raw dots
	Size int

	// The system image.Font font rendering interface
	Face font.Face

	// enhanced metric information for the font
	Metrics FontMetrics
}

// NewFontFace returns a new font face
func NewFontFace(nm string, sz int, face font.Face) *FontFace {
	ff := &FontFace{Name: nm, Size: sz, Face: face}
	ff.ComputeMetrics()
	return ff
}

// FontMetrics are our enhanced dot-scale font metrics compared to what is available in
// the standard font.Metrics lib, including Ex and Ch being defined in terms of
// the actual letter x and 0
type FontMetrics struct { //types:add

	// reference 1.0 spacing line height of font in dots -- computed from font as ascent + descent + lineGap, where lineGap is specified by the font as the recommended line spacing
	Height float32

	// Em size of font -- this is NOT actually the width of the letter M, but rather the specified point size of the font (in actual display dots, not points) -- it does NOT include the descender and will not fit the entire height of the font
	Em float32

	// Ex size of font -- this is the actual height of the letter x in the font
	Ex float32

	// Ch size of font -- this is the actual width of the 0 glyph in the font
	Ch float32
}

// ComputeMetrics computes the Height, Em, Ex, Ch and Rem metrics associated
// with current font and overall units context
func (fs *FontFace) ComputeMetrics() {
	// apd := fs.Face.Metrics().Ascent + fs.Face.Metrics().Descent
	fmet := fs.Face.Metrics()
	fs.Metrics.Height = math32.Ceil(math32.FromFixed(fmet.Height))
	fs.Metrics.Em = float32(fs.Size) // conventional definition
	xb, _, ok := fs.Face.GlyphBounds('x')
	if ok {
		fs.Metrics.Ex = math32.FromFixed(xb.Max.Y - xb.Min.Y)
		// note: metric.Ex is typically 0?
		// if fs.Metrics.Ex != metex {
		// 	fmt.Printf("computed Ex: %v  metric ex: %v\n", fs.Metrics.Ex, metex)
		// }
	} else {
		metex := math32.FromFixed(fmet.XHeight)
		if metex != 0 {
			fs.Metrics.Ex = metex
		} else {
			fs.Metrics.Ex = 0.5 * fs.Metrics.Em
		}
	}
	xb, _, ok = fs.Face.GlyphBounds('0')
	if ok {
		fs.Metrics.Ch = math32.FromFixed(xb.Max.X - xb.Min.X)
	} else {
		fs.Metrics.Ch = 0.5 * fs.Metrics.Em
	}
}
