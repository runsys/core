// Copyright (c) 2022, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgpu

import (
	"fmt"
	"log"
	"sort"

	"log/slog"

	vk "github.com/goki/vulkan"
)

// MemSizeAlign returns the size aligned according to align byte increments
// e.g., if align = 16 and size = 12, it returns 16
func MemSizeAlign(size, align int) int {
	if size%align == 0 {
		return size
	}
	nb := size / align
	return (nb + 1) * align
}

// MemReg is a region of memory for transferring to / from GPU
type MemReg struct {
	Offset    int
	Size      int
	BuffType  BuffTypes
	BuffIndex int // for storage buffers, storage buffer index
}

// VarMem is memory allocation info per Var, for Storage types.
// Used in initial allocation algorithm.
type VarMem struct {

	// variable -- all Values of given Var are stored in the same Buffer
	Var *Var

	// index into storage buffer array holding this value
	Buff int

	// total size needed for this value, excluding alignment padding
	Size int

	// allocated offset within storage buffer for start of Var memory
	Offset int
}

// Memory manages memory for the GPU, using separate buffers for
// different roles, defined in the BuffTypes and managed by a MemBuff.
// Memory is organized by Vars with associated Values.
type Memory struct {
	GPU *GPU

	// logical device that this memory is managed for -- set from System
	Device Device

	// command pool for memory transfers
	CmdPool CmdPool

	// Vars variables used in shaders, which manage associated Values containing specific value instances of each var
	Vars Vars

	// memory buffers, organized by different Roles of vars.  Storage is managed separately in StorageBuffs
	Buffs [BuffTypesN]*MemBuff

	// memory buffers for storage -- vals are allocated to buffers during AllocHost, grouping based on what fits within GPU limits
	StorageBuffs []*MemBuff

	// memory allocation records for storage buffer allocation, per variable
	StorageMems []*VarMem
}

// Init configures the Memory for use with given gpu, device, and associated queueindex
func (mm *Memory) Init(gp *GPU, device *Device) {
	mm.GPU = gp
	mm.Device = *device
	mm.CmdPool.ConfigTransient(device)
	for bt := VtxIndexBuff; bt < BuffTypesN; bt++ {
		mm.Buffs[bt] = &MemBuff{Type: bt, GPU: mm.GPU}
	}
	mm.Vars.Mem = mm
}

// Destroy destroys all vulkan allocations, using given dev
func (mm *Memory) Destroy(dev vk.Device) {
	mm.Free()
	mm.Vars.Destroy(dev)
	mm.CmdPool.Destroy(dev)
	mm.GPU = nil
}

// Config should be called after all Values have been configured
// and are ready to go with their initial data.
// Does: AllocHost(), AllocDev().
// Note: dynamic binding must be called separately after this.
func (mm *Memory) Config(dev vk.Device) {
	mm.Free()
	mm.Vars.Config()
	mm.AllocHost()
	mm.AllocDev()
}

// AllocHost allocates memory for all buffers
func (mm *Memory) AllocHost() {
	for bt := VtxIndexBuff; bt < BuffTypesN; bt++ {
		mm.AllocHostBuff(bt)
	}
}

// AllocHostBuff allocates host memory for given buffer
func (mm *Memory) AllocHostBuff(bt BuffTypes) {
	switch bt {
	case StorageBuff:
		mm.AllocHostStorageBuff()
	default:
		buff := mm.Buffs[bt]
		buff.AlignBytes = buff.Type.AlignBytes(mm.GPU)
		bsz := mm.Vars.MemSize(buff)
		buff.AllocHost(mm.Device.Device, bsz)
		mm.Vars.AllocHost(buff, 0)
	}
}

