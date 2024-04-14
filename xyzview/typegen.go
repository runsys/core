// Code generated by "core generate"; DO NOT EDIT.

package xyzview

import (
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
	"cogentcore.org/core/xyz"
)

// ManipPointType is the [types.Type] for [ManipPoint]
var ManipPointType = types.AddType(&types.Type{Name: "cogentcore.org/core/xyzview.ManipPoint", IDName: "manip-point", Doc: "ManipPoint is a manipulation control point", Directives: []types.Directive{{Tool: "core", Directive: "no-new"}}, Embeds: []types.Field{{Name: "Solid"}}, Instance: &ManipPoint{}})

// NodeType returns the [*types.Type] of [ManipPoint]
func (t *ManipPoint) NodeType() *types.Type { return ManipPointType }

// New returns a new [*ManipPoint] value
func (t *ManipPoint) New() tree.Node { return &ManipPoint{} }

// SetMat sets the [ManipPoint.Mat]
func (t *ManipPoint) SetMat(v xyz.Material) *ManipPoint { t.Mat = v; return t }

// SceneType is the [types.Type] for [Scene]
var SceneType = types.AddType(&types.Type{Name: "cogentcore.org/core/xyzview.Scene", IDName: "scene", Doc: "Scene is a core.Widget that manages a xyz.Scene,\nproviding the basic rendering logic for the 3D scene\nin the 2D core GUI context.", Embeds: []types.Field{{Name: "WidgetBase"}}, Fields: []types.Field{{Name: "XYZ", Doc: "XYZ is the 3D xyz.Scene"}, {Name: "SelectionMode", Doc: "how to deal with selection / manipulation events"}, {Name: "CurrentSelected", Doc: "currently selected node"}, {Name: "CurrentManipPoint", Doc: "currently selected manipulation control point"}, {Name: "SelectionParams", Doc: "parameters for selection / manipulation box"}}, Instance: &Scene{}})

// NewScene adds a new [Scene] with the given name to the given parent:
// Scene is a core.Widget that manages a xyz.Scene,
// providing the basic rendering logic for the 3D scene
// in the 2D core GUI context.
func NewScene(parent tree.Node, name ...string) *Scene {
	return parent.NewChild(SceneType, name...).(*Scene)
}

// NodeType returns the [*types.Type] of [Scene]
func (t *Scene) NodeType() *types.Type { return SceneType }

// New returns a new [*Scene] value
func (t *Scene) New() tree.Node { return &Scene{} }

// SetSelectionMode sets the [Scene.SelectionMode]:
// how to deal with selection / manipulation events
func (t *Scene) SetSelectionMode(v SelectionModes) *Scene { t.SelectionMode = v; return t }

// SetCurrentSelected sets the [Scene.CurrentSelected]:
// currently selected node
func (t *Scene) SetCurrentSelected(v xyz.Node) *Scene { t.CurrentSelected = v; return t }

// SetCurrentManipPoint sets the [Scene.CurrentManipPoint]:
// currently selected manipulation control point
func (t *Scene) SetCurrentManipPoint(v *ManipPoint) *Scene { t.CurrentManipPoint = v; return t }

// SetSelectionParams sets the [Scene.SelectionParams]:
// parameters for selection / manipulation box
func (t *Scene) SetSelectionParams(v SelectionParams) *Scene { t.SelectionParams = v; return t }

// SetTooltip sets the [Scene.Tooltip]
func (t *Scene) SetTooltip(v string) *Scene { t.Tooltip = v; return t }

// SceneViewType is the [types.Type] for [SceneView]
var SceneViewType = types.AddType(&types.Type{Name: "cogentcore.org/core/xyzview.SceneView", IDName: "scene-view", Doc: "SceneView provides a toolbar controller for an xyz.Scene,\nand manipulation abilities.", Embeds: []types.Field{{Name: "Layout"}}, Instance: &SceneView{}})

// NewSceneView adds a new [SceneView] with the given name to the given parent:
// SceneView provides a toolbar controller for an xyz.Scene,
// and manipulation abilities.
func NewSceneView(parent tree.Node, name ...string) *SceneView {
	return parent.NewChild(SceneViewType, name...).(*SceneView)
}

// NodeType returns the [*types.Type] of [SceneView]
func (t *SceneView) NodeType() *types.Type { return SceneViewType }

// New returns a new [*SceneView] value
func (t *SceneView) New() tree.Node { return &SceneView{} }

// SetTooltip sets the [SceneView.Tooltip]
func (t *SceneView) SetTooltip(v string) *SceneView { t.Tooltip = v; return t }
