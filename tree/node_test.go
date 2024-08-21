// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogentcore.org/core/base/slicesx"
	. "cogentcore.org/core/tree"
	"cogentcore.org/core/tree/testdata"
	"cogentcore.org/core/types"
)

func TestNodeAddChild(t *testing.T) {
	parent := NewNodeBase()
	child := &NodeBase{}
	parent.AddChild(child)
	child.SetName("child1")
	assert.Equal(t, 1, len(parent.Children))
	assert.Equal(t, parent, child.Parent)
	assert.Equal(t, "/node-base/child1", child.Path())
}

func TestNodeEmbedAddChild(t *testing.T) {
	parent := testdata.NewNodeEmbed()
	child := &testdata.NodeEmbed{}
	parent.AddChild(child)
	child.SetName("child1")
	assert.Len(t, parent.Children, 1)
	assert.Equal(t, parent, child.Parent)
	assert.Equal(t, "/node-embed/child1", child.Path())
}

func TestNodeEmbedNewChild(t *testing.T) {
	parent := testdata.NewNodeEmbed()
	child := parent.NewChild(parent.NodeType())
	child.AsTree().SetName("child1")
	assert.Len(t, parent.Children, 1)
	assert.Equal(t, "/node-embed/child1", child.AsTree().Path())
	assert.Equal(t, parent.NodeType(), child.AsTree().NodeType())
}

func TestNodePath(t *testing.T) {
	parent := testdata.NewNodeEmbed()
	child1 := parent.NewChild(parent.NodeType())
	child2 := parent.NewChild(parent.NodeType())
	child3 := parent.NewChild(parent.NodeType())
	assert.Len(t, parent.Children, 3)
	assert.Equal(t, "/node-embed/node-embed-0", child1.AsTree().Path())
	assert.Equal(t, "/node-embed/node-embed-1", child2.AsTree().Path())
	assert.Equal(t, "/node-embed/node-embed-2", child3.AsTree().Path())
}

func TestNodeEscapePaths(t *testing.T) {
	parent := NewNodeBase()
	child1 := NewNodeBase(parent).SetName("child1.go")
	child2 := NewNodeBase(parent).SetName("child1/child1")
	child3 := NewNodeBase(parent).SetName("child1/child1.go")
	schild2 := NewNodeBase(child2).SetName("subchild1")
	assert.Len(t, parent.Children, 3)
	assert.Equal(t, `/node-base/child1.go`, child1.Path())
	assert.Equal(t, `child1.go`, child1.PathFrom(parent))
	assert.Equal(t, `/node-base/child1\\child1`, child2.Path())
	assert.Equal(t, `/node-base/child1\\child1.go`, child3.Path())
	assert.Equal(t, `/node-base/child1\\child1/subchild1`, schild2.Path())
	assert.Equal(t, child1, parent.FindPath(child1.PathFrom(parent)))
	assert.Equal(t, child3, parent.FindPath(child3.PathFrom(parent)))
	assert.Equal(t, child3, parent.FindPath(child3.PathFrom(parent)))
	assert.Equal(t, schild2, parent.FindPath(schild2.PathFrom(parent)))
	assert.Equal(t, schild2, child2.FindPath(schild2.PathFrom(child2)))
}

func TestNodePathFrom(t *testing.T) {
	a := NewNodeBase().SetName("a")
	b := NewNodeBase(a).SetName("b")
	c := NewNodeBase(b).SetName("c")
	d := NewNodeBase(c).SetName("d")
	NewNodeBase(d).SetName("e")

	assert.Equal(t, "c/d", d.PathFrom(b))
}

func TestNodeDeleteChild(t *testing.T) {
	parent := NewNodeBase()
	child := NewNodeBase(parent)
	assert.Len(t, parent.Children, 1)
	assert.True(t, parent.DeleteChild(child))
	assert.Len(t, parent.Children, 0)
}

func TestNodeDeleteChildByName(t *testing.T) {
	parent := NewNodeBase()
	child := NewNodeBase(parent)
	child.SetName("child1")
	assert.Len(t, parent.Children, 1)
	assert.True(t, parent.DeleteChildByName("child1"))
	assert.Len(t, parent.Children, 0)
}

func TestNodeFindName(t *testing.T) {
	names := []string{"name0", "name1", "name2", "name3", "name4", "name5"}
	parent := NewNodeBase()
	for _, name := range names {
		child := NewNodeBase(parent)
		child.SetName(name)
	}
	assert.Len(t, parent.Children, len(names))
	for i, nm := range names {
		for st := range names { // test all starting indexes
			idx := IndexByName(parent.Children, nm, st)
			assert.Equal(t, i, idx)
		}
	}
}