// AllocHostStorageBuff allocates host memory for all storage buffers
func (mm *Memory) AllocHostStorageBuff() {
	// 1. collect StorageMems
	alignBytes := StorageBuff.AlignBytes(mm.GPU)
	mm.Vars.MemSizeStorage(mm, alignBytes)

	// 2. allocate to buffers
	sort.SliceStable(mm.StorageMems, func(i, j int) bool {
		return mm.StorageMems[i].Size < mm.StorageMems[j].Size
	})

	mm.StorageBuffs = nil
	maxBuff := int(mm.GPU.GPUProperties.Limits.MaxStorageBufferRange)
	curSz := 0
	var stb *MemBuff // current storage buffer
	for mi, vm := range mm.StorageMems {
		if vm.Size > maxBuff {
			slog.Error("vgpu.Memory AllocHostStorageBuff: variable needs more memory than Max available", "Variable", vm.Var.Name, "NeedsMemory", vm.Size, "MaxAvailable", maxBuff)
			vm.Size = maxBuff
		}
		newSz := curSz + vm.Size
		if newSz < maxBuff {
			if stb == nil {
				stb = &MemBuff{Type: StorageBuff, GPU: mm.GPU, AlignBytes: alignBytes}
				mm.StorageBuffs = append(mm.StorageBuffs, stb)
			}
			curSz = newSz
		} else {
			stb = &MemBuff{Type: StorageBuff, GPU: mm.GPU, AlignBytes: alignBytes}
			mm.StorageBuffs = append(mm.StorageBuffs, stb)
			curSz = vm.Size
		}
		vm.Var.StorageBuff = len(mm.StorageBuffs) - 1
		vm.Offset = stb.Size
		stb.Size = curSz
		if Debug {
			fmt.Printf("%02d:  Var: %s  Sz: 0x%X  Buf: %d  BufSz: 0x%X\n", mi, vm.Var.Name, vm.Size, vm.Var.StorageBuff, stb.Size)
		}
	}
	// 3. alloc host on buffers
	for _, stb := range mm.StorageBuffs {
		sz := stb.Size
		stb.Size = 0 // force alloc
		stb.AllocHost(mm.Device.Device, sz)
	}
	// 4. alloc host on vars
	for _, vm := range mm.StorageMems {
		stb := mm.StorageBuffs[vm.Var.StorageBuff]
		vm.Var.AllocHost(stb, vm.Offset)
	}
}

// AllocDev allocates device memory for all bufers
func (mm *Memory) AllocDev() {
	for bt := VtxIndexBuff; bt < BuffTypesN; bt++ {
		mm.AllocDevBuff(bt)
	}
}

// AllocDevBuff allocates memory on the device for given buffer
func (mm *Memory) AllocDevBuff(bt BuffTypes) {
	buff := mm.Buffs[bt]
	switch bt {
	case StorageBuff:
		for _, stb := range mm.StorageBuffs {
			stb.AllocDev(mm.Device.Device)
		}
	case TextureBuff:
		if buff.Size > 0 {
			mm.Vars.AllocTextures(mm)
		}
	default:
		if buff.Size > 0 {
			buff.AllocDev(mm.Device.Device)
		}
	}
}

// NewBuffer makes a buffer of given size, usage
func (mm *Memory) NewBuffer(size int, usage vk.BufferUsageFlagBits) vk.Buffer {
	return NewBuffer(mm.Device.Device, size, usage)
}

// AllocBuffMem allocates memory for given buffer, with given properties
func (mm *Memory) AllocBuffMem(buffer vk.Buffer, properties vk.MemoryPropertyFlagBits) vk.DeviceMemory {
	return AllocBuffMem(mm.GPU, mm.Device.Device, buffer, properties)
}

// FreeBuffMem frees given device memory to nil
func (mm *Memory) FreeBuffMem(memory *vk.DeviceMemory) {
	FreeBuffMem(mm.Device.Device, memory)
}

// Free frees memory for all buffers -- returns true if any freed
func (mm *Memory) Free() bool {
	freed := false
	for _, vm := range mm.StorageMems {
		vm.Var.Free()
	}
	mm.StorageMems = nil
	for bt := VtxIndexBuff; bt < BuffTypesN; bt++ {
		fr := mm.FreeBuff(bt)
		if fr {
			freed = true
		}
	}
	return freed
}

// FreeBuff frees any allocated memory in buffer -- returns true if freed
func (mm *Memory) FreeBuff(bt BuffTypes) bool {
	switch bt {
	case StorageBuff:
		freed := false
		for _, stb := range mm.StorageBuffs {
			if stb.Size > 0 {
				freed = true
				stb.Free(mm.Device.Device)
			}
		}
		mm.StorageBuffs = nil
		return freed
	default:
		buff := mm.Buffs[bt]
		mm.Vars.Free(buff)
		if buff.Size == 0 {
			return false
		}
		buff.Free(mm.Device.Device)
	}
	return true
}

// Deactivate deactivates device memory for all buffs
func (mm *Memory) Deactivate() {
	for bt := VtxIndexBuff; bt < BuffTypesN; bt++ {
		mm.DeactivateBuff(bt)
	}
}

