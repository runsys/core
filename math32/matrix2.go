// Copyright (c) 2020, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math32

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"

	"cogentcore.org/core/base/errors"
	"golang.org/x/image/math/fixed"
)

/*
This is heavily modified from: https://github.com/fogleman/gg

Copyright (C) 2016 Michael Fogleman

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Matrix2 is a 3x2 matrix.
type Matrix2 struct {
	XX, YX, XY, YY, X0, Y0 float32
}

// Identity2 returns a new identity [Matrix2] matrix.
func Identity2() Matrix2 {
	return Matrix2{
		1, 0,
		0, 1,
		0, 0,
	}
}

func (m Matrix2) IsIdentity() bool {
	return m.XX == 1 && m.YX == 0 && m.XY == 0 && m.YY == 1 && m.X0 == 0 && m.Y0 == 0
}

// Translate2D returns a Matrix2 2D matrix with given translations
func Translate2D(x, y float32) Matrix2 {
	return Matrix2{
		1, 0,
		0, 1,
		x, y,
	}
}

// Scale2D returns a Matrix2 2D matrix with given scaling factors
func Scale2D(x, y float32) Matrix2 {
	return Matrix2{
		x, 0,
		0, y,
		0, 0,
	}
}

// Rotate2D returns a Matrix2 2D matrix with given rotation, specified in radians
func Rotate2D(angle float32) Matrix2 {
	c := float32(Cos(angle))
	s := float32(Sin(angle))
	return Matrix2{
		c, s,
		-s, c,
		0, 0,
	}
}

// Shear2D returns a Matrix2 2D matrix with given shearing
func Shear2D(x, y float32) Matrix2 {
	return Matrix2{
		1, y,
		x, 1,
		0, 0,
	}
}

// Skew2D returns a Matrix2 2D matrix with given skewing
func Skew2D(x, y float32) Matrix2 {
	return Matrix2{
		1, Tan(y),
		Tan(x), 1,
		0, 0,
	}
}

// Mul returns a*b
func (a Matrix2) Mul(b Matrix2) Matrix2 {
	return Matrix2{
		XX: a.XX*b.XX + a.XY*b.YX,
		YX: a.YX*b.XX + a.YY*b.YX,
		XY: a.XX*b.XY + a.XY*b.YY,
		YY: a.YX*b.XY + a.YY*b.YY,
		X0: a.XX*b.X0 + a.XY*b.Y0 + a.X0,
		Y0: a.YX*b.X0 + a.YY*b.Y0 + a.Y0,
	}
}

// SetMul sets a to a*b
func (a *Matrix2) SetMul(b Matrix2) {
	*a = a.Mul(b)
}

// MulVector2AsVector multiplies the Vector2 as a vector without adding translations.
// This is for directional vectors and not points.
func (a Matrix2) MulVector2AsVector(v Vector2) Vector2 {
	tx := a.XX*v.X + a.XY*v.Y
	ty := a.YX*v.X + a.YY*v.Y
	return Vec2(tx, ty)
}

// MulVector2AsPoint multiplies the Vector2 as a point, including adding translations.
func (a Matrix2) MulVector2AsPoint(v Vector2) Vector2 {
	tx := a.XX*v.X + a.XY*v.Y + a.X0
	ty := a.YX*v.X + a.YY*v.Y + a.Y0
	return Vec2(tx, ty)
}

// MulVector2AsPointCenter multiplies the Vector2 as a point relative to given center-point
// including adding translations.
func (a Matrix2) MulVector2AsPointCenter(v, ctr Vector2) Vector2 {
	rel := v.Sub(ctr)
	tx := ctr.X + a.XX*rel.X + a.XY*rel.Y + a.X0
	ty := ctr.Y + a.YX*rel.X + a.YY*rel.Y + a.Y0
	return Vec2(tx, ty)
}

// MulCenter multiplies the Matrix2, first subtracting given translation center point
// from the translation components, and then adding it back in.
func (a Matrix2) MulCenter(b Matrix2, ctr Vector2) Matrix2 {
	a.X0 -= ctr.X
	a.Y0 -= ctr.Y
	rv := a.Mul(b)
	rv.X0 += ctr.X
	rv.Y0 += ctr.Y
	return rv
}

// SetMulCenter sets the matrix to the result of [Matrix2.MulCenter].
func (a *Matrix2) SetMulCenter(b Matrix2, ctr Vector2) {
	*a = a.MulCenter(b, ctr)
}

// MulFixedAsPoint multiplies the fixed point as a point, including adding translations.
func (a Matrix2) MulFixedAsPoint(fp fixed.Point26_6) fixed.Point26_6 {
	x := fixed.Int26_6((float32(fp.X)*a.XX + float32(fp.Y)*a.XY) + a.X0*32)
	y := fixed.Int26_6((float32(fp.X)*a.YX + float32(fp.Y)*a.YY) + a.Y0*32)
	return fixed.Point26_6{x, y}
}

func (a Matrix2) Translate(x, y float32) Matrix2 {
	return a.Mul(Translate2D(x, y))
}

func (a Matrix2) Scale(x, y float32) Matrix2 {
	return a.Mul(Scale2D(x, y))
}

func (a Matrix2) Rotate(angle float32) Matrix2 {
	return a.Mul(Rotate2D(angle))
}

func (a Matrix2) Shear(x, y float32) Matrix2 {
	return a.Mul(Shear2D(x, y))
}

func (a Matrix2) Skew(x, y float32) Matrix2 {
	return a.Mul(Skew2D(x, y))
}

// ExtractRot extracts the rotation component from a given matrix
func (a Matrix2) ExtractRot() float32 {
	return Atan2(-a.XY, a.XX)
}

// ExtractXYScale extracts the X and Y scale factors after undoing any
// rotation present -- i.e., in the original X, Y coordinates
func (a Matrix2) ExtractScale() (scx, scy float32) {
	rot := a.ExtractRot()
	tx := a.Rotate(-rot)
	scxv := tx.MulVector2AsVector(Vec2(1, 0))
	scyv := tx.MulVector2AsVector(Vec2(0, 1))
	return scxv.X, scyv.Y
}

// Inverse returns inverse of matrix, for inverting transforms
func (a Matrix2) Inverse() Matrix2 {
	// homogenous rep, rc indexes, mapping into Matrix3 code
	// XX YX X0   n11 n12 n13    a b x
	// XY YY Y0   n21 n22 n23    c d y
	// 0  0  1    n31 n32 n33    0 0 1

	// t11 := a.YY
	// t12 := -a.YX
	// t13 := a.Y0*a.YX - a.YY*a.X0
	det := a.XX*a.YY - a.XY*a.YX // ad - bc
	detInv := 1 / det

	b := Matrix2{}
	b.XX = a.YY * detInv  // a = d
	b.XY = -a.XY * detInv // c = -c
	b.YX = -a.YX * detInv // b = -b
	b.YY = a.XX * detInv  // d = a
	b.X0 = (a.Y0*a.XY - a.YY*a.X0) * detInv
	b.Y0 = (a.X0*a.YX - a.XX*a.Y0) * detInv
	return b
}

// ParseFloat32 logs any strconv.ParseFloat errors
func ParseFloat32(pstr string) (float32, error) {
	r, err := strconv.ParseFloat(pstr, 32)
	if err != nil {
		log.Printf("core.ParseFloat32: error parsing float32 number from: %v, %v\n", pstr, err)
		return float32(0.0), err
	}
	return float32(r), nil
}

// ParseAngle32 returns radians angle from string that can specify units (deg,
// grad, rad) -- deg is assumed if not specified
func ParseAngle32(pstr string) (float32, error) {
	units := "deg"
	lstr := strings.ToLower(pstr)
	if strings.Contains(lstr, "deg") {
		units = "deg"
		lstr = strings.TrimSuffix(lstr, "deg")
	} else if strings.Contains(lstr, "grad") {
		units = "grad"
		lstr = strings.TrimSuffix(lstr, "grad")
	} else if strings.Contains(lstr, "rad") {
		units = "rad"
		lstr = strings.TrimSuffix(lstr, "rad")
	}
	r, err := strconv.ParseFloat(lstr, 32)
	if err != nil {
		log.Printf("core.ParseAngle32: error parsing float32 number from: %v, %v\n", lstr, err)
		return float32(0.0), err
	}
	switch units {
	case "deg":
		return DegToRad(float32(r)), nil
	case "grad":
		return float32(r) * Pi / 200, nil
	case "rad":
		return float32(r), nil
	}
	return float32(r), nil
}

// ReadPoints reads a set of floating point values from a SVG format number
// string -- returns a slice or nil if there was an error
func ReadPoints(pstr string) []float32 {
	lastIndex := -1
	var pts []float32
	lr := ' '
	for i, r := range pstr {
		if !unicode.IsNumber(r) && r != '.' && !(r == '-' && lr == 'e') && r != 'e' {
			if lastIndex != -1 {
				s := pstr[lastIndex:i]
				p, err := ParseFloat32(s)
				if err != nil {
					return nil
				}
				pts = append(pts, p)
			}
			if r == '-' {
				lastIndex = i
			} else {
				lastIndex = -1
			}
		} else if lastIndex == -1 {
			lastIndex = i
		}
		lr = r
	}
	if lastIndex != -1 && lastIndex != len(pstr) {
		s := pstr[lastIndex:]
		p, err := ParseFloat32(s)
		if err != nil {
			return nil
		}
		pts = append(pts, p)
	}
	return pts
}

// PointsCheckN checks the number of points read and emits an error if not equal to n
func PointsCheckN(pts []float32, n int, errmsg string) error {
	if len(pts) != n {
		return fmt.Errorf("%v incorrect number of points: %v != %v", errmsg, len(pts), n)
	}
	return nil
}

// SetString processes the standard SVG-style transform strings
func (a *Matrix2) SetString(str string) error {
	errmsg := "math32.Matrix2.SetString:"
	str = strings.ToLower(strings.TrimSpace(str))
	*a = Identity2()
	if str == "none" {
		*a = Identity2()
		return nil
	}
	// could have multiple transforms
	for {
		pidx := strings.IndexByte(str, '(')
		if pidx < 0 {
			err := fmt.Errorf("%s no params for transform: %v", errmsg, str)
			return errors.Log(err)
		}
		cmd := str[:pidx]
		vals := str[pidx+1:]
		nxt := ""
		eidx := strings.IndexByte(vals, ')')
		if eidx > 0 {
			nxt = strings.TrimSpace(vals[eidx+1:])
			if strings.HasPrefix(nxt, ";") {
				nxt = strings.TrimSpace(strings.TrimPrefix(nxt, ";"))
			}
			vals = vals[:eidx]
		}
		pts := ReadPoints(vals)
		switch cmd {
		case "matrix":
			if err := PointsCheckN(pts, 6, errmsg); err != nil {
				errors.Log(err)
			} else {
				*a = Matrix2{pts[0], pts[1], pts[2], pts[3], pts[4], pts[5]}
			}
		case "translate":
			if len(pts) == 1 {
				*a = a.Translate(pts[0], 0)
			} else if len(pts) == 2 {
				*a = a.Translate(pts[0], pts[1])
			} else {
				errors.Log(PointsCheckN(pts, 2, errmsg))
			}
		case "translatex":
			if err := PointsCheckN(pts, 1, errmsg); err != nil {
				errors.Log(err)
			} else {
				*a = a.Translate(pts[0], 0)
			}
		case "translatey":
			if err := PointsCheckN(pts, 1, errmsg); err != nil {
				errors.Log(err)
			} else {
				*a = a.Translate(0, pts[0])
			}
		case "scale":
			if len(pts) == 1 {
				*a = a.Scale(pts[0], pts[0])
			} else if len(pts) == 2 {
				*a = a.Scale(pts[0], pts[1])
			} else {
				err := fmt.Errorf("%v incorrect number of points: 2 != %v", errmsg, len(pts))
				errors.Log(err)
			}
		case "scalex":
			if err := PointsCheckN(pts, 1, errmsg); err != nil {
				errors.Log(err)
			} else {
				*a = a.Scale(pts[0], 1)
			}
		case "scaley":
			if err := PointsCheckN(pts, 1, errmsg); err != nil {
				errors.Log(err)
			} else {
				*a = a.Scale(1, pts[0])
			}
		case "rotate":
			ang := DegToRad(pts[0]) // always in degrees in this form
			if len(pts) == 3 {
				*a = a.Translate(pts[1], pts[2]).Rotate(ang).Translate(-pts[1], -pts[2])
			} else if len(pts) == 1 {
				*a = a.Rotate(ang)
			} else {
				errors.Log(PointsCheckN(pts, 1, errmsg))
			}
		case "skew":
			if err := PointsCheckN(pts, 2, errmsg); err != nil {
				errors.Log(err)
			} else {
				*a = a.Skew(pts[0], pts[1])
			}
		case "skewx":
			if err := PointsCheckN(pts, 1, errmsg); err != nil {
				errors.Log(err)
			} else {
				*a = a.Skew(pts[0], 0)
			}
		case "skewy":
			if err := PointsCheckN(pts, 1, errmsg); err != nil {
				errors.Log(err)
			} else {
				*a = a.Skew(0, pts[0])
			}
		default:
			return fmt.Errorf("unknown command %q", cmd)
		}
		if nxt == "" {
			break
		}
		if !strings.Contains(nxt, "(") {
			break
		}
		str = nxt
	}
	return nil
}

// String returns the XML-based string representation of the transform
func (a *Matrix2) String() string {
	if a.IsIdentity() {
		return "none"
	}
	if a.YX == 0 && a.XY == 0 { // no rotation, emit scale and translate
		str := ""
		if a.X0 != 0 || a.Y0 != 0 {
			str += fmt.Sprintf("translate(%g,%g)", a.X0, a.Y0)
		}
		if a.XX != 1 || a.YY != 1 {
			if str != "" {
				str += " "
			}
			str += fmt.Sprintf("scale(%g,%g)", a.XX, a.YY)
		}
		return str
	}
	// just report the whole matrix
	return fmt.Sprintf("matrix(%g,%g,%g,%g,%g,%g)", a.XX, a.YX, a.XY, a.YY, a.X0, a.Y0)
}
