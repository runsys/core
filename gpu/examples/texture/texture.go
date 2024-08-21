// Copyright (c) 2022, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"runtime"
	"time"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/gpu/examples/images"
	"cogentcore.org/core/math32"
)

//go:embed texture.wgsl
var texture string

func init() {
	// a must lock main thread for gpu!
	runtime.LockOSThread()
}

type CamView struct {
	Model      math32.Matrix4
	View       math32.Matrix4
	Projection math32.Matrix4
}

func main() {
	gp := gpu.NewGPU()
	gp.Config("texture")

	var resize func(size image.Point)
	size := image.Point{1024, 768}
	sp, terminate, pollEvents, size, err := gpu.GLFWCreateWindow(gp, size, "Draw Texture", &resize)
	if err != nil {
		return
	}

	sf := gpu.NewSurface(gp, sp, size, 4, gpu.UndefinedType)
	sy := gpu.NewGraphicsSystem(gp, "texture", sf)
	fmt.Printf("format: %s\n", sf.Format.String())

	resize = func(size image.Point) { sf.SetSize(size) }
	destroy := func() {
		sy.Release()
		sf.Release()
		gp.Release()
		terminate()
	}

	pl := sy.AddGraphicsPipeline("texture")
	sy.SetClearColor(color.RGBA{50, 50, 50, 255})
	// pl.SetFrontFace(wgpu.FrontFaceCW)
	// pl.SetCullMode(wgpu.CullModeBack)

	sh := pl.AddShader("texture")
	sh.OpenCode(texture)
	pl.AddEntry(sh, gpu.VertexShader, "vs_main")
	pl.AddEntry(sh, gpu.FragmentShader, "fs_main")

	vgp := sy.Vars().AddVertexGroup()
	tgp := sy.Vars().AddGroup(gpu.SampledTexture) // texture in 0 so frag only gets 0
	ugp := sy.Vars().AddGroup(gpu.Uniform)

	posv := vgp.Add("Pos", gpu.Float32Vector3, 0, gpu.VertexShader)
	clrv := vgp.Add("Color", gpu.Float32Vector3, 0, gpu.VertexShader)
	txcv := vgp.Add("TexCoord", gpu.Float32Vector2, 0, gpu.VertexShader)
	idxv := vgp.Add("Index", gpu.Uint16, 0, gpu.VertexShader)
	idxv.Role = gpu.Index

	camv := ugp.AddStruct("Camera", gpu.Float32Matrix4.Bytes()*3, 1, gpu.VertexShader)

	txv := tgp.Add("TexSampler", gpu.TextureRGBA32, 1, gpu.FragmentShader)

	vgp.SetNValues(1)
	ugp.SetNValues(1)
	tgp.SetNValues(3)

	sy.Config()

	imgFiles := []string{"ground.png", "wood.png", "teximg.jpg"}
	imgs := make([]image.Image, len(imgFiles))
	for i, fnm := range imgFiles {
		imgs[i], _, _ = imagex.OpenFS(images.Images, fnm)
		img := txv.Values.Values[i]
		img.SetFromGoImage(imgs[i], 0)
		// img.Texture.Sampler.Border = gpu.BorderBlack
		// img.Texture.Sampler.UMode = gpu.ClampToBorder
		// img.Texture.Sampler.VMode = gpu.ClampToBorder
	}

	gpu.SetValueFrom(posv.Values.Values[0], []float32{
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
		0.5, 0.5, 0.0,
		-0.5, 0.5, 0.0})
	gpu.SetValueFrom(clrv.Values.Values[0], []float32{
		1.0, 0.0, 0.0,
		0.0, 1.0, 0.0,
		0.0, 0.0, 1.0,
		1.0, 1.0, 0.0})
	gpu.SetValueFrom(txcv.Values.Values[0], []float32{
		1.0, 0.0,
		0.0, 0.0,
		0.0, 1.0,
		1.0, 1.0})
	gpu.SetValueFrom(idxv.Values.Values[0], []uint16{3, 2, 0, 2, 1, 0}) // reversed binding

	// This is the standard camera view projection computation
	campos := math32.Vec3(0, 0, 2)
	target := math32.Vec3(0, 0, 0)
	var lookq math32.Quat
	lookq.SetFromRotationMatrix(math32.NewLookAt(campos, target, math32.Vec3(0, 1, 0)))
	scale := math32.Vec3(1, 1, 1)
	var cview math32.Matrix4
	cview.SetTransform(campos, lookq, scale)
	view, _ := cview.Inverse()

	var camo CamView
	camo.Model.SetIdentity()
	camo.View.CopyFrom(view)
	aspect := float32(sf.Format.Size.X) / float32(sf.Format.Size.Y)
	camo.Projection.SetPerspective(45, aspect, 0.01, 100)
	cam := camv.Values.Values[0]
	gpu.SetValueFrom(cam, []CamView{camo}) // note: always use slice to copy

	frameCount := 0
	stTime := time.Now()

	renderFrame := func() {
		// fmt.Printf("frame: %d\n", frameCount)
		// rt := time.Now()
		camo.Model.SetRotationY(.004 * float32(frameCount))
		gpu.SetValueFrom(cam, []CamView{camo})

		imgIndex := (frameCount / 10) % len(imgs)
		pl.Vars().SetCurrentValue(0, "TexSampler", imgIndex)

		rp, err := sy.BeginRenderPass()
		if err != nil {
			return
		}
		pl.BindPipeline(rp)
		pl.BindDrawIndexed(rp)
		rp.End()
		sy.EndRenderPass(rp)

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

	fpsDelay := time.Second / 60
	fpsTicker := time.NewTicker(fpsDelay)
	for {
		select {
		case <-exitC:
			fpsTicker.Stop()
			destroy()
			return
		case <-fpsTicker.C:
			if !pollEvents() {
				exitC <- struct{}{}
				continue
			}
			renderFrame()
		}
	}
}
