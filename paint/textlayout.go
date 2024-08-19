// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paint

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"golang.org/x/image/font"
)

// RuneSpanPos returns the position (span, rune index within span) within a
// sequence of spans of a given absolute rune index, starting in the first
// span -- returns false if index is out of range (and returns the last position).
func (tx *Text) RuneSpanPos(idx int) (si, ri int, ok bool) {
	if idx < 0 || len(tx.Spans) == 0 {
		return 0, 0, false
	}
	ri = idx
	for si = range tx.Spans {
		if ri < 0 {
			ri = 0
		}
		sr := &tx.Spans[si]
		if ri >= len(sr.Render) {
			ri -= len(sr.Render)
			continue
		}
		return si, ri, true
	}
	si = len(tx.Spans) - 1
	ri = len(tx.Spans[si].Render)
	return si, ri, false
}

// SpanPosToRuneIndex returns the absolute rune index for a given span, rune
// index position -- i.e., the inverse of RuneSpanPos.  Returns false if given
// input position is out of range, and returns last valid index in that case.
func (tx *Text) SpanPosToRuneIndex(si, ri int) (idx int, ok bool) {
	idx = 0
	for i := range tx.Spans {
		sr := &tx.Spans[i]
		if si > i {
			idx += len(sr.Render)
			continue
		}
		if ri <= len(sr.Render) {
			return idx + ri, true
		}
		return idx + (len(sr.Render)), false
	}
	return 0, false
}

// RuneRelPos returns the relative (starting) position of the given rune
// index, counting progressively through all spans present (adds Span RelPos
// and rune RelPos) -- this is typically the baseline position where rendering
// will start, not the upper left corner. If index > length, then uses
// LastPos.  Returns also the index of the span that holds that char (-1 = no
// spans at all) and the rune index within that span, and false if index is
// out of range.
func (tx *Text) RuneRelPos(idx int) (pos math32.Vector2, si, ri int, ok bool) {
	si, ri, ok = tx.RuneSpanPos(idx)
	if ok {
		sr := &tx.Spans[si]
		return sr.RelPos.Add(sr.Render[ri].RelPos), si, ri, true
	}
	nsp := len(tx.Spans)
	if nsp > 0 {
		sr := &tx.Spans[nsp-1]
		return sr.LastPos, nsp - 1, len(sr.Render), false
	}
	return math32.Vector2{}, -1, -1, false
}

// RuneEndPos returns the relative ending position of the given rune index,
// counting progressively through all spans present(adds Span RelPos and rune
// RelPos + rune Size.X for LR writing). If index > length, then uses LastPos.
// Returns also the index of the span that holds that char (-1 = no spans at
// all) and the rune index within that span, and false if index is out of
// range.
func (tx *Text) RuneEndPos(idx int) (pos math32.Vector2, si, ri int, ok bool) {
	si, ri, ok = tx.RuneSpanPos(idx)
	if ok {
		sr := &tx.Spans[si]
		spos := sr.RelPos.Add(sr.Render[ri].RelPos)
		spos.X += sr.Render[ri].Size.X
		return spos, si, ri, true
	}
	nsp := len(tx.Spans)
	if nsp > 0 {
		sr := &tx.Spans[nsp-1]
		return sr.LastPos, nsp - 1, len(sr.Render), false
	}
	return math32.Vector2{}, -1, -1, false
}

// PosToRune returns the rune span and rune indexes for given relative X,Y
// pixel position, if the pixel position lies within the given text area.
// If not, returns false.  It is robust to left-right out-of-range positions,
// returning the first or last rune index respectively.
func (tx *Text) PosToRune(pos math32.Vector2) (si, ri int, ok bool) {
	ok = false
	if pos.X < 0 || pos.Y < 0 { // note: don't bail on X yet
		return
	}
	sz := tx.BBox.Size()
	if pos.Y >= sz.Y {
		si = len(tx.Spans) - 1
		sr := tx.Spans[si]
		ri = len(sr.Render)
		ok = true
		return
	}
	if len(tx.Spans) == 0 {
		ok = true
		return
	}
	yoff := tx.Spans[0].RelPos.Y // baseline offset applied to everything
	for li, sr := range tx.Spans {
		st := sr.RelPos
		st.Y -= yoff
		lp := sr.LastPos
		lp.Y += tx.LineHeight - yoff // todo: only for LR
		b := math32.Box2{Min: st, Max: lp}
		nr := len(sr.Render)
		if !b.ContainsPoint(pos) {
			if pos.Y >= st.Y && pos.Y < lp.Y {
				if pos.X < st.X {
					return li, 0, true
				}
				return li, nr + 1, true
			}
			continue
		}
		for j := range sr.Render {
			r := &sr.Render[j]
			sz := r.Size
			sz.Y = tx.LineHeight // todo: only LR
			if j < nr-1 {
				nxt := &sr.Render[j+1]
				sz.X = nxt.RelPos.X - r.RelPos.X
			}
			ep := st.Add(sz)
			b := math32.Box2{Min: st, Max: ep}
			if b.ContainsPoint(pos) {
				return li, j, true
			}
			st.X += sz.X // todo: only LR
		}
	}
	return 0, 0, false
}

