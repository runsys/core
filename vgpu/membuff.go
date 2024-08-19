// Copyright (c) 2022, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgpu

import (
	"log"
	"unsafe"

	vk "github.com/goki/vulkan"
)

// MemBuff is a memory buffer holding a particular type of memory
// with staging Host-based memory and Device memory
type MemBuff struct {
	GPU *GPU

	// type of memory in this buffer
	Type BuffTypes

	// allocated buffer size
	Size int

	// logical descriptor for host CPU-visible memory, for staging
	Host vk.Buffer `display:"-"`

	// host CPU-visible memory, for staging
	HostMem vk.DeviceMemory `display:"-"`

	// logical descriptor for device GPU-local memory, for computation
	Dev vk.Buffer `display:"-"`

	// device GPU-local memory, for computation
	DevMem vk.DeviceMemory `display:"-"`

	// memory mapped pointer into host memory -- remains mapped
	HostPtr unsafe.Pointer `display:"-"`

	// alignment of offsets into this buffer
	AlignBytes int

	// true if memory has been allocated, copied, transfered
	Active bool `edit:"-"`
}

// AllocHost allocates memory for this buffer of given size in bytes,
// freeing any existing memory allocated first.
// Host and Dev buffers are made, and host memory is allocated and mapped
// for staging purposes.  Call AllocDev to allocate device memory.
// Returns true if new memory was allocated.
func (mb *MemBuff) AllocHost(dev vk.Device, bsz int) bool {
	if bsz == mb.Size {
		return false
	}
	mb.Free(dev)
	if bsz == 0 {
		mb.Size = 0
		return false
	}
	usage := BuffUsages[mb.Type]
	hostUse := usage
	devUse := usage
	if mb.Type.IsReadOnly() {
		hostUse |= vk.BufferUsageTransferSrcBit
		devUse |= vk.BufferUsageTransferDstBit
	} else {
		hostUse |= vk.BufferUsageTransferSrcBit | vk.BufferUsageTransferDstBit
		devUse |= vk.BufferUsageTransferSrcBit | vk.BufferUsageTransferDstBit
	}
	mb.Host = NewBuffer(dev, bsz, hostUse)
	// mb.HostMem = AllocBuffMem(mb.GPU, dev, mb.Host, vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit)
	mb.HostMem = AllocBuffMem(mb.GPU, dev, mb.Host, vk.MemoryPropertyHostVisibleBit)
	mb.Size = bsz
	mb.HostPtr = MapMemory(dev, mb.HostMem, mb.Size)

	if mb.Type != TextureBuff {
		mb.Dev = NewBuffer(dev, bsz, devUse)
	}

	return true
}

// AllocDev allocates device local memory for this buffer.
func (mb *MemBuff) AllocDev(dev vk.Device) {
	mb.DevMem = AllocBuffMem(mb.GPU, dev, mb.Dev, vk.MemoryPropertyDeviceLocalBit)
}

// Free frees all memory for this buffer, including destroying
// buffers which have size associated with them.
func (mb *MemBuff) Free(dev vk.Device) {
	if mb.Size == 0 {
		return
	}
	if mb.Type != TextureBuff {
		FreeBuffMem(dev, &mb.DevMem)
		vk.DestroyBuffer(dev, mb.Dev, nil)
	}

	vk.UnmapMemory(dev, mb.HostMem)
	FreeBuffMem(dev, &mb.HostMem)
	vk.DestroyBuffer(dev, mb.Host, nil)
	mb.Size = 0
	mb.HostPtr = nil
	mb.Active = false
}

////////////////////////////////////////////////////////////////

// BuffTypes are memory buffer types managed by the Memory object
type BuffTypes int32 //enums:enum

const (
	// VtxIndexBuff is a buffer holding Vertex and Index values
	VtxIndexBuff BuffTypes = iota

	// UniformBuff holds Uniform and UniformTexel objects: read-only, small footprint
	UniformBuff

	// StorageBuff holds Storage and StorageTexel: read-write, larger
	// mostly for compute shaders
	StorageBuff

	// TextureBuff holds Images / Textures -- hardware optimizes allocation
	// on device side, and staging-side is general
	TextureBuff
)

// IsReadOnly returns true if buffer is read-only (most), else read-write (Storage)
func (bt BuffTypes) IsReadOnly() bool {
	if bt == StorageBuff {
		return false
	}
	return true
}

// AlignBytes returns alignment bytes for offsets into given buffer
func (bt BuffTypes) AlignBytes(gp *GPU) int {
	switch bt {
	case StorageBuff:
		return int(gp.GPUProperties.Limits.MinStorageBufferOffsetAlignment)
	case UniformBuff, VtxIndexBuff:
		return int(gp.GPUProperties.Limits.MinUniformBufferOffsetAlignment)
	case TextureBuff:
		return int(gp.GPUProperties.Limits.MinTexelBufferOffsetAlignment)
	}
	return int(gp.GPUProperties.Limits.MinUniformBufferOffsetAlignment)
}

