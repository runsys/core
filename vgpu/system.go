// Copyright (c) 2022, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgpu

import (
	"fmt"
	"log"

	"github.com/goki/ki/ints"
	vk "github.com/vulkan-go/vulkan"
)

// System manages a system of Pipelines that all share
// a common collection of Vars, Vals, and a Memory manager.
// For example, this could be a collection of different
// pipelines for different material types, or different
// compute operations performed on a common set of data.
// It maintains its own logical device and associated queue.
type System struct {
	Name        string               `desc:"optional name of this System"`
	GPU         *GPU                 `desc:"gpu device"`
	Device      Device               `desc:"logical device for this System -- has its own queues"`
	Compute     bool                 `desc:"if true, this is a compute system -- otherwise is graphics"`
	Pipelines   []*Pipeline          `desc:"all pipelines"`
	PipelineMap map[string]*Pipeline `desc:"map of all pipelines -- names must be unique"`
	Vars        Vars                 `desc:"the common set of variables used by all Piplines"`
	Mem         Memory               `desc:"manages all the memory for all the Vals"`
	RenderPass  RenderPass           `desc:"renderpass with depth buffer for this system"`
	Framebuffer Framebuffer          `desc:"shared framebuffer to render into, if not rendering into Surface"`
}

// InitGraphics initializes the System for graphics use, using
// the graphics device from the Surface associated with this system
// or another device can be initialized by calling
// sy.Device.Init(gp, vk.QueueGraphicsBit)
func (sy *System) InitGraphics(gp *GPU, name string, dev *Device) error {
	sy.GPU = gp
	sy.Name = name
	sy.Compute = false
	sy.Device = *dev
	sy.Mem.Init(gp, &sy.Device)
	return nil
}

// InitCompute initializes the System for compute functionality,
// which creates its own Compute device.
func (sy *System) InitCompute(gp *GPU, name string) error {
	sy.GPU = gp
	sy.Name = name
	sy.Compute = true
	sy.Device.Init(gp, vk.QueueComputeBit)
	sy.Mem.Init(gp, &sy.Device)
	return nil
}

func (sy *System) Destroy() {
	for _, pl := range sy.Pipelines {
		pl.Destroy()
	}
	sy.Vars.Destroy(sy.Device.Device)
	if sy.Compute {
		sy.Device.Destroy()
	} else {
		sy.RenderPass.Destroy()
	}
	sy.Mem.Destroy(sy.Device.Device)
	sy.GPU = nil
}

// AddPipeline adds given pipeline
func (sy *System) AddPipeline(pl *Pipeline) {
	if sy.PipelineMap == nil {
		sy.PipelineMap = make(map[string]*Pipeline)
	}
	sy.Pipelines = append(sy.Pipelines, pl)
	sy.PipelineMap[pl.Name] = pl
}

// NewPipeline returns a new pipeline added to this System,
// initialized for use in this system.
func (sy *System) NewPipeline(name string) *Pipeline {
	pl := &Pipeline{Name: name}
	pl.Init(sy)
	sy.AddPipeline(pl)
	return pl
}

// ConfigRenderPass configures the renderpass, including the image
// format that we're rendering to (from the Surface or Framebuffer)
// and the depth buffer format (pass UndefType for no depth buffer).
func (sy *System) ConfigRenderPass(imgFmt *ImageFormat, depthFmt Types) {
	sy.RenderPass.Config(sy.Device.Device, imgFmt, depthFmt)
}

// Config configures the entire system, after everything has been
// setup (Pipelines, Vars, etc).  Memory / Vals do not yet need to
// be configured and are not Config'd by this call.
func (sy *System) Config() {
	sy.Vars.Config()
	if sy.GPU.Debug {
		fmt.Printf("%s\n", sy.Vars.StringDoc())
	}
	sy.Vars.DescLayout(sy.Device.Device)
	for _, pl := range sy.Pipelines {
		pl.Config()
	}
}

