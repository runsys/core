// Copyright (c) 2022, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgpu

import (
	"errors"
	"unsafe"

	vk "github.com/goki/vulkan"
)

// Device holds Device and associated Queue info
type Device struct {

	// logical device
	Device vk.Device

	// queue index for device
	QueueIndex uint32

	// queue for device
	Queue vk.Queue
}

// Init initializes a device based on QueueFlagBits
func (dv *Device) Init(gp *GPU, flags vk.QueueFlagBits) error {
	err := dv.FindQueue(gp, flags)
	if err != nil {
		return err
	}
	dv.MakeDevice(gp)
	return nil
}

// FindQueue finds queue for given flag bits, sets in QueueIndex
// returns error if not found.
func (dv *Device) FindQueue(gp *GPU, flags vk.QueueFlagBits) error {
	// Get queue family properties
	var queueCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(gp.GPU, &queueCount, nil)
	queueProperties := make([]vk.QueueFamilyProperties, queueCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(gp.GPU, &queueCount, queueProperties)
	if queueCount == 0 { // probably should try another GPU
		return errors.New("vulkan error: no queue families found on GPU 0")
	}

	// Find a suitable queue family for the target Vulkan mode
	found := false
	required := vk.QueueFlags(flags)
	for i := uint32(0); i < queueCount; i++ {
		queueProperties[i].Deref()
		if queueProperties[i].QueueFlags&required != 0 {
			dv.QueueIndex = i
			found = true
			break
		}
	}
	if !found {
		err := errors.New("GPU vulkan error: could not found queue with graphics capabilities")
		return err
	}
	return nil
}

// MakeDevice and Queue based on QueueIndex
func (dv *Device) MakeDevice(gp *GPU) {
	queueInfos := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: dv.QueueIndex,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}

	feats := vk.PhysicalDeviceFeatures{
		SamplerAnisotropy:                       vk.True, // used in Sampler.Config
		ShaderSampledImageArrayDynamicIndexing:  vk.True,
		ShaderUniformBufferArrayDynamicIndexing: vk.True,
		ShaderStorageBufferArrayDynamicIndexing: vk.True,
	}
	gp.SetGPUOpts(&feats, gp.EnabledOpts)

	// log.Printf("features: %#v\n", feats)

	var device vk.Device
	ret := vk.CreateDevice(gp.GPU, &vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(queueInfos)),
		PQueueCreateInfos:       queueInfos,
		EnabledExtensionCount:   uint32(len(gp.DeviceExts)),
		PpEnabledExtensionNames: gp.DeviceExts,
		EnabledLayerCount:       uint32(len(gp.ValidationLayers)),
		PpEnabledLayerNames:     gp.ValidationLayers,
		PEnabledFeatures:        []vk.PhysicalDeviceFeatures{feats},
		PNext:                   unsafe.Pointer(gp.DeviceFeaturesNeeded),
	}, nil, &device)
	IfPanic(NewError(ret))

	/* note: not using this for PNext:
	   unsafe.Pointer(&vk.PhysicalDeviceShaderAtomicFloatFeatures{
	   				SType:                        vk.StructureTypePhysicalDeviceShaderAtomicFloatFeatures,
	   				ShaderBufferFloat32AtomicAdd: vk.True,
	   			}),
	*/

	// _ = ret
	dv.Device = device

	var queue vk.Queue
	vk.GetDeviceQueue(dv.Device, dv.QueueIndex, 0, &queue)
	dv.Queue = queue
}

func (dv *Device) Destroy() {
	if dv.Device == nil {
		return
	}
	vk.DeviceWaitIdle(dv.Device)
	vk.DestroyDevice(dv.Device, nil)
	dv.Device = nil
}

// DeviceWaitIdle waits until the device is idle and ready
// for commands -- maybe useful to call if getting not ready
// errors in particular situations
func (dv *Device) DeviceWaitIdle() {
	vk.DeviceWaitIdle(dv.Device)
}

// NewGraphicsDevice returns a new Graphics Device, on given GPU.
// This is suitable for no display offscreen rendering.
// Typically use the Surface Device for rendering to a display window.
func NewGraphicsDevice(gp *GPU) (*Device, error) {
	dev := &Device{}
	if err := dev.Init(gp, vk.QueueGraphicsBit); err != nil {
		return nil, err
	}
	return dev, nil
}
