// Copyright 2022 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package shape provides a library of 3D shapes, built from indexed
triangle meshes, which can be added together in `ShapeGroup` lists.
Each `Shape` can report the number of points and indexes based on
configured parameters, and keeps track of its offset within an overall
[math32.ArrayF32] allocated based on total numbers.
In this way, separate Allocate then Configure phases are supported,
as required by the gpu Memory allocation system.

Basic building blocks (e.g., Plane, SphereSector) have standalone methods,
in addition to Shape elements.
*/
package shape