// DeactivateBuff deactivates device memory in given buffer
func (mm *Memory) DeactivateBuff(bt BuffTypes) {
	switch bt {
	case StorageBuff:
		for _, stb := range mm.StorageBuffs {
			mm.FreeBuffMem(&stb.DevMem)
			stb.Active = false
		}
	default:
		buff := mm.Buffs[bt]
		mm.FreeBuffMem(&buff.DevMem)
		buff.Active = false
	}
}

// SyncToGPU syncs all modified Value regions from CPU to GPU device memory, for all buffs
func (mm *Memory) SyncToGPU() {
	for bt := VtxIndexBuff; bt < BuffTypesN; bt++ {
		mm.SyncToGPUBuff(bt)
	}
}

// SyncToGPUBuff syncs all modified Value regions from CPU to GPU device memory, for given buff
func (mm *Memory) SyncToGPUBuff(bt BuffTypes) {
	switch bt {
	case StorageBuff:
		for i, stb := range mm.StorageBuffs {
			mods := mm.Vars.ModRegsStorage(i, stb)
			if len(mods) > 0 {
				mm.TransferRegsToGPU(mods)
			}
		}
	case TextureBuff:
		mm.SyncValuesTextures(mm.Buffs[bt])
	default:
		mods := mm.Vars.ModRegs(bt)
		if len(mods) > 0 {
			mm.TransferRegsToGPU(mods)
		}
	}
}

// SyncRegionValueName returns memory region for syncing given value
// from GPU device memory to CPU host memory,
// specifying value by name for given named variable in given set.
// Variable can only only be Storage memory -- otherwise an error is returned.
// Multiple regions can be combined into one transfer call for greater efficiency.
func (mm *Memory) SyncRegionValueName(set int, varNm, valNm string) (MemReg, error) {
	vr, vl, err := mm.Vars.ValueByNameTry(set, varNm, valNm)
	if err != nil {
		return MemReg{}, err
	}
	if vr.BuffType() != StorageBuff {
		err = fmt.Errorf("SyncRegionValueName: Variable must be in Storage buffer, not: %s", vr.BuffType().String())
		if Debug {
			log.Println(err)
			return MemReg{}, err
		}
	}
	return vl.MemReg(vr), nil
}

// SyncRegionValueIndex returns memory region for syncing given value
// from GPU device memory to CPU host memory,
// specifying value by index for given named variable, in given set.
// Variable can only only be Storage memory -- otherwise an error is returned.
// Multiple regions can be combined into one transfer call for greater efficiency.
func (mm *Memory) SyncRegionValueIndex(set int, varNm string, valIndex int) (MemReg, error) {
	vr, vl, err := mm.Vars.ValueByIndexTry(set, varNm, valIndex)
	if err != nil {
		return MemReg{}, err
	}
	if vr.BuffType() != StorageBuff {
		err = fmt.Errorf("SyncRegionValueIndex: Variable must be in Storage buffer, not: %s", vr.BuffType().String())
		if Debug {
			log.Println(err)
			return MemReg{}, err
		}
	}
	return vl.MemReg(vr), nil
}

// SyncValueNameFromGPU syncs given value from GPU device memory to CPU host memory,
// specifying value by name for given named variable in given set.
// Variable can only only be Storage memory -- otherwise an error is returned.
func (mm *Memory) SyncValueNameFromGPU(set int, varNm, valNm string) error {
	mr, err := mm.SyncRegionValueName(set, varNm, valNm)
	if err != nil {
		return err
	}
	mm.TransferRegsFromGPU([]MemReg{mr})
	return nil
}

// SyncValueIndexFromGPU syncs given value from GPU device memory to CPU host memory,
// specifying value by index for given named variable, in given set.
// Variable can only only be Storage memory -- otherwise an error is returned.
func (mm *Memory) SyncValueIndexFromGPU(set int, varNm string, valIndex int) error {
	mr, err := mm.SyncRegionValueIndex(set, varNm, valIndex)
	if err != nil {
		return err
	}
	mm.TransferRegsFromGPU([]MemReg{mr})
	return nil
}

// SyncStorageRegionsFromGPU syncs given regions from the Storage buffer memory
// from GPU to CPU, in one call.   Use SyncRegValueIndexFromCPU to get the regions.
func (mm *Memory) SyncStorageRegionsFromGPU(regs ...MemReg) {
	mm.TransferRegsFromGPU(regs)
}