// SetVals sets the Vals for given Set of Vars, by name, in order
// that they appear in the Set list of roles and vars
func (sy *System) SetVals(set int, vals ...string) {
	nv := len(vals)
	var ws []vk.WriteDescriptorSet
	sd := sy.Vars.SetDesc[set]
	nv = ints.MinInt(nv, len(sd.Vars))
	for _, vnm := range vals {
		vl, err := sy.Mem.Vals.ValByNameTry(vnm)
		if err != nil {
			log.Println(err)
			continue
		}
		vl.Var.CurVal = vl
		if vl.Var.Role < Uniform {
			continue
		}
		wd := vk.WriteDescriptorSet{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          sd.DescSet,
			DstBinding:      uint32(vl.Var.BindLoc),
			DescriptorCount: 1,
			DescriptorType:  vl.Var.Role.VkDescriptor(),
		}
		if vl.Var.Role < TextureRole {
			off := vk.DeviceSize(vl.Offset)
			if vl.Var.Role.IsDynamic() {
				off = 0 // off must be 0 for dynamic
			}
			buff := sy.Mem.Buffs[vl.BuffType()]
			wd.PBufferInfo = []vk.DescriptorBufferInfo{{
				Offset: off,
				Range:  vk.DeviceSize(vl.MemSize),
				Buffer: buff.Dev,
			}}
			if vl.Var.Role.IsDynamic() {
				sy.Vars.DynOffs[vl.Var.DynOffIdx] = uint32(vl.Offset)
			}
		} else {
			wd.PImageInfo = []vk.DescriptorImageInfo{{
				ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
				ImageView:   vl.Texture.View,
				Sampler:     vl.Texture.VkSampler,
			}}
		}
		ws = append(ws, wd)
	}
	if len(ws) > 0 {
		vk.UpdateDescriptorSets(sy.Device.Device, uint32(len(ws)), ws, 0, nil)
	}
}

//////////////////////////////////////////////////////////////
// Set graphics options

// SetTopology sets the topology of vertex position data.
// TriangleList is the default.
// Also for Strip modes, restartEnable allows restarting a new
// strip by inserting a ??
// For all pipelines, to keep graphics settings consistent.
func (sy *System) SetTopology(topo Topologies, restartEnable bool) {
	for _, pl := range sy.Pipelines {
		pl.SetTopology(topo, restartEnable)
	}
}

// SetRasterization sets various options for how to rasterize shapes:
// Defaults are: vk.PolygonModeFill, vk.CullModeBackBit, vk.FrontFaceCounterClockwise, 1.0
// For all pipelines, to keep graphics settings consistent.
func (sy *System) SetRasterization(polygonMode vk.PolygonMode, cullMode vk.CullModeFlagBits, frontFace vk.FrontFace, lineWidth float32) {
	for _, pl := range sy.Pipelines {
		pl.SetRasterization(polygonMode, cullMode, frontFace, lineWidth)
	}
}

// SetCullFace sets the face culling mode: true = back, false = front
// use CullBack, CullFront constants
func (sy *System) SetCullFace(back bool) {
	for _, pl := range sy.Pipelines {
		pl.SetCullFace(back)
	}
}

// SetFrontFace sets the winding order for what counts as a front face
// true = CCW, false = CW
func (sy *System) SetFrontFace(ccw bool) {
	for _, pl := range sy.Pipelines {
		pl.SetFrontFace(ccw)
	}
}

// SetLineWidth sets the rendering line width -- 1 is default.
func (sy *System) SetLineWidth(lineWidth float32) {
	for _, pl := range sy.Pipelines {
		pl.SetLineWidth(lineWidth)
	}
}

// SetColorBlend determines the color blending function:
// either 1-source alpha (alphaBlend) or no blending:
// new color overwrites old.  Default is alphaBlend = true
// For all pipelines, to keep graphics settings consistent.
func (sy *System) SetColorBlend(alphaBlend bool) {
	for _, pl := range sy.Pipelines {
		pl.SetColorBlend(alphaBlend)
	}
}

// SetClearOff turns off clearing at start of rendering.
// call SetClearColor to turn back on.
func (sy *System) SetClearOff() {
	for _, pl := range sy.Pipelines {
		pl.SetClearOff()
	}
}

// SetClearColor sets the RGBA colors to set when starting new render
// For all pipelines, to keep graphics settings consistent.
func (sy *System) SetClearColor(r, g, b, a float32) {
	for _, pl := range sy.Pipelines {
		pl.SetClearColor(r, g, b, a)
	}
}

// SetClearDepthStencil sets the depth and stencil values when starting new render
// For all pipelines, to keep graphics settings consistent.
func (sy *System) SetClearDepthStencil(depth float32, stencil uint32) {
	for _, pl := range sy.Pipelines {
		pl.SetClearDepthStencil(depth, stencil)
	}
}
