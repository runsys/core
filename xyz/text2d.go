// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xyz

import (
	"fmt"
	"image"
	"image/draw"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/gpu/phong"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
)

// Text2D presents 2D rendered text on a vertically oriented plane, using a texture.
// Call SetText() which calls RenderText to update fortext changes (re-renders texture).
// The native scale is such that a unit height value is the height of the default font
// set by the font-size property, and the X axis is scaled proportionally based on the
// rendered text size to maintain the aspect ratio.  Further scaling can be applied on
// top of that by setting the Pose.Scale values as usual.
// Standard styling properties can be set on the node to set font size, family,
// and text alignment relative to the Pose.Pos position (e.g., Left, Top puts the
// upper-left corner of text at Pos).
// Note that higher quality is achieved by using a larger font size (36 default).
// The margin property creates blank margin of the background color around the text
// (2 px default) and the background-color defaults to transparent
// but can be set to any color.
type Text2D struct {
	Solid

	// the text string to display
	Text string

	// styling settings for the text
	Styles styles.Style `set:"-" json:"-" xml:"-"`

	// position offset of start of text rendering relative to upper-left corner
	TextPos math32.Vector2 `set:"-" xml:"-" json:"-"`

	// render data for text label
	TextRender paint.Text `set:"-" xml:"-" json:"-"`

	// render state for rendering text
	RenderState paint.State `set:"-" copier:"-" json:"-" xml:"-" display:"-"`

	// automatically set to true if the font render color is the default
	// colors.Scheme.OnSurface.  If so, it is automatically updated if the default
	// changes, e.g., in light mode vs dark mode switching.
	usesDefaultColor bool
}

func (txt *Text2D) Init() {
	txt.Defaults()
}

func (txt *Text2D) Defaults() {
	txt.Solid.Defaults()
	txt.Pose.Scale.SetScalar(.005)
	txt.Pose.RotateOnAxis(0, 1, 0, 180)
	txt.Styles.Defaults()
	txt.Styles.Font.Size.Pt(36)
	txt.Styles.Margin.Set(units.Px(2))
	txt.Material.Bright = 4 // this is key for making e.g., a white background show up as white..
}

// TextSize returns the size of the text plane, applying all *local* scaling factors
// if nothing rendered yet, returns false
func (txt *Text2D) TextSize() (math32.Vector2, bool) {
	txt.Pose.Defaults() // only if nil
	sz := math32.Vector2{}
	tx := txt.Material.Texture
	if tx == nil {
		return sz, false
	}
	tsz := tx.Image().Bounds().Size()
	fsz := float32(txt.Styles.Font.Size.Dots)
	if fsz == 0 {
		fsz = 36
	}
	sz.Set(txt.Pose.Scale.X*float32(tsz.X)/fsz, txt.Pose.Scale.Y*float32(tsz.Y)/fsz)
	return sz, true
}

func (txt *Text2D) Config() {
	tm := txt.Scene.PlaneMesh2D()
	txt.SetMesh(tm)
	txt.Solid.Config()
	txt.RenderText()
	txt.Validate()
}

func (txt *Text2D) RenderText() {
	// TODO(kai): do we need to set unit context sizes? (units.Context.SetSizes)
	st := &txt.Styles
	fr := st.FontRender()
	if fr.Color == colors.Scheme.OnSurface {
		txt.usesDefaultColor = true
	}
	if txt.usesDefaultColor {
		fr.Color = colors.Scheme.OnSurface
	}
	if st.Font.Face == nil {
		st.Font = paint.OpenFont(fr, &st.UnitContext)
	}
	st.ToDots()

	txt.TextRender.SetHTML(txt.Text, fr, &txt.Styles.Text, &txt.Styles.UnitContext, nil)
	sz := txt.TextRender.BBox.Size()
	txt.TextRender.LayoutStdLR(&txt.Styles.Text, fr, &txt.Styles.UnitContext, sz)
	if txt.TextRender.BBox.Size() != sz {
		sz = txt.TextRender.BBox.Size()
		txt.TextRender.LayoutStdLR(&txt.Styles.Text, fr, &txt.Styles.UnitContext, sz)
		if txt.TextRender.BBox.Size() != sz {
			sz = txt.TextRender.BBox.Size()
		}
	}
	marg := txt.Styles.TotalMargin()
	sz.SetAdd(marg.Size())
	txt.TextPos = marg.Pos().Round()
	szpt := sz.ToPointRound()
	if szpt == (image.Point{}) {
		szpt = image.Point{10, 10}
	}
	bounds := image.Rectangle{Max: szpt}
	var img *image.RGBA
	var tx Texture
	var err error
	if txt.Material.Texture == nil {
		txname := "__Text2D_" + txt.Name
		tx, err = txt.Scene.TextureByNameTry(txname)
		if err != nil {
			tx = &TextureBase{Name: txname}
			img = image.NewRGBA(bounds)
			tx.AsTextureBase().RGBA = img
			txt.Scene.SetTexture(tx)
			txt.Material.SetTexture(tx)
		} else {
			if gpu.Debug {
				fmt.Printf("xyz.Text2D: error: texture name conflict: %s\n", txname)
			}
			txt.Material.SetTexture(tx)
			img = tx.Image()
		}
	} else {
		tx = txt.Material.Texture
		img = tx.Image()
		if img.Bounds() != bounds {
			img = image.NewRGBA(bounds)
		}
		tx.AsTextureBase().RGBA = img
		txt.Scene.Phong.SetTexture(tx.AsTextureBase().Name, phong.NewTexture(img))
	}
	rs := &txt.RenderState
	if rs.Image != img || rs.Image.Bounds() != img.Bounds() {
		rs.Init(szpt.X, szpt.Y, img)
	}
	rs.PushBounds(bounds)
	pt := styles.Paint{}
	pt.Defaults()
	pt.FromStyle(st)
	ctx := &paint.Context{State: rs, Paint: &pt}
	if st.Background != nil {
		draw.Draw(img, bounds, st.Background, image.Point{}, draw.Src)
	}
	txt.TextRender.Render(ctx, txt.TextPos)
	rs.PopBounds()
}

// Validate checks that text has valid mesh and texture settings, etc
func (txt *Text2D) Validate() error {
	// todo: validate more stuff here
	return txt.Solid.Validate()
}

func (txt *Text2D) UpdateWorldMatrix(parWorld *math32.Matrix4) {
	sz, ok := txt.TextSize()
	if ok {
		sc := math32.Vec3(sz.X, sz.Y, txt.Pose.Scale.Z)
		ax, ay := txt.Styles.Text.AlignFactors()
		al := txt.Styles.Text.AlignV
		switch al {
		case styles.Start:
			ay = -0.5
		case styles.Center:
			ay = 0
		case styles.End:
			ay = 0.5
		}
		ps := txt.Pose.Pos
		ps.X += (0.5 - ax) * sz.X
		ps.Y += ay * sz.Y
		txt.Pose.Matrix.SetTransform(ps, txt.Pose.Quat, sc)
	} else {
		txt.Pose.UpdateMatrix()
	}
	txt.Pose.UpdateWorldMatrix(parWorld)
}

func (txt *Text2D) IsTransparent() bool {
	return colors.ToUniform(txt.Styles.Background).A < 255
}

func (txt *Text2D) RenderClass() RenderClasses {
	if txt.IsTransparent() {
		return RClassTransTexture
	}
	return RClassOpaqueTexture
}
