// Copyright (c) 2022, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgpu

import (
	"image"

	vk "github.com/goki/vulkan"
)

// RenderFrame is an offscreen, non-window-backed rendering target,
// functioning like a Surface
type RenderFrame struct {

	// pointer to gpu device, for convenience
	GPU *GPU

	// device for this surface -- each window surface has its own device, configured for that surface
	Device Device

	// the Render for this RenderFrame, typically from a System
	Render *Render

	// has the current image format and dimensions
	Format ImageFormat

	// number of frames to maintain in the swapchain -- e.g., 2 = double-buffering, 3 = triple-buffering -- initially set to a requested amount, and after Init reflects actual number
	NFrames int

	// Framebuffers representing the Image owned by the RenderFrame -- we iterate through these in rendering subsequent frames
	Frames []*Framebuffer

	// semaphore used internally for waiting on acquisition of next frame
	ImageAcquired vk.Semaphore `display:"-"`

	// semaphore that surface user can wait on, will be activated when image has been acquired in AcquireNextFrame method
	RenderDone vk.Semaphore `display:"-"`

	// fence for rendering command running
	RenderFence vk.Fence `display:"-"`

	// do we own the device?
	OwnDevice bool
}

// NewRenderFrameOwnDevice returns a new renderframe initialized for given GPU,
// of given size.
// This version creates a new Graphics device -- for purely offscreen usage.
func NewRenderFrameOwnDevice(gp *GPU, size image.Point) *RenderFrame {
	rf := &RenderFrame{}
	rf.Defaults()
	rf.Format.Size = size
	rf.Init(gp, true) // make own device
	return rf
}

// NewRenderFrame returns a new renderframe initialized for given GPU,
// of given size.
// using given device, e.g., from a Surface -- to transition images
// from renderframe to surface, they must use the same device.
// if device is nil, own device is created.
func NewRenderFrame(gp *GPU, dev *Device, size image.Point) *RenderFrame {
	rf := &RenderFrame{}
	rf.Defaults()
	rf.Format.Size = size
	if dev != nil {
		rf.Device = *dev
		rf.Init(gp, false)
	} else {
		rf.Init(gp, true) // make dev
	}
	return rf
}

func (rf *RenderFrame) Defaults() {
	rf.NFrames = 1
	rf.Format.Defaults()
	// note: screen-correct results obtained by using Srgb here, which forces
	// this format in the final output.  Looks like what comes out from direct rendering.
	rf.Format.Set(1024, 768, vk.FormatR8g8b8a8Srgb)
	// rf.Format.Set(1024, 768, vk.FormatR8g8b8a8Unorm)
	rf.Format.SetMultisample(4)
}

// Init initializes the device and all other resources for the renderframe.
func (rf *RenderFrame) Init(gp *GPU, makeDevice bool) error {
	rf.GPU = gp
	if makeDevice {
		rf.OwnDevice = true
		rf.Device.Init(gp, vk.QueueGraphicsBit)
	} else {
		rf.OwnDevice = false
	}
	rf.Config()
	return nil
}

// Config configures the framebuffers etc
func (rf *RenderFrame) Config() {
	dev := rf.Device.Device
	rf.ImageAcquired = NewSemaphore(dev)
	rf.RenderDone = NewSemaphore(dev)
	rf.RenderFence = NewFence(dev)

	rf.Frames = make([]*Framebuffer, rf.NFrames)
	for i := 0; i < rf.NFrames; i++ {
		fr := &Framebuffer{}
		fr.ConfigRenderImage(rf.GPU, dev, rf.Format)
		rf.Frames[i] = fr
	}
}

// Free frees any existing (for ReInit or Destroy)
func (rf *RenderFrame) Free() {
	dev := rf.Device.Device
	vk.DeviceWaitIdle(dev)
	vk.DestroySemaphore(dev, rf.ImageAcquired, nil)
	vk.DestroySemaphore(dev, rf.RenderDone, nil)
	vk.DestroyFence(dev, rf.RenderFence, nil)
	for _, fr := range rf.Frames {
		fr.Destroy()
	}
	rf.Frames = nil
}

// SetRender sets the Render and updates frames accordingly
func (rf *RenderFrame) SetRender(rp *Render) {
	rf.Render = rp
	for _, fr := range rf.Frames {
		fr.ConfigRender(rp)
	}
}

// SetSize sets the size for the render frame,
// doesn't do anything if already that size (returns fale)
func (rf *RenderFrame) SetSize(size image.Point) bool {
	if rf.Format.Size == size {
		return false
	}
	rf.Format.Size = size
	rf.ReConfig()
	return true
}

// ReConfig reconfigures rendering
func (rf *RenderFrame) ReConfig() {
	rf.Render.SetSize(rf.Format.Size)
	rf.ReConfigFrames()
}

// ReConfigFrames re-configures the Famebuffers
// using exiting settings.
// Assumes Config has been called.
func (rf *RenderFrame) ReConfigFrames() {
	for _, fr := range rf.Frames {
		fr.ConfigRenderImage(rf.GPU, rf.Device.Device, rf.Format)
		fr.ConfigRender(rf.Render)
	}
}

func (rf *RenderFrame) Destroy() {
	rf.Free()
	if rf.OwnDevice {
		rf.Device.Destroy()
	}
	rf.GPU = nil
}

// SubmitRender submits a rendering command that must have been added
// to the given command buffer, calling CmdEnd on the buffer first.
// This buffer triggers the associated Fence logic to control the
// sequencing of render commands over time.
// The ImageAcquired semaphore before the command is run.
func (rf *RenderFrame) SubmitRender(cmd vk.CommandBuffer) {
	dev := rf.Device.Device
	vk.ResetFences(dev, 1, []vk.Fence{rf.RenderFence})
	CmdEnd(cmd)
	ret := vk.QueueSubmit(rf.Device.Queue, 1, []vk.SubmitInfo{{
		SType: vk.StructureTypeSubmitInfo,
		PWaitDstStageMask: []vk.PipelineStageFlags{
			vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		},
		// WaitSemaphoreCount:   1,
		// PWaitSemaphores:      []vk.Semaphore{rf.ImageAcquired},
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{cmd},
		// SignalSemaphoreCount: 1,
		// PSignalSemaphores:    []vk.Semaphore{rf.RenderDone},
	}}, rf.RenderFence)
	IfPanic(NewError(ret))
}

// WaitForRender waits until the last submitted render completes
func (rf *RenderFrame) WaitForRender() {
	dev := rf.Device.Device
	vk.WaitForFences(dev, 1, []vk.Fence{rf.RenderFence}, vk.True, vk.MaxUint64)
	vk.ResetFences(dev, 1, []vk.Fence{rf.RenderFence})
}

// GrabImage grabs rendered image of given index to Framebuffer.ImageGrab.
// must have waited for render already.
func (rf *RenderFrame) GrabImage(cmd vk.CommandBuffer, idx int) {
	dev := rf.Device.Device
	rf.Frames[idx].GrabImage(dev, cmd)
}

// GrabDepthImage grabs rendered depth image from the Render,
// must have waited for render already.
func (rf *RenderFrame) GrabDepthImage(cmd vk.CommandBuffer) {
	dev := rf.Device.Device
	rf.Render.GrabDepthImage(dev, cmd)
}
