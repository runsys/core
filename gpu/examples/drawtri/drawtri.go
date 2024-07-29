// Copyright (c) 2022, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"cogentcore.org/core/gpu"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rajveermalviya/go-webgpu/wgpu"
	wgpuext_glfw "github.com/rajveermalviya/go-webgpu/wgpuext/glfw"
)

func init() {
	// must lock main thread for gpu!
	runtime.LockOSThread()
}

func main() {
	if gpu.Init() != nil {
		return
	}

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err := glfw.CreateWindow(1024, 768, "Draw Triangle", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	gp := gpu.NewGPU()
	gp.Config("drawtri")

	sp := gp.Instance.CreateSurface(wgpuext_glfw.GetSurfaceDescriptor(window))
	width, height := window.GetSize()
	sf := gpu.NewSurface(gp, sp, width, height)

	fmt.Printf("format: %s\n", sf.Format.String())

	sy := gp.NewGraphicsSystem("drawtri", sf.Device)
	pl := sy.AddGraphicsPipeline("drawtri")
	sy.ConfigRender(&sf.Format, gpu.UndefType)
	pl.Primitive.CullMode = wgpu.CullMode_None
	// sy.SetCullMode(wgpu.CullMode_None)
	// sf.SetRender(&sy.Render)

	sh := pl.AddShader("trianglelit")
	sh.OpenFile("trianglelit.wgsl")
	pl.AddEntry(sh, gpu.VertexShader, "vs_main")
	pl.AddEntry(sh, gpu.FragmentShader, "fs_main")

	sy.Config()
	// no vars..

	destroy := func() {
		// vk.DeviceWaitIdle(sf.Device.Device)
		// todo: poll
		sy.Release()
		sf.Release()
		gp.Release()
		window.Destroy()
		gpu.Terminate()
	}

	frameCount := 0
	stTime := time.Now()

	renderFrame := func() {
		// fmt.Printf("frame: %d\n", frameCount)
		// rt := time.Now()
		view, err := sf.AcquireNextTexture()
		if err != nil {
			slog.Error(err.Error())
			return
		}
		// fmt.Printf("\nacq: %v\n", time.Now().Sub(rt))
		cmd := sy.NewCommandEncoder()
		rp := sy.BeginRenderPass(cmd, view)
		// fmt.Printf("rp: %v\n", time.Now().Sub(rt))
		pl.BindPipeline(rp)
		rp.Draw(3, 1, 0, 0)
		rp.End()
		sf.SubmitRender(cmd) // this is where it waits for the 16 msec
		// fmt.Printf("submit %v\n", time.Now().Sub(rt))
		sf.Present()
		// fmt.Printf("present %v\n\n", time.Now().Sub(rt))
		frameCount++
		eTime := time.Now()
		dur := float64(eTime.Sub(stTime)) / float64(time.Second)
		if dur > 10 {
			fps := float64(frameCount) / dur
			fmt.Printf("fps: %.0f\n", fps)
			frameCount = 0
			stTime = eTime
		}
	}

	exitC := make(chan struct{}, 2)

	fpsDelay := time.Second / 6
	fpsTicker := time.NewTicker(fpsDelay)
	for {
		select {
		case <-exitC:
			fpsTicker.Stop()
			destroy()
			return
		case <-fpsTicker.C:
			if window.ShouldClose() {
				exitC <- struct{}{}
				continue
			}
			glfw.PollEvents()
			renderFrame()
		}
	}
}
