// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package xyz provides a 3D graphics framework in Go.
package xyz

//go:generate core generate -add-types

import (
	"fmt"
	"image"

	"cogentcore.org/core/base/ordmap"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/vgpu"
	"cogentcore.org/core/vgpu/vphong"
)

// Set Update3DTrace to true to get a trace of 3D updating
var Update3DTrace = false

// Scene is the overall scenegraph containing nodes as children.
// It renders to its own vgpu.RenderFrame.
// The Image of this Frame is usable directly or, via xyzcore.Scene,
// where it is copied into an overall core.Scene image.
//
// There is default navigation event processing (disabled by setting NoNav)
// where mouse drag events Orbit the camera (Shift = Pan, Alt = PanTarget)
// and arrow keys do Orbit, Pan, PanTarget with same key modifiers.
// Spacebar restores original "default" camera, and numbers save (1st time)
// or restore (subsequently) camera views (Control = always save)
//
// A Group at the top-level named "TrackCamera" will automatically track
// the camera (i.e., its Pose is copied) -- Solids in that group can
// set their relative Pos etc to display relative to the camera, to achieve
// "first person" effects.
type Scene struct {
	tree.NodeBase

	// Background is the background of the scene,
	// which is used directly as a solid color in Vulkan.
	Background image.Image

	// NeedsConfig means that a GPU resource (Lights, Texture, Meshes,
	// or more complex Nodes that require ConfigNodes) has been changed
	// and a Config call is required.
	NeedsConfig bool `set:"-"`

	// NeedsUpdate means that Node Pose has changed and an update pass
	// is required to update matrix and bounding boxes.
	NeedsUpdate bool `set:"-"`

	// NeedsRender means that something has been updated (minimally the
	// Camera pose) and a new Render is required.
	NeedsRender bool `set:"-"`

	// Viewport-level viewbox within any parent Viewport2D
	Geom math32.Geom2DInt `set:"-"`

	// number of samples in multisampling -- must be a power of 2, and must be 1 if grabbing the Depth buffer back from the RenderFrame
	MultiSample int `default:"4"`

	// render using wireframe instead of filled polygons -- this must be set prior to configuring the Phong rendering system (i.e., just after Scene is made)
	Wireframe bool `default:"false"`

	// camera determines view onto scene
	Camera Camera `set:"-"`

	// all lights used in the scene
	Lights ordmap.Map[string, Light] `set:"-"`

	// meshes -- holds all the mesh data -- must be configured prior to rendering
	Meshes ordmap.Map[string, Mesh] `set:"-"`

	// textures -- must be configured prior to rendering -- a maximum of 16 textures is supported for full cross-platform portability
	Textures ordmap.Map[string, Texture] `set:"-"`

	// library of objects that can be used in the scene
	Library map[string]*Group `set:"-"`

	// don't activate the standard navigation keyboard and mouse event processing to move around the camera in the scene
	NoNav bool

	// saved cameras -- can Save and Set these to view the scene from different angles
	SavedCams map[string]Camera `set:"-"`

	// the vphong rendering system
	Phong vphong.Phong `set:"-"`

	// the vgpu render frame holding the rendered scene
	Frame *vgpu.RenderFrame `set:"-"`

	// image used to hold a copy of the Frame image, for ImageCopy() call.
	// This is re-used across calls to avoid large memory allocations,
	// so it will automatically update after every ImageCopy call.
	// If a persistent image is required, call [iox/imagex.CloneAsRGBA].
	imgCopy image.RGBA `set:"-"`
}

func (sc *Scene) Init() {
	sc.MultiSample = 4
	sc.Camera.Defaults()
	sc.Background = colors.Scheme.Surface
}

// NewOffscreenScene returns a new [Scene] designed for offscreen
// rendering of 3D content. This can be used in unit tests and other
// cases not involving xyzcore. It makes a new [vgpu.NoDisplayGPU].
func NewOffscreenScene() *Scene {
	gpu, device, err := vgpu.NoDisplayGPU("offscreen")
	if err != nil {
		panic(fmt.Errorf("xyz.NewOffscreenScene: error initializing vgpu.NoDisplayGPU: %w", err))
	}
	sc := NewScene().SetSize(image.Pt(1280, 960))
	sc.ConfigFrame(gpu, device)
	return sc
}

