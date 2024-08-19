// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"image"
	"path/filepath"
	"testing"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/stretchr/testify/assert"
	"golang.org/x/image/draw"
)

var testImagePath = Filename(filepath.Join("..", "docs", "image.png"))

func TestImage(t *testing.T) {
	b := NewBody()
	img := NewImage(b)
	assert.NoError(t, img.Open(testImagePath))
	b.AssertRender(t, "image/basic")
}

func TestImageSize(t *testing.T) {
	b := NewBody()
	img := NewImage(b)
	assert.NoError(t, img.Open(testImagePath))
	img.Styler(func(s *styles.Style) {
		s.Min.Set(units.Dp(256))
	})
	b.AssertRender(t, "image/size")
}

func TestImageDirect(t *testing.T) {
	b := NewBody()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(img, image.Rect(10, 5, 100, 90), colors.Scheme.Warn.Container, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(20, 20, 60, 50), colors.Scheme.Success.Base, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 70, 80, 100), colors.Scheme.Error.Base, image.Point{}, draw.Src)
	NewImage(b).SetImage(img)
	b.AssertRender(t, "image/direct")
}

func TestImageCropped(t *testing.T) {
	b := NewBody()
	b.Styler(func(s *styles.Style) {
		s.Max.Set(units.Dp(75))
		s.Overflow.Set(styles.OverflowAuto)
	})
	img := NewImage(b)
	assert.NoError(t, img.Open(testImagePath))
	b.AssertRender(t, "image/cropped")
}

func TestImageScrolled(t *testing.T) {
	b := NewBody()
	b.Styler(func(s *styles.Style) {
		s.Max.Set(units.Dp(75))
		s.Overflow.Set(styles.OverflowAuto)
	})
	img := NewImage(b)
	assert.NoError(t, img.Open(testImagePath))
	b.AssertRender(t, "image/scrolled", func() {
		b.SystemEvents().Scroll(image.Pt(10, 10), math32.Vec2(2, 3))
	})
}