//////////////////////////////////////////////////////////////////////////////////
//  TextStyle-based Layout Routines

// Layout does basic standard layout of text using Text style parameters, assigning
// relative positions to spans and runes according to given styles, and given
// size overall box.  Nonzero values used to constrain, with the width used as a
// hard constraint to drive word wrapping (if a word wrap style is present).
// Returns total resulting size box for text, which can be larger than the given
// size, if the text requires more size to fit everything.
// Font face in styles.Font is used for determining line spacing here.
// Other versions can do more expensive calculations of variable line spacing as needed.
func (tr *Text) Layout(txtSty *styles.Text, fontSty *styles.FontRender, ctxt *units.Context, size math32.Vector2) math32.Vector2 {
	// todo: switch on layout types once others are supported
	return tr.LayoutStdLR(txtSty, fontSty, ctxt, size)
}

// LayoutStdLR does basic standard layout of text in LR direction.
func (tr *Text) LayoutStdLR(txtSty *styles.Text, fontSty *styles.FontRender, ctxt *units.Context, size math32.Vector2) math32.Vector2 {
	if len(tr.Spans) == 0 {
		return math32.Vector2{}
	}

	// pr := profile.Start("TextLayout")
	// defer pr.End()
	//
	tr.Dir = styles.LRTB
	fontSty.Font = OpenFont(fontSty, ctxt)
	fht := fontSty.Face.Metrics.Height
	tr.FontHeight = fht
	dsc := math32.FromFixed(fontSty.Face.Face.Metrics().Descent)
	lspc := txtSty.EffLineHeight(fht)
	tr.LineHeight = lspc
	lpad := (lspc - fht) / 2 // padding above / below text box for centering in line

	maxw := float32(0)

	// first pass gets rune positions and wraps text as needed, and gets max width
	si := 0
	for si < len(tr.Spans) {
		sr := &(tr.Spans[si])
		if err := sr.IsValid(); err != nil {
			si++
			continue
		}
		if sr.LastPos.X == 0 { // don't re-do unless necessary
			sr.SetRunePosLR(txtSty.LetterSpacing.Dots, txtSty.WordSpacing.Dots, fontSty.Face.Metrics.Ch, txtSty.TabSize)
		}
		if sr.IsNewPara() {
			sr.RelPos.X = txtSty.Indent.Dots
		} else {
			sr.RelPos.X = 0
		}
		ssz := sr.SizeHV()
		ssz.X += sr.RelPos.X
		if size.X > 0 && ssz.X > size.X && txtSty.HasWordWrap() {
			for {
				wp := sr.FindWrapPosLR(size.X, ssz.X)
				if wp > 0 && wp < len(sr.Text)-1 {
					nsr := sr.SplitAtLR(wp)
					tr.InsertSpan(si+1, nsr)
					ssz = sr.SizeHV()
					ssz.X += sr.RelPos.X
					if ssz.X > maxw {
						maxw = ssz.X
					}
					si++
					if si >= len(tr.Spans) {
						break
					}
					sr = &(tr.Spans[si]) // keep going with nsr
					sr.SetRunePosLR(txtSty.LetterSpacing.Dots, txtSty.WordSpacing.Dots, fontSty.Face.Metrics.Ch, txtSty.TabSize)
					ssz = sr.SizeHV()

					// fixup links
					for li := range tr.Links {
						tl := &tr.Links[li]
						if tl.StartSpan == si-1 {
							if tl.StartIndex >= wp {
								tl.StartIndex -= wp
								tl.StartSpan++
							}
						} else if tl.StartSpan > si-1 {
							tl.StartSpan++
						}
						if tl.EndSpan == si-1 {
							if tl.EndIndex >= wp {
								tl.EndIndex -= wp
								tl.EndSpan++
							}
						} else if tl.EndSpan > si-1 {
							tl.EndSpan++
						}
					}

					if ssz.X <= size.X {
						if ssz.X > maxw {
							maxw = ssz.X
						}
						break
					}
				} else {
					if ssz.X > maxw {
						maxw = ssz.X
					}
					break
				}
			}
		} else {
			if ssz.X > maxw {
				maxw = ssz.X
			}
		}
		si++
	}
	// have maxw, can do alignment cases..

	// make sure links are still in range
	for li := range tr.Links {
		tl := &tr.Links[li]
		stsp := tr.Spans[tl.StartSpan]
		if tl.StartIndex >= len(stsp.Text) {
			tl.StartIndex = len(stsp.Text) - 1
		}
		edsp := tr.Spans[tl.EndSpan]
		if tl.EndIndex >= len(edsp.Text) {
			tl.EndIndex = len(edsp.Text) - 1
		}
	}

	if maxw > size.X {
		size.X = maxw
	}

	// vertical alignment
	nsp := len(tr.Spans)
	npara := 0
	for si := 1; si < nsp; si++ {
		sr := &(tr.Spans[si])
		if sr.IsNewPara() {
			npara++
		}
	}

	vht := lspc*float32(nsp) + float32(npara)*txtSty.ParaSpacing.Dots
	if vht > size.Y {
		size.Y = vht
	}
	tr.BBox.Min.SetZero()
	tr.BBox.Max = math32.Vec2(maxw, vht)

	vpad := float32(0) // padding at top to achieve vertical alignment
	vextra := size.Y - vht
	if vextra > 0 {
		switch txtSty.AlignV {
		case styles.Center:
			vpad = vextra / 2
		case styles.End:
			vpad = vextra
		}
	}

	vbaseoff := lspc - lpad - dsc // offset of baseline within overall line
	vpos := vpad + vbaseoff

	for si := range tr.Spans {
		sr := &(tr.Spans[si])
		if si > 0 && sr.IsNewPara() {
			vpos += txtSty.ParaSpacing.Dots
		}
		sr.RelPos.Y = vpos
		sr.LastPos.Y = vpos
		ssz := sr.SizeHV()
		ssz.X += sr.RelPos.X
		hextra := size.X - ssz.X
		if hextra > 0 {
			switch txtSty.Align {
			case styles.Center:
				sr.RelPos.X += hextra / 2
			case styles.End:
				sr.RelPos.X += hextra
			}
		}
		vpos += lspc
	}
	return size
}

