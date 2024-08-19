// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"sync/atomic"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/slicesx"
	"cogentcore.org/core/types"
)

// admin.go has infrastructure code outside of the [Node] interface.

// New returns a new node of the given type with the given optional parent.
// If the name is unspecified, it defaults to the ID (kebab-case) name of
// the type, plus the [Node.NumLifetimeChildren] of the parent.
func New[T NodeValue](parent ...Node) *T {
	n := new(T)
	ni := any(n).(Node)
	initNode(ni)
	if len(parent) == 0 {
		ni.AsTree().SetName(ni.AsTree().NodeType().IDName)
		return n
	}
	p := parent[0]
	p.AsTree().Children = append(p.AsTree().Children, ni)
	SetParent(ni, p)
	return n
}

// NewOfType returns a new node of the given [types.Type] with the given optional parent.
// If the name is unspecified, it defaults to the ID (kebab-case) name of
// the type, plus the [Node.NumLifetimeChildren] of the parent.
func NewOfType(typ *types.Type, parent ...Node) Node {
	if len(parent) == 0 {
		n := newOfType(typ)
		initNode(n)
		n.AsTree().SetName(n.AsTree().NodeType().IDName)
		return n
	}
	return parent[0].AsTree().NewChild(typ)
}

// initNode initializes the node.
func initNode(n Node) {
	nb := n.AsTree()
	if nb.This != n {
		nb.This = n
		nb.This.Init()
	}
}

// checkThis checks that [Node.This] is non-nil.
// It returns and logs an error otherwise.
func checkThis(n Node) error {
	if n.AsTree().This != nil {
		return nil
	}
	return errors.Log(fmt.Errorf("tree.Node %q has nil tree.NodeBase.This; you must use tree.New or New* so that the node is initialized", n.AsTree().Path()))
}

// SetParent sets the parent of the given node to the given parent node.
// This is only for nodes with no existing parent; see [MoveToParent] to
// move nodes that already have a parent. It does not add the node to the
// parent's list of children; see [Node.AddChild] for a version that does.
// It automatically gets the [Node.This] of the parent.
func SetParent(child Node, parent Node) {
	n := child.AsTree()
	n.Parent = parent.AsTree().This
	setUniqueName(n, false)
	child.AsTree().This.OnAdd()
	n.WalkUpParent(func(pn Node) bool {
		pn.AsTree().This.OnChildAdded(child)
		return Continue
	})
}

// MoveToParent removes the given node from its current parent
// and adds it as a child of the given new parent.
// The old and new parents can be in different trees (or not).
func MoveToParent(child Node, parent Node) {
	oldParent := child.AsTree().Parent
	if oldParent != nil {
		idx := IndexOf(oldParent.AsTree().Children, child)
		if idx >= 0 {
			oldParent.AsTree().Children = slices.Delete(oldParent.AsTree().Children, idx, idx+1)
		}
	}
	parent.AsTree().AddChild(child)
}

// ParentByType returns the first parent of the given node that is
// of the given type, if any such node exists.
func ParentByType[T Node](n Node) T {
	if IsRoot(n) {
		var z T
		return z
	}
	if p, ok := n.AsTree().Parent.(T); ok {
		return p
	}
	return ParentByType[T](n.AsTree().Parent)
}

// ChildByType returns the first child of the given node that is
// of the given type, if any such node exists.
func ChildByType[T Node](n Node, startIndex ...int) T {
	nb := n.AsTree()
	idx := slicesx.Search(nb.Children, func(ch Node) bool {
		_, ok := ch.(T)
		return ok
	}, startIndex...)
	ch := nb.Child(idx)
	if ch == nil {
		var z T
		return z
	}
	return ch.(T)
}

// IsRoot returns whether the given node is the root node in its tree.
func IsRoot(n Node) bool {
	return n.AsTree().Parent == nil
}

// Root returns the root node of the given node's tree.
func Root(n Node) Node {
	if IsRoot(n) {
		return n
	}
	return Root(n.AsTree().Parent)
}

// nodeType is the [reflect.Type] of [Node].
var nodeType = reflect.TypeFor[Node]()

// IsNode returns whether the given type or a pointer to it
// implements the [Node] interface.
func IsNode(typ reflect.Type) bool {
	if typ == nil {
		return false
	}
	return typ.Implements(nodeType) || reflect.PointerTo(typ).Implements(nodeType)
}

// newOfType returns a new instance of the given [Node] type.
func newOfType(typ *types.Type) Node {
	return reflect.New(reflect.TypeOf(typ.Instance).Elem()).Interface().(Node)
}

// SetUniqueName sets the name of the node to be unique, using
// the number of lifetime children of the parent node as a unique
// identifier. If the node already has a name, it adds the unique id
// to it. Otherwise, it uses the type name of the node plus the unique id.
func SetUniqueName(n Node) {
	setUniqueName(n, true)
}

// setUniqueName is the implementation of [SetUniqueName] that takes whether
// to add the unique id to the name even if it is already set.
func setUniqueName(n Node, addIfSet bool) {
	nb := n.AsTree()
	pn := nb.Parent
	if pn == nil {
		return
	}
	c := atomic.AddUint64(&pn.AsTree().numLifetimeChildren, 1)
	id := "-" + strconv.FormatUint(c-1, 10) // must subtract 1 so we start at 0
	if nb.Name == "" {
		nb.SetName(nb.NodeType().IDName + id)
	} else if addIfSet {
		nb.SetName(nb.Name + id)
	}
}
