package main

import (
	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/video"
)

func main() {
	b := core.NewBody("Basic Video Example")
	fr := core.NewFrame(b)

	fr.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
	})
	core.NewText(fr).SetText("video:").Styler(func(s *styles.Style) {
		s.SetTextWrap(false)
	})
	v := video.NewVideo(fr)
	v.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
	})
	errors.Log(v.Open("deer.mp4"))
	v.OnShow(func(e events.Event) {
		v.Play(0, 0)
	})
	b.RunMainWindow()
}
