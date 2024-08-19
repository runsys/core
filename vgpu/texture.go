// Copyright (c) 2022, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgpu

import (
	vk "github.com/goki/vulkan"
)

// Texture supplies an Image and a Sampler
type Texture struct {
	Image

	// sampler for image
	Sampler
}

func (tx *Texture) Defaults() {
	tx.Image.Format.Defaults()
	tx.Sampler.Defaults()
}

func (tx *Texture) Destroy() {
	tx.Sampler.Destroy(tx.Image.Dev)
	tx.Image.Destroy()
}

// AllocTexture allocates texture device image, stdview, and sampler
func (tx *Texture) AllocTexture() {
	tx.AllocImage()
	if tx.Sampler.VkSampler == vk.NullSampler {
		tx.Sampler.Config(tx.GPU, tx.Dev)
	}
	tx.ConfigStdView()
}

///////////////////////////////////////////////////

// Sampler represents a vulkan image sampler
type Sampler struct {
	Name string

	// for U (horizontal) axis -- what to do when going off the edge
	UMode SamplerModes

	// for V (vertical) axis -- what to do when going off the edge
	VMode SamplerModes

	// for W (horizontal) axis -- what to do when going off the edge
	WMode SamplerModes

	// border color for Clamp modes
	Border BorderColors

	// the vulkan sampler
	VkSampler vk.Sampler
}

func (sm *Sampler) Defaults() {
	sm.UMode = Repeat
	sm.VMode = Repeat
	sm.WMode = Repeat
	sm.Border = BorderTrans
}

// Config configures sampler on device
func (sm *Sampler) Config(gp *GPU, dev vk.Device) {
	sm.Destroy(dev)
	var samp vk.Sampler
	ret := vk.CreateSampler(dev, &vk.SamplerCreateInfo{
		SType:                   vk.StructureTypeSamplerCreateInfo,
		MagFilter:               vk.FilterLinear,
		MinFilter:               vk.FilterLinear,
		AddressModeU:            sm.UMode.VkMode(),
		AddressModeV:            sm.VMode.VkMode(),
		AddressModeW:            sm.WMode.VkMode(),
		AnisotropyEnable:        vk.True,
		MaxAnisotropy:           gp.GPUProperties.Limits.MaxSamplerAnisotropy,
		BorderColor:             sm.Border.VkColor(),
		UnnormalizedCoordinates: vk.False,
		CompareEnable:           vk.False,
		MipmapMode:              vk.SamplerMipmapModeLinear,
	}, nil, &samp)
	IfPanic(NewError(ret))
	sm.VkSampler = samp
}

func (sm *Sampler) Destroy(dev vk.Device) {
	if sm.VkSampler != vk.NullSampler {
		vk.DestroySampler(dev, sm.VkSampler, nil)
		sm.VkSampler = vk.NullSampler
	}
}

// Texture image sampler modes
type SamplerModes int32 //enums:enum

const (
	// Repeat the texture when going beyond the image dimensions.
	Repeat SamplerModes = iota

	// Like repeat, but inverts the coordinates to mirror the image when going beyond the dimensions.
	MirroredRepeat

	// Take the color of the edge closest to the coordinate beyond the image dimensions.
	ClampToEdge

	// Return a solid color when sampling beyond the dimensions of the image.
	ClampToBorder

	// Like clamp to edge, but instead uses the edge opposite to the closest edge.
	MirrorClampToEdge
)

func (sm SamplerModes) VkMode() vk.SamplerAddressMode {
	return VulkanSamplerModes[sm]
}

var VulkanSamplerModes = map[SamplerModes]vk.SamplerAddressMode{
	Repeat:            vk.SamplerAddressModeRepeat,
	MirroredRepeat:    vk.SamplerAddressModeMirroredRepeat,
	ClampToEdge:       vk.SamplerAddressModeClampToEdge,
	ClampToBorder:     vk.SamplerAddressModeClampToBorder,
	MirrorClampToEdge: vk.SamplerAddressModeMirrorClampToEdge,
}

//////////////////////////////////////////////////////

// Texture image sampler modes
type BorderColors int32 //enums:enum -trim-prefix Border

const (
	// Repeat the texture when going beyond the image dimensions.
	BorderTrans BorderColors = iota
	BorderBlack
	BorderWhite
)

func (bc BorderColors) VkColor() vk.BorderColor {
	return VulkanBorderColors[bc]
}

var VulkanBorderColors = map[BorderColors]vk.BorderColor{
	BorderTrans: vk.BorderColorIntTransparentBlack,
	BorderBlack: vk.BorderColorIntOpaqueBlack,
	BorderWhite: vk.BorderColorIntOpaqueWhite,
}
