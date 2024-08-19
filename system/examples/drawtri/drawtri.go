// Copyright 2023 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"
	"fmt"
	"image"
	"time"

	"cogentcore.org/core/cursors"
	"cogentcore.org/core/events"
	"cogentcore.org/core/system"
	_ "cogentcore.org/core/system/driver"
	"cogentcore.org/core/vgpu"
)

//go:embed *.spv
var content embed.FS

func main() {
	opts := &system.NewWindowOptions{
		Size:      image.Pt(1024, 768),
		StdPixels: true,
		Title:     "System Test Window",
	}
	w, err := system.TheApp.NewWindow(opts)
	if err != nil {
		panic(err)
	}

	fmt.Println("got new window", w)

	system.TheApp.Cursor(w).SetSize(32)

	var sf *vgpu.Surface
	var sy *vgpu.System
	var pl *vgpu.Pipeline

	make := func() {
		// note: drawer is always created and ready to go
		// we are creating an additional rendering system here.
		sf = w.Drawer().Surface().(*vgpu.Surface)
		sy = sf.GPU.NewGraphicsSystem("drawidx", &sf.Device)

		destroy := func() {
			sy.Destroy()
		}
		w.SetDestroyGPUResourcesFunc(destroy)

		pl = sy.NewPipeline("drawtri")
		sy.ConfigRender(&sf.Format, vgpu.UndefType)
		sf.SetRender(&sy.Render)

		pl.AddShaderEmbed("trianglelit", vgpu.VertexShader, content, "trianglelit.spv")
		pl.AddShaderEmbed("vtxcolor", vgpu.FragmentShader, content, "vtxcolor.spv")

		sy.Config()

		fmt.Println("made and configured pipelines")
	}

	frameCount := 0
	cur := cursors.Arrow
	stTime := time.Now()

	renderFrame := func() {
		if sf == nil {
			make()
		}
		// fmt.Printf("frame: %d\n", frameCount)
		// rt := time.Now()
		idx, ok := sf.AcquireNextImage()
		if !ok {
			return
		}
		// fmt.Printf("\nacq: %v\n", time.Now().Sub(rt))
		descIndex := 0 // if running multiple frames in parallel, need diff sets
		cmd := sy.CmdPool.Buff
		sy.ResetBeginRenderPass(cmd, sf.Frames[idx], descIndex)
		// fmt.Printf("rp: %v\n", time.Now().Sub(rt))
		pl.BindPipeline(cmd)
		pl.Draw(cmd, 3, 1, 0, 0)
		sy.EndRenderPass(cmd)
		sf.SubmitRender(cmd) // this is where it waits for the 16 msec
		// fmt.Printf("submit %v\n", time.Now().Sub(rt))
		sf.PresentImage(idx)
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
		if frameCount%60 == 0 {
			cur++
			if cur >= cursors.CursorN {
				cur = cursors.Arrow
			}
			err := system.TheApp.Cursor(w).Set(cur)
			if err != nil {
				fmt.Println("error setting cursor:", err)
			}
		}
	}

	go func() {
		for {
			evi := w.Events().Deque.NextEvent()
			et := evi.Type()
			if et != events.WindowPaint && et != events.MouseMove {
				fmt.Println("got event", evi)
			}
			switch et {
			case events.Window:
				ev := evi.(*events.WindowEvent)
				fmt.Println("got window event", ev)
				switch ev.Action {
				case events.WinShow:
					make()
				case events.WinClose:
					fmt.Println("got events.Close; quitting")
					system.TheApp.Quit()
				}
			case events.WindowPaint:
				// fmt.Println("paint")
				renderFrame()
			}
		}
	}()
	system.TheApp.MainLoop()
}
