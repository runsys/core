// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xyz

import (
	"cogentcore.org/core/math32"
)

// BBox contains bounding box and other gross solid properties
type BBox struct {

	// bounding box in local coords
	BBox math32.Box3

	// bounding sphere in local coords
	BSphere math32.Sphere

	// area
	Area float32

	// volume
	Volume float32
}

// SetBounds sets BBox from min, max and updates other factors based on that
func (bb *BBox) SetBounds(min, max math32.Vector3) {
	bb.BBox.Set(&min, &max)
	bb.UpdateFromBBox()
}

// UpdateFromBBox updates other values from BBox
func (bb *BBox) UpdateFromBBox() {
	bb.BSphere.SetFromBox(bb.BBox)
	sz := bb.BBox.Size()
	bb.Area = 2*sz.X + 2*sz.Y + 2*sz.Z
	bb.Volume = sz.X * sz.Y * sz.Z
}