func TestNodeFindType(t *testing.T) {
	parent := NewNodeBase()
	ne := testdata.NewNodeEmbed(parent)
	nb := NewNodeBase(parent)

	c0 := ChildByType[*testdata.NodeEmbed](parent)
	assert.Equal(t, ne, c0)
	c1 := ChildByType[*NodeBase](parent)
	assert.Equal(t, nb, c1)
}

func TestNodeMove(t *testing.T) {
	parent := testdata.NewNodeEmbed()
	testdata.NewNodeEmbed(parent).SetName("child0")
	child1 := NewNodeBase(parent).SetName("child1")
	testdata.NewNodeEmbed(parent).SetName("child2")
	testdata.NewNodeEmbed(child1).SetName("subchild1")
	NewNodeBase(parent).SetName("child3")

	bf := fmt.Sprintf("mv before:\n%v\n", parent.Children)
	parent.Children = slicesx.Move(parent.Children, 3, 1)
	a31 := fmt.Sprintf("mv 3 -> 1:\n%v\n", parent.Children)
	parent.Children = slicesx.Move(parent.Children, 0, 3)
	a03 := fmt.Sprintf("mv 0 -> 3:\n%v\n", parent.Children)
	parent.Children = slicesx.Move(parent.Children, 1, 2)
	a12 := fmt.Sprintf("mv 1 -> 2:\n%v\n", parent.Children)

	bft := `mv before:
[/node-embed/child0 /node-embed/child1 /node-embed/child2 /node-embed/child3]
`
	assert.Equal(t, bft, bf)
	a31t := `mv 3 -> 1:
[/node-embed/child0 /node-embed/child3 /node-embed/child1 /node-embed/child2]
`
	assert.Equal(t, a31t, a31)
	a03t := `mv 0 -> 3:
[/node-embed/child3 /node-embed/child1 /node-embed/child2 /node-embed/child0]
`
	assert.Equal(t, a03t, a03)
	a12t := `mv 1 -> 2:
[/node-embed/child3 /node-embed/child2 /node-embed/child1 /node-embed/child0]
`
	assert.Equal(t, a12t, a12)
}

func TestNodeUpdate(t *testing.T) {
	parent := testdata.NewNodeEmbed()
	testdata.NewNodeEmbed(parent).SetName("child0")
	child1 := NewNodeBase(parent).SetName("child1")
	testdata.NewNodeEmbed(parent).SetName("child2")
	testdata.NewNodeEmbed(child1).SetName("subchild1")
	NewNodeBase(parent).SetName("child3")

	plan1 := TypePlan{
		{Type: types.For[testdata.NodeEmbed](), Name: "child2"},
		{Type: types.For[testdata.NodeEmbed](), Name: "child3"},
		{Type: types.For[testdata.NodeEmbed](), Name: "child1"},
	}

	Update(parent, plan1)

	cf1 := fmt.Sprintf("plan1:\n%v\n", parent.Children)

	plan2 := TypePlan{
		{types.For[testdata.NodeEmbed](), "child4"},
		{types.For[NodeBase](), "child1"}, // note: changing this to Node type removes child1.subchild1
		{types.For[testdata.NodeEmbed](), "child5"},
		{types.For[testdata.NodeEmbed](), "child3"},
		{types.For[testdata.NodeEmbed](), "child6"},
	}

	Update(parent, plan2)

	cf2 := fmt.Sprintf("plan2:\n%v\n", parent.Children)

	cf1t := `plan1:
[/node-embed/child2 /node-embed/child3 /node-embed/child1]
`
	assert.Equal(t, cf1t, cf1)

	cf2t := `plan2:
[/node-embed/child4 /node-embed/child1 /node-embed/child5 /node-embed/child3 /node-embed/child6]
`
	assert.Equal(t, cf2t, cf2)
}

