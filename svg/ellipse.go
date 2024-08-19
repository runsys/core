// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package svg

import (
	"cogentcore.org/core/math32"
)

// Ellipse is a SVG ellipse
type Ellipse struct {
	NodeBase

	// position of the center of the ellipse
	Pos math32.Vector2 `xml:"{cx,cy}"`

	// radii of the ellipse in the horizontal, vertical axes
	Radii math32.Vector2 `xml:"{rx,ry}"`
}

func (g *Ellipse) SVGName() string { return "ellipse" }

func (g *Ellipse) Init() {
	g.NodeBase.Init()
	g.Radii.Set(1, 1)
}

func (g *Ellipse) SetNodePos(pos math32.Vector2) {
	g.Pos = pos.Sub(g.Radii)
}

func (g *Ellipse) SetNodeSize(sz math32.Vector2) {
	g.Radii = sz.MulScalar(0.5)
}

func (g *Ellipse) LocalBBox() math32.Box2 {
	bb := math32.Box2{}
	hlw := 0.5 * g.LocalLineWidth()
	bb.Min = g.Pos.Sub(g.Radii.AddScalar(hlw))
	bb.Max = g.Pos.Add(g.Radii.AddScalar(hlw))
	return bb
}

func (g *Ellipse) Render(sv *SVG) {
	vis, pc := g.PushTransform(sv)
	if !vis {
		return
	}
	pc.DrawEllipse(g.Pos.X, g.Pos.Y, g.Radii.X, g.Radii.Y)
	pc.FillStrokeClear()

	g.BBoxes(sv)
	g.RenderChildren(sv)

	pc.PopTransform()
}

// ApplyTransform applies the given 2D transform to the geometry of this node
// each node must define this for itself
func (g *Ellipse) ApplyTransform(sv *SVG, xf math32.Matrix2) {
	rot := xf.ExtractRot()
	if rot != 0 || !g.Paint.Transform.IsIdentity() {
		g.Paint.Transform.SetMul(xf)
		g.SetProperty("transform", g.Paint.Transform.String())
	} else {
		g.Pos = xf.MulVector2AsPoint(g.Pos)
		g.Radii = xf.MulVector2AsVector(g.Radii)
		g.GradientApplyTransform(sv, xf)
	}
}

// ApplyDeltaTransform applies the given 2D delta transforms to the geometry of this node
// relative to given point.  Trans translation and point are in top-level coordinates,
// so must be transformed into local coords first.
// Point is upper left corner of selection box that anchors the translation and scaling,
// and for rotation it is the center point around which to rotate
func (g *Ellipse) ApplyDeltaTransform(sv *SVG, trans math32.Vector2, scale math32.Vector2, rot float32, pt math32.Vector2) {
	crot := g.Paint.Transform.ExtractRot()
	if rot != 0 || crot != 0 {
		xf, lpt := g.DeltaTransform(trans, scale, rot, pt, false) // exclude self
		g.Paint.Transform.SetMulCenter(xf, lpt)
		g.SetProperty("transform", g.Paint.Transform.String())
	} else {
		xf, lpt := g.DeltaTransform(trans, scale, rot, pt, true) // include self
		g.Pos = xf.MulVector2AsPointCenter(g.Pos, lpt)
		g.Radii = xf.MulVector2AsVector(g.Radii)
		g.GradientApplyTransformPt(sv, xf, lpt)
	}
}

// WriteGeom writes the geometry of the node to a slice of floating point numbers
// the length and ordering of which is specific to each node type.
// Slice must be passed and will be resized if not the correct length.
func (g *Ellipse) WriteGeom(sv *SVG, dat *[]float32) {
	SetFloat32SliceLen(dat, 4+6)
	(*dat)[0] = g.Pos.X
	(*dat)[1] = g.Pos.Y
	(*dat)[2] = g.Radii.X
	(*dat)[3] = g.Radii.Y
	g.WriteTransform(*dat, 4)
	g.GradientWritePts(sv, dat)
}

// ReadGeom reads the geometry of the node from a slice of floating point numbers
// the length and ordering of which is specific to each node type.
func (g *Ellipse) ReadGeom(sv *SVG, dat []float32) {
	g.Pos.X = dat[0]
	g.Pos.Y = dat[1]
	g.Radii.X = dat[2]
	g.Radii.Y = dat[3]
	g.ReadTransform(dat, 4)
	g.GradientReadPts(sv, dat)
}
