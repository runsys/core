// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cursorimg

import (
	"image/color"
	"image/draw"
	"os"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/cursors"
)

func TestHotspot(t *testing.T) {
	err := os.MkdirAll("testdata", 0750)
	if err != nil {
		t.Fatal(err)
	}
	const size = 64
	for _, c := range cursors.CursorValues() {
		if c == cursors.None {
			continue
		}
		cur, err := Get(c, size)
		if err != nil {
			t.Fatal(err)
		}
		red := color.RGBA{R: 255, A: 255}
		hs := cur.Hotspot
		cur.Image.(draw.Image).Set(hs.X, hs.Y, red)
		err = imagex.Save(cur.Image, "testdata/"+c.String()+".png")
		if err != nil {
			t.Fatal(err)
		}
	}
}