func TestNodeWalk(t *testing.T) {
	parent := NewNodeBase()
	testdata.NewNodeEmbed(parent).SetName("child0")
	child1 := NewNodeBase(parent).SetName("child1")
	testdata.NewNodeEmbed(parent).SetName("child2")
	schild1 := testdata.NewNodeEmbed(child1).SetName("subchild1")
	NewNodeBase(parent).SetName("child3")

	res := []string{}

	schild1.WalkUp(func(n Node) bool {
		res = append(res, n.AsTree().Name)
		return Continue
	})

	trg := []string{"subchild1", "child1", "node-base"}
	assert.Equal(t, trg, res)
	res = res[:0]

	parent.WalkDownPost(func(n Node) bool {
		return Continue
	},
		func(n Node) bool {
			res = append(res, fmt.Sprintf("[%s]", n.AsTree().Name))
			return Continue
		})
	trg = []string{"[child0]", "[subchild1]", "[child1]", "[child2]", "[child3]", "[node-base]"}
	assert.Equal(t, trg, res)
	res = res[:0]

	// test for Break working
	parent.WalkDownPost(func(n Node) bool {
		if n.AsTree().Name == "child1" {
			return Break
		}
		return Continue
	},
		func(n Node) bool {
			if n.AsTree().Name == "child1" {
				return Break
			}
			res = append(res, fmt.Sprintf("[%s]", n.AsTree().Name))
			return Continue
		})
	trg = []string{"[child0]", "[child2]", "[child3]", "[node-base]"}
	assert.Equal(t, trg, res)
	res = res[:0]

	parent.WalkDownBreadth(func(n Node) bool {
		res = append(res, fmt.Sprintf("[%v]", n.AsTree().Name))
		return Continue
	})
	trg = []string{"[node-base]", "[child0]", "[child1]", "[child2]", "[child3]", "[subchild1]"}
	assert.Equal(t, trg, res)
	res = res[:0]

	// test for return false
	parent.WalkDownBreadth(func(n Node) bool {
		if n.AsTree().Name == "child1" {
			return Break
		}
		res = append(res, fmt.Sprintf("[%v]", n.AsTree().Name))
		return Continue
	})
	trg = []string{"[node-base]", "[child0]", "[child2]", "[child3]"}
	assert.Equal(t, trg, res)
	res = res[:0]
}

func TestNodeWalkPath(t *testing.T) {
	parent := NewNodeBase()
	testdata.NewNodeEmbed(parent).SetName("child0")
	child1 := NewNodeBase(parent).SetName("child1")
	testdata.NewNodeEmbed(parent).SetName("child2")
	testdata.NewNodeEmbed(child1).SetName("subchild1")
	NewNodeBase(parent).SetName("child3")

	res := []string{}

	parent.WalkDown(func(n Node) bool {
		res = append(res, n.AsTree().Path())
		return Continue
	})
	assert.Equal(t, []string{"/node-base", "/node-base/child0", "/node-base/child1", "/node-base/child1/subchild1", "/node-base/child2", "/node-base/child3"}, res)
}

func TestProperties(t *testing.T) {
	n := testdata.NewNodeEmbed()

	n.SetProperty("intprop", 42)
	assert.Equal(t, 42, n.Property("intprop"))

	n.SetProperty("floatprop", 42.0)
	assert.Equal(t, 42.0, n.Property("floatprop"))

	n.SetProperty("stringprop", "test string")
	assert.Equal(t, "test string", n.Property("stringprop"))

	n.DeleteProperty("floatprop")
	assert.Equal(t, nil, n.Property("floatprop"))

	assert.Equal(t, nil, n.Property("randomprop"))

	assert.Equal(t, map[string]any{"intprop": 42, "stringprop": "test string"}, n.Properties)
}

func TestPropertiesJSON(t *testing.T) {
	testProperties := map[string]any{
		"floatprop":  3.1415,
		"stringprop": "type string",
		"#subproperties": map[string]any{
			"sp1": "#FFE",
			"sp2": 42.2,
		},
	}

	b, err := json.MarshalIndent(testProperties, "", "  ")
	require.NoError(t, err)

	res := map[string]any{}
	err = json.Unmarshal(b, &res)
	require.NoError(t, err)

	assert.Equal(t, testProperties, res)
}

// Test type directives: replacement for type properties
func TestDirectives(t *testing.T) {
	n := testdata.NewNodeEmbed()
	typ := n.NodeType()

	dir := typ.Directives[0]
	assert.Equal(t, types.Directive{Tool: "direct", Directive: "value"}, dir)
}

func TestSetUniqueName(t *testing.T) {
	root := NewNodeBase()
	assert.Equal(t, "node-base", root.Name)
	child := NewNodeBase(root)
	assert.Equal(t, "node-base-0", child.Name)
	child.SetName("my-name")
	assert.Equal(t, "my-name", child.Name)

	// does not change with SetParent when there is already a name
	SetParent(child, root)
	assert.Equal(t, "my-name", child.Name)

	// but does change with SetUniqueName when there is already a name
	SetUniqueName(child)
	assert.Equal(t, "my-name-2", child.Name)

	// ensure SetUniqueName works with different types
	newChild := testdata.NewNodeEmbed(root)
	assert.Equal(t, "node-embed-3", newChild.Name)
}

