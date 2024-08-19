// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

// Body is the common interface for all body types
type Body interface {
	Node

	// AsBodyBase returns the body as a BodyBase
	AsBodyBase() *BodyBase
}

// BodyBase is the base type for all specific Body types
type BodyBase struct {
	NodeBase

	// rigid body properties, including mass, bounce, friction etc
	Rigid Rigid

	// visualization name -- looks up an entry in the scene library that provides the visual representation of this body
	Vis string

	// default color of body for basic InitLibrary configuration
	Color string
}

func (bb *BodyBase) AsBody() Body {
	return bb.This.(Body)
}

func (bb *BodyBase) AsBodyBase() *BodyBase {
	return bb
}

func (bb *BodyBase) GroupBBox() {}