// TransferToGPU transfers entire staging to GPU for all buffs
func (mm *Memory) TransferToGPU() {
	for bt := VtxIndexBuff; bt < BuffTypesN; bt++ {
		mm.TransferToGPUBuff(bt)
	}
}

// TransferToGPUBuff transfers entire staging to GPU for given buffer
func (mm *Memory) TransferToGPUBuff(bt BuffTypes) {
	switch bt {
	case StorageBuff:
		regs := make([]MemReg, len(mm.StorageBuffs))
		for i, stb := range mm.StorageBuffs {
			regs[i] = MemReg{Offset: 0, Size: stb.Size, BuffType: bt, BuffIndex: i}
		}
		mm.TransferRegsToGPU(regs)
	case TextureBuff:
		mm.TransferAllValuesTextures(mm.Buffs[bt])
	default:
		buff := mm.Buffs[bt]
		mm.TransferRegsToGPU([]MemReg{{Offset: 0, Size: buff.Size, BuffType: bt}})
	}
}

// TransferRegsToGPU transfers memory from CPU to GPU for given regions,
// using a one-time memory command buffer.
// All buffs must be of the same type.
func (mm *Memory) TransferRegsToGPU(regs []MemReg) {
	if len(regs) == 0 {
		return
	}
	bt := regs[0].BuffType
	if bt == StorageBuff {
		mm.TransferStorageRegsToGPU(regs)
		return
	}
	buff := mm.Buffs[bt]
	if buff.Size == 0 || buff.DevMem == vk.NullDeviceMemory {
		return
	}
	cmd := mm.CmdPool.NewBuffer(&mm.Device)
	mm.CmdPool.BeginCmdOneTime()
	mm.CmdTransferRegsToGPU(cmd, buff, regs)
	mm.CmdPool.EndSubmitWaitFree(&mm.Device)
}

// TransferRegsFromGPU transfers memory from GPU to CPU for given regions,
// using a one-time memory command buffer.
// All buffs must be of the same type.
func (mm *Memory) TransferRegsFromGPU(regs []MemReg) {
	if len(regs) == 0 {
		return
	}
	bt := regs[0].BuffType
	if bt == StorageBuff {
		mm.TransferStorageRegsFromGPU(regs)
		return
	}
	buff := mm.Buffs[bt]
	if buff.Size == 0 || buff.DevMem == vk.NullDeviceMemory {
		return
	}
	cmd := mm.CmdPool.NewBuffer(&mm.Device)
	mm.CmdPool.BeginCmdOneTime()
	mm.CmdTransferRegsFromGPU(cmd, buff, regs)
	mm.CmdPool.EndSubmitWaitFree(&mm.Device)
}

// CmdTransferRegsToGPU transfers memory from CPU to GPU for given regions
// by recording command to given buffer.
func (mm *Memory) CmdTransferRegsToGPU(cmd vk.CommandBuffer, buff *MemBuff, regs []MemReg) {
	if buff.Size == 0 || buff.DevMem == vk.NullDeviceMemory || len(regs) == 0 {
		return
	}
	rg := make([]vk.BufferCopy, len(regs))
	for i, mr := range regs {
		rg[i] = vk.BufferCopy{SrcOffset: vk.DeviceSize(mr.Offset), DstOffset: vk.DeviceSize(mr.Offset), Size: vk.DeviceSize(mr.Size)}
	}
	vk.CmdCopyBuffer(cmd, buff.Host, buff.Dev, uint32(len(rg)), rg)
}

// CmdTransferRegsFromGPU transfers memory from GPU to CPU for given regions
// by recording command to given buffer.
func (mm *Memory) CmdTransferRegsFromGPU(cmd vk.CommandBuffer, buff *MemBuff, regs []MemReg) {
	if buff.Size == 0 || buff.DevMem == vk.NullDeviceMemory || len(regs) == 0 {
		return
	}
	rg := make([]vk.BufferCopy, len(regs))
	for i, mr := range regs {
		rg[i] = vk.BufferCopy{SrcOffset: vk.DeviceSize(mr.Offset), DstOffset: vk.DeviceSize(mr.Offset), Size: vk.DeviceSize(mr.Size)}
	}
	vk.CmdCopyBuffer(cmd, buff.Dev, buff.Host, uint32(len(rg)), rg)
}