// Transform applies given 2D transform matrix to the text character rotations,
// scaling, and positions, so that the text is rendered according to that transform.
// The fontSty is the font style used for specifying the font originally.
func (tr *Text) Transform(mat math32.Matrix2, fontSty *styles.FontRender, ctxt *units.Context) {
	orgsz := fontSty.Size
	tmpsty := styles.FontRender{}
	tmpsty = *fontSty
	rot := mat.ExtractRot()
	scx, scy := mat.ExtractScale()
	scalex := scx / scy
	if scalex == 1 {
		scalex = 0
	}
	for si := range tr.Spans {
		sr := &(tr.Spans[si])
		sr.RelPos = mat.MulVector2AsVector(sr.RelPos)
		sr.LastPos = mat.MulVector2AsVector(sr.LastPos)
		for i := range sr.Render {
			rn := &sr.Render[i]
			if rn.Face != nil {
				tmpsty.Size = units.Value{Value: orgsz.Value * scy, Unit: orgsz.Unit, Dots: orgsz.Dots * scy} // rescale by y
				tmpsty.Font = OpenFont(&tmpsty, ctxt)
				rn.Face = tmpsty.Face.Face
			}
			rn.RelPos = mat.MulVector2AsVector(rn.RelPos)
			rn.Size.Y *= scy
			rn.Size.X *= scx
			rn.RotRad = rot
			rn.ScaleX = scalex
		}
	}
	tr.BBox = tr.BBox.MulMatrix2(mat)
}

// UpdateBBox updates the overall text bounding box
// based on actual glyph bounding boxes.
func (tr *Text) UpdateBBox() {
	tr.BBox.SetEmpty()
	for si := range tr.Spans {
		sr := &(tr.Spans[si])
		var curfc font.Face
		for i := range sr.Render {
			r := sr.Text[i]
			rn := &sr.Render[i]
			if rn.Face != nil {
				curfc = rn.Face
			}
			gbf, _, ok := curfc.GlyphBounds(r)
			if ok {
				gb := math32.B2FromFixed(gbf)
				gb.Translate(rn.RelPos)
				tr.BBox.ExpandByBox(gb)
			}
		}
	}
}