// BuffUsages maps BuffTypes into buffer usage flags
var BuffUsages = map[BuffTypes]vk.BufferUsageFlagBits{
	VtxIndexBuff: vk.BufferUsageVertexBufferBit | vk.BufferUsageIndexBufferBit,
	UniformBuff:  vk.BufferUsageUniformBufferBit | vk.BufferUsageUniformTexelBufferBit,
	StorageBuff:  vk.BufferUsageStorageBufferBit | vk.BufferUsageStorageTexelBufferBit,
	TextureBuff:  vk.BufferUsageStorageTexelBufferBit,
}

/////////////////////////////////////////////////////////////////////
// Basic memory functions

// NewBuffer makes a buffer of given size, usage
func NewBuffer(dev vk.Device, size int, usage vk.BufferUsageFlagBits) vk.Buffer {
	if size == 0 {
		return vk.NullBuffer
	}
	var buffer vk.Buffer
	ret := vk.CreateBuffer(dev, &vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Usage:       vk.BufferUsageFlags(usage),
		Size:        vk.DeviceSize(size),
		SharingMode: vk.SharingModeExclusive,
	}, nil, &buffer)
	IfPanic(NewError(ret))
	return buffer
}

// AllocBuffMem allocates memory for given buffer, with given properties
func AllocBuffMem(gp *GPU, dev vk.Device, buffer vk.Buffer, properties vk.MemoryPropertyFlagBits) vk.DeviceMemory {
	// Ask device about its memory requirements.
	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(dev, buffer, &memReqs)
	memReqs.Deref()

	memProperties := gp.MemoryProperties
	memType, ok := FindRequiredMemoryType(memProperties, vk.MemoryPropertyFlagBits(memReqs.MemoryTypeBits), properties)
	if !ok {
		log.Println("vulkan warning: failed to find required memory type")
	}

	var memory vk.DeviceMemory
	// Allocate device memory and bind to the buffer.
	ret := vk.AllocateMemory(dev, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memType,
	}, nil, &memory)
	IfPanic(NewError(ret))
	vk.BindBufferMemory(dev, buffer, memory, 0)
	return memory
}

// MapMemory maps the buffer memory, returning a pointer into start of buffer memory
func MapMemory(dev vk.Device, mem vk.DeviceMemory, size int) unsafe.Pointer {
	var buffPtr unsafe.Pointer
	ret := vk.MapMemory(dev, mem, 0, vk.DeviceSize(size), 0, &buffPtr)
	if IsError(ret) {
		log.Printf("vulkan MapMemory warning: failed to map device memory for data (len=%d)", size)
		return nil
	}
	return buffPtr
}

// MapMemoryAll maps the WholeSize of buffer memory,
// returning a pointer into start of buffer memory
func MapMemoryAll(dev vk.Device, mem vk.DeviceMemory) unsafe.Pointer {
	var buffPtr unsafe.Pointer
	ret := vk.MapMemory(dev, mem, 0, vk.DeviceSize(vk.WholeSize), 0, &buffPtr)
	if IsError(ret) {
		log.Printf("vulkan MapMemory warning: failed to map device memory for data")
		return nil
	}
	return buffPtr
}

// FreeBuffMem frees given device memory to nil
func FreeBuffMem(dev vk.Device, memory *vk.DeviceMemory) {
	if *memory == vk.NullDeviceMemory {
		return
	}
	vk.FreeMemory(dev, *memory, nil)
	*memory = vk.NullDeviceMemory
}

// DestroyBuffer destroys given buffer and nils the pointer
func DestroyBuffer(dev vk.Device, buff *vk.Buffer) {
	if *buff == vk.NullBuffer {
		return
	}
	vk.DestroyBuffer(dev, *buff, nil)
	*buff = vk.NullBuffer
}

func FindRequiredMemoryType(properties vk.PhysicalDeviceMemoryProperties,
	deviceRequirements, hostRequirements vk.MemoryPropertyFlagBits) (uint32, bool) {

	for i := uint32(0); i < vk.MaxMemoryTypes; i++ {
		if deviceRequirements&(vk.MemoryPropertyFlagBits(1)<<i) != 0 {
			properties.MemoryTypes[i].Deref()
			flags := properties.MemoryTypes[i].PropertyFlags
			if flags&vk.MemoryPropertyFlags(hostRequirements) != 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func FindRequiredMemoryTypeFallback(properties vk.PhysicalDeviceMemoryProperties,
	deviceRequirements, hostRequirements vk.MemoryPropertyFlagBits) (uint32, bool) {

	for i := uint32(0); i < vk.MaxMemoryTypes; i++ {
		if deviceRequirements&(vk.MemoryPropertyFlagBits(1)<<i) != 0 {
			properties.MemoryTypes[i].Deref()
			flags := properties.MemoryTypes[i].PropertyFlags
			if flags&vk.MemoryPropertyFlags(hostRequirements) != 0 {
				return i, true
			}
		}
	}
	// Fallback to the first one available.
	if hostRequirements != 0 {
		return FindRequiredMemoryType(properties, deviceRequirements, 0)
	}
	return 0, false
}