// Update is a global update of everything: config, update, and re-render
func (sc *Scene) Update() {
	sc.Config()
	sc.SetNeedsUpdate()
}

// SaveCamera saves the current camera with given name -- can be restored later with SetCamera.
// "default" is a special name that is automatically saved on first render, and
// restored with the spacebar under default NavEvents.
// Numbered cameras 0-9 also saved / restored with corresponding keys.
func (sc *Scene) SaveCamera(name string) {
	if sc.SavedCams == nil {
		sc.SavedCams = make(map[string]Camera)
	}
	sc.SavedCams[name] = sc.Camera
	// fmt.Printf("saved camera %s: %v\n", name, sc.Camera.Pose.GenGoSet(".Pose"))
}

// SetCamera sets the current camera to that of given name -- error if not found.
// "default" is a special name that is automatically saved on first render, and
// restored with the spacebar under default NavEvents.
// Numbered cameras 0-9 also saved / restored with corresponding keys.
func (sc *Scene) SetCamera(name string) error {
	cam, ok := sc.SavedCams[name]
	if !ok {
		return fmt.Errorf("xyz.Scene: %v saved camera of name: %v not found", sc.Name, name)
	}
	sc.Camera = cam
	return nil
}

// DeleteUnusedMeshes deletes all unused meshes
func (sc *Scene) DeleteUnusedMeshes() {
	// used := make(map[string]struct{})
	// iterate over scene, add to used, then iterate over mats and if not used, delete.
}

// Validate traverses the scene and validates all the elements -- errors are logged
// and a non-nil return indicates that at least one error was found.
func (sc *Scene) Validate() error {
	// var errs []error // todo -- could do this
	// if err != nil {
	// 	*errs = append(*errs, err)
	// }
	hasError := false
	sc.WalkDown(func(k tree.Node) bool {
		if k == sc.This {
			return tree.Continue
		}
		ni, _ := AsNode(k)
		if !ni.IsVisible() {
			return tree.Break
		}
		err := ni.Validate()
		if err != nil {
			hasError = true
		}
		return tree.Continue
	})
	if hasError {
		return fmt.Errorf("xyz.Scene: %v Validate found at least one error (see log)", sc.Path())
	}
	return nil
}

// SetSize sets the size of the [Scene.Frame].
func (sc *Scene) SetSize(sz image.Point) *Scene {
	if sz.X == 0 || sz.Y == 0 {
		return sc
	}
	if sc.Frame != nil {
		csz := sc.Frame.Format.Size
		if csz == sz {
			sc.Geom.Size = sz // make sure
			return sc
		}
	}
	if sc.Frame != nil {
		sc.Frame.SetSize(sz)
	}
	sc.Geom.Size = sz // make sure
	// fmt.Printf("vp %v resized to: %v, bounds: %v\n", vp.Path(), nwsz, vp.Pixels.Bounds())
	return sc
}

func (sc *Scene) Destroy() {
	sc.Phong.Destroy()
	if sc.Frame != nil {
		sc.Frame.Destroy()
		sc.Frame = nil
		// fmt.Println("Phong, Frame destroyed")
	}
}

// SolidsIntersectingPoint finds all the solids that contain given 2D window coordinate
func (sc *Scene) SolidsIntersectingPoint(pos image.Point) []Node {
	var objs []Node
	for _, c := range sc.Children {
		cn, _ := AsNode(c)
		cn.AsTree().WalkDown(func(k tree.Node) bool {
			ni, _ := AsNode(k)
			if ni == nil {
				return tree.Break // going into a different type of thing, bail
			}
			if !ni.IsSolid() {
				return tree.Continue
			}
			// if nb.PosInWinBBox(pos) {
			objs = append(objs, ni)
			// }
			return tree.Continue
		})
	}
	return objs
}