func TestTreeMod(t *testing.T) {
	// TODO: clean up these commented out tree tests
	/*
		tree1 := NodeBase{}
		typ := tree1.NodeType()
		tree1.InitName(&tree1, "tree1")
		// child11 :=
		tree1.NewChild(typ, "child11")
		child12 := tree1.NewChild(typ, "child12")
		// child13 :=
		tree1.NewChild(typ, "child13")
		// schild12 :=
		child12.NewChild(typ, "subchild12")

		tree2 := NodeBase{}
		tree2.InitName(&tree2, "tree2")
		// child21 :=
		tree2.NewChild(typ, "child21")
		child22 := tree2.NewChild(typ, "child22")
		// child23 :=
		tree2.NewChild(typ, "child23")
		// schild22 :=
		child22.NewChild(typ, "subchild22")

		// fmt.Printf("#################################\n")

		// fmt.Printf("Trees before:\n%v%v", tree1, tree2)
		MoveToParent(child12.This, tree2.This)

		// fmt.Printf("#################################\n")
		// fmt.Printf("Trees after add child12 move:\n%v%v", tree1, tree2)

		tree2.DeleteChild(child12)

		// fmt.Printf("#################################\n")
	*/
}

/*

func TestNodeFieldJSONSave(t *testing.T) {
	parent := testdata.NodeField2{}
	parent.InitName(&parent, "par1")
	typ := parent.NodeType()
	parent.Mbr1 = "bloop"
	parent.Mbr2 = 32
	// child1 :=
	parent.NewChild(typ, "child1")
	child2 := parent.NewChild(typ, "child1").(*testdata.NodeField2)
	// child3 :=
	parent.NewChild(typ, "child1")
	child2.NewChild(typ, "subchild1")

	var buf bytes.Buffer
	err := parent.WriteJSON(&buf, true)
	if err != nil {
		t.Error(err)
		// } else {
		// 	fmt.Printf("json output:\n%v\n", string(buf.Bytes()))
	}
	b := buf.Bytes()

	tstload := testdata.NodeField2{}
	tstload.InitName(&tstload, "")
	err = tstload.ReadJSON(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	} else {
		var buf2 bytes.Buffer
		err = tstload.WriteJSON(&buf2, true)
		if err != nil {
			t.Error(err)
		}
		tstb := buf2.Bytes()
		// fmt.Printf("test loaded json output: %v\n", string(tstb))
		if !bytes.Equal(tstb, b) {
			t.Error("original and unmarshal'd json rep are not equivalent")
			os.WriteFile("/tmp/jsonout1", b, 0644)
			os.WriteFile("/tmp/jsonout2", tstb, 0644)
		}
	}

	nwnd, err := ReadNewJSON(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	} else {
		var buf2 bytes.Buffer
		err = nwnd.WriteJSON(&buf2, true)
		if err != nil {
			t.Error(err)
		}
		tstb := buf2.Bytes()
		// fmt.Printf("test loaded json output: %v\n", string(tstb))
		if !bytes.Equal(tstb, b) {
			t.Error("original and unmarshal'd json rep are not equivalent")
		}
	}
}

*/

func TestClone(t *testing.T) {
	/*
		parent := testdata.NodeField2{}
		parent.InitName(&parent, "par1")
		typ := parent.NodeType()
		parent.Mbr1 = "bloop"
		parent.Mbr2 = 32
		// child1 :=
		parent.NewChild(typ, "child1")
		child2 := parent.NewChild(typ, "child1").(*testdata.NodeField2)
		// child3 :=
		parent.NewChild(typ, "child1")
		child2.NewChild(typ, "subchild1")

			var buf bytes.Buffer
			err := parent.WriteJSON(&buf, true)
			if err != nil {
				t.Error(err)
				// } else {
				// 	fmt.Printf("json output:\n%v\n", string(buf.Bytes()))
			}
			b := buf.Bytes()

			tstload := parent.Clone()
			var buf2 bytes.Buffer
			err = tstload.WriteJSON(&buf2, true)
			if err != nil {
				t.Error(err)
			}
			tstb := buf2.Bytes()
			// fmt.Printf("test loaded json output: %v\n", string(tstb))
			if !bytes.Equal(tstb, b) {
				t.Error("original and unmarshal'd json rep are not equivalent")
				os.WriteFile("/tmp/jsonout1", b, 0644)
				os.WriteFile("/tmp/jsonout2", tstb, 0644)
			}
	*/
}

func TestLiveType(t *testing.T) {
	type myNode struct {
		NodeBase
		MyField string
	}
	n := New[myNode]()
	assert.Equal(t, "cogentcore.org/core/tree_test.myNode", n.NodeType().Name)
	assert.Equal(t, "my-node", n.NodeType().IDName)
	assert.IsType(t, &myNode{}, n.NodeType().Instance)
	assert.IsType(t, &myNode{}, n.NewInstance())
}

func BenchmarkNodeType(b *testing.B) {
	n := NewNodeBase()
	for range b.N {
		n.NodeType()
	}
}

func BenchmarkNodeNewInstance(b *testing.B) {
	n := NewNodeBase()
	for range b.N {
		n.NewInstance()
	}
}