// TransferStorageRegsToGPU transfers memory from CPU to GPU for given regions,
// using a one-time memory command buffer, for Storage
func (mm *Memory) TransferStorageRegsToGPU(regs []MemReg) {
	if len(regs) == 0 {
		return
	}
	cmd := mm.CmdPool.NewBuffer(&mm.Device)
	mm.CmdPool.BeginCmdOneTime()
	mm.CmdTransferStorageRegsToGPU(cmd, regs)
	mm.CmdPool.EndSubmitWaitFree(&mm.Device)
}

// TransferStorageRegsFromGPU transfers memory from GPU to CPU for given regions,
// using a one-time memory command buffer, for Storage
func (mm *Memory) TransferStorageRegsFromGPU(regs []MemReg) {
	if len(regs) == 0 {
		return
	}
	cmd := mm.CmdPool.NewBuffer(&mm.Device)
	CmdBeginOneTime(cmd)
	mm.CmdTransferStorageRegsFromGPU(cmd, regs)
	mm.CmdPool.EndSubmitWaitFree(&mm.Device)
}

// CmdTransferStorageRegsToGPU transfers memory from CPU to GPU for given storage regions
// by recording command to given command buffer.
func (mm *Memory) CmdTransferStorageRegsToGPU(cmd vk.CommandBuffer, regs []MemReg) {
	if len(regs) == 0 {
		return
	}
	sort.Slice(regs, func(i, j int) bool {
		return regs[i].BuffIndex < regs[j].BuffIndex
	})
	buffIndex := regs[0].BuffIndex
	buff := mm.StorageBuffs[buffIndex]
	mm.CmdTransferStorageBuffRegsToGPU(cmd, buff, buffIndex, regs)
	for i := 1; i < len(regs); i++ {
		if regs[i].BuffIndex == buffIndex {
			continue
		}
		buffIndex = regs[i].BuffIndex
		buff = mm.StorageBuffs[buffIndex]
		mm.CmdTransferStorageBuffRegsToGPU(cmd, buff, buffIndex, regs)
	}
}

// CmdTransferStorageBuffRegsToGPU transfers memory from CPU to GPU for given regions
// by recording command to given storage buffer of given index.
func (mm *Memory) CmdTransferStorageBuffRegsToGPU(cmd vk.CommandBuffer, buff *MemBuff, buffIndex int, regs []MemReg) {
	if buff.Size == 0 || buff.DevMem == vk.NullDeviceMemory || len(regs) == 0 {
		return
	}
	var rg []vk.BufferCopy
	for _, mr := range regs {
		if mr.BuffIndex != buffIndex {
			continue
		}
		rg = append(rg, vk.BufferCopy{SrcOffset: vk.DeviceSize(mr.Offset), DstOffset: vk.DeviceSize(mr.Offset), Size: vk.DeviceSize(mr.Size)})
	}
	vk.CmdCopyBuffer(cmd, buff.Host, buff.Dev, uint32(len(rg)), rg)
}

// CmdTransferStorageRegsFromGPU transfers memory from GPU to CPU for given storage regions
// by recording command to given command buffer.
func (mm *Memory) CmdTransferStorageRegsFromGPU(cmd vk.CommandBuffer, regs []MemReg) {
	if len(regs) == 0 {
		return
	}
	sort.Slice(regs, func(i, j int) bool {
		return regs[i].BuffIndex < regs[j].BuffIndex
	})
	buffIndex := regs[0].BuffIndex
	buff := mm.StorageBuffs[buffIndex]
	mm.CmdTransferStorageBuffRegsFromGPU(cmd, buff, buffIndex, regs)
	for i := 1; i < len(regs); i++ {
		if regs[i].BuffIndex == buffIndex {
			continue
		}
		buffIndex = regs[i].BuffIndex
		buff = mm.StorageBuffs[buffIndex]
		mm.CmdTransferStorageBuffRegsFromGPU(cmd, buff, buffIndex, regs)
	}
}

