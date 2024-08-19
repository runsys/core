// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colors

import (
	"image"
	"image/color"
)

// Uniform returns a new [image.Uniform] filled completely with the given color.
func Uniform(c color.Color) image.Image {
	return image.NewUniform(c)
}

// ToUniform converts the given image to a uniform [color.RGBA] color.
func ToUniform(img image.Image) color.RGBA {
	if img == nil {
		return color.RGBA{}
	}
	return AsRGBA(img.At(0, 0))
}

// Pattern returns a new unbounded [image.Image] represented by the given pattern function.
func Pattern(f func(x, y int) color.Color) image.Image {
	return &pattern{f}
}

type pattern struct {
	f func(x, y int) color.Color
}

func (p *pattern) ColorModel() color.Model {
	return color.RGBAModel
}

func (p *pattern) Bounds() image.Rectangle {
	return image.Rect(-1e9, -1e9, 1e9, 1e9)
}

func (p *pattern) At(x, y int) color.Color {
	return p.f(x, y)
}