// CmdTransferStorageBuffRegsFromGPU transfers memory from GPU to CPU for given regions
// by recording command to given storage buffer of given index.
func (mm *Memory) CmdTransferStorageBuffRegsFromGPU(cmd vk.CommandBuffer, buff *MemBuff, buffIndex int, regs []MemReg) {
	if buff.Size == 0 || buff.DevMem == vk.NullDeviceMemory || len(regs) == 0 {
		return
	}
	var rg []vk.BufferCopy
	for _, mr := range regs {
		if mr.BuffIndex != buffIndex {
			continue
		}
		rg = append(rg, vk.BufferCopy{SrcOffset: vk.DeviceSize(mr.Offset), DstOffset: vk.DeviceSize(mr.Offset), Size: vk.DeviceSize(mr.Size)})
	}
	vk.CmdCopyBuffer(cmd, buff.Dev, buff.Host, uint32(len(rg)), rg)
}

////////////////////////////////////////////////////////////////////////////
// Texture functions

// TransferTexturesToGPU transfers texture image memory from CPU to GPU
// for given images.  Transitions device image as destination, then to Shader
// read only source, for use in fragment shader texture sampling.
// The image Host.Offset *must* be accurate for the given buffer, whether its own
// individual buffer or the shared memory-managed buffer.
func (mm *Memory) TransferTexturesToGPU(buff vk.Buffer, imgs ...*Image) {
	if len(imgs) == 0 {
		return
	}
	cmd := mm.CmdPool.NewBuffer(&mm.Device)
	CmdBeginOneTime(cmd)

	for _, im := range imgs {
		im.TransitionForDst(cmd, vk.PipelineStageTopOfPipeBit)
		if im.IsHostOwner() {
			cr := im.CopyRec()
			vk.CmdCopyBufferToImage(cmd, im.Host.Buff, im.Image, vk.ImageLayoutTransferDstOptimal, 1, []vk.BufferImageCopy{cr})
		} else {
			cr := im.CopyRec()
			vk.CmdCopyBufferToImage(cmd, buff, im.Image, vk.ImageLayoutTransferDstOptimal, 1, []vk.BufferImageCopy{cr})
		}
		im.TransitionDstToShader(cmd)
	}
	mm.CmdPool.EndSubmitWaitFree(&mm.Device)
}

// TransferImagesFromGPU transfers image memory from GPU to CPU for given images.
// the image Host.Offset *must* be accurate for the given buffer, whether its own
// individual buffer or the shared memory-managed buffer.
func (mm *Memory) TransferImagesFromGPU(buff vk.Buffer, imgs ...*Image) {
	cmd := mm.CmdPool.NewBuffer(&mm.Device)
	CmdBeginOneTime(cmd)

	for _, im := range imgs {
		vk.CmdCopyImageToBuffer(cmd, im.Image, vk.ImageLayoutTransferDstOptimal, buff, 1, []vk.BufferImageCopy{im.CopyRec()})
	}
	mm.CmdPool.EndSubmitWaitFree(&mm.Device)
}

// TransferAllValuesTextures copies all vals images from host buffer to device memory
func (mm *Memory) TransferAllValuesTextures(buff *MemBuff) {
	var imgs []*Image
	vs := &mm.Vars
	ns := vs.NSets()
	for si := vs.StartSet(); si < ns; si++ {
		st := vs.SetMap[si]
		if st == nil || st.Set == PushSet {
			continue
		}
		for _, vr := range st.Vars {
			if vr.Role != TextureRole {
				continue
			}
			vals := vr.Values.ActiveValues()
			for _, vl := range vals {
				if vl.Texture == nil {
					continue
				}
				imgs = append(imgs, &vl.Texture.Image)
			}
		}
	}
	if len(imgs) > 0 {
		mm.TransferTexturesToGPU(buff.Host, imgs...)
	}
}

// SyncValuesTextures syncs all changed vals images from host buffer to device memory
func (mm *Memory) SyncValuesTextures(buff *MemBuff) {
	var imgs []*Image
	vs := &mm.Vars
	ns := vs.NSets()
	for si := vs.StartSet(); si < ns; si++ {
		st := vs.SetMap[si]
		if st == nil || st.Set == PushSet {
			continue
		}
		for _, vr := range st.Vars {
			if vr.Role != TextureRole {
				continue
			}
			vals := vr.Values.ActiveValues()
			for _, vl := range vals {
				if vl.Texture == nil || !vl.IsMod() {
					continue
				}
				imgs = append(imgs, &vl.Texture.Image)
				vl.ClearMod()
			}
		}
	}
	if len(imgs) > 0 {
		mm.TransferTexturesToGPU(buff.Host, imgs...)
	}
}
