// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command docs provides documentation of Cogent Core,
// hosted at https://cogentcore.org/core.
package main

//go:generate core generate

import (
	"embed"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/htmlcore"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/pages"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/texteditor"
	"cogentcore.org/core/tree"
)

//go:embed content
var content embed.FS

//go:embed *.svg name.png weld-icon.png
var resources embed.FS

//go:embed image.png
var myImage embed.FS

//go:embed icon.svg
var mySVG embed.FS

//go:embed file.go
var myFile embed.FS

func main() {
	b := core.NewBody("Cogent Core Docs")
	pg := pages.NewPage(b).SetSource(fsx.Sub(content, "content"))
	pg.Context.WikilinkResolver = htmlcore.PkgGoDevWikilink("cogentcore.org/core")
	b.AddAppBar(pg.MakeToolbar)

	htmlcore.ElementHandlers["home-page"] = homePage

	b.RunMainWindow()
}

var home *core.Frame

func makeBlock[T tree.NodeValue](title, text string, graphic func(w *T)) {
	tree.AddChildAt(home, title, func(w *core.Frame) {
		w.Styler(func(s *styles.Style) {
			s.Gap.Set(units.Em(1))
			s.Grow.Set(1, 0)
			if home.SizeClass() == core.SizeCompact {
				s.Direction = styles.Column
			}
		})
		w.Maker(func(p *tree.Plan) {
			graphicFirst := w.IndexInParent()%2 != 0 && w.SizeClass() != core.SizeCompact
			if graphicFirst {
				tree.Add(p, graphic)
			}
			tree.Add(p, func(w *core.Frame) {
				w.Styler(func(s *styles.Style) {
					s.Direction = styles.Column
					s.Text.Align = styles.Start
					s.Grow.Set(1, 1)
				})
				tree.AddChild(w, func(w *core.Text) {
					w.SetType(core.TextHeadlineLarge).SetText(title)
					w.Styler(func(s *styles.Style) {
						s.Font.Weight = styles.WeightBold
						s.Color = colors.Scheme.Primary.Base
					})
				})
				tree.AddChild(w, func(w *core.Text) {
					w.SetType(core.TextTitleLarge).SetText(text)
				})
			})
			if !graphicFirst {
				tree.Add(p, graphic)
			}
		})
	})
}

func homePage(ctx *htmlcore.Context) bool {
	home = core.NewFrame(ctx.BlockParent)
	home.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		s.CenterAll()
	})

	tree.AddChild(home, func(w *core.SVG) {
		errors.Log(w.ReadString(core.AppIcon))
	})
	tree.AddChild(home, func(w *core.Image) {
		errors.Log(w.OpenFS(resources, "name.png"))
		w.Styler(func(s *styles.Style) {
			x := func(uc *units.Context) float32 {
				return min(uc.Dp(612), uc.Pw(90))
			}
			s.Min.Set(units.Custom(x), units.Custom(func(uc *units.Context) float32 {
				return x(uc) * (128.0 / 612.0)
			}))
		})
	})
	tree.AddChild(home, func(w *core.Text) {
		w.SetType(core.TextHeadlineMedium).SetText("A cross-platform framework for building powerful, fast, and elegant 2D and 3D apps")
	})
	tree.AddChild(home, func(w *core.Button) {
		w.SetText("Get started")
		w.OnClick(func(e events.Event) {
			ctx.OpenURL("getting-started")
		})
	})

	initIcon := func(w *core.Icon) *core.Icon {
		w.Styler(func(s *styles.Style) {
			s.Min.Set(units.Dp(256))
			s.Color = colors.Scheme.Primary.Base
		})
		return w
	}

	makeBlock("CODE ONCE, RUN EVERYWHERE", "With Cogent Core, you can write your app once and it will instantly run on macOS, Windows, Linux, iOS, Android, and the Web, automatically scaling to any screen. Instead of struggling with platform-specific code in a multitude of languages, you can easily write and maintain a single pure Go codebase.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.Devices)
	})

	makeBlock("EFFORTLESS ELEGANCE", "Cogent Core is built on Go, a high-level language designed for building elegant, readable, and scalable code with full type safety and a robust design that never gets in your way. Cogent Core makes it easy to get started with cross-platform app development in just two commands and seven lines of simple code.", func(w *texteditor.Editor) {
		w.SetNewBuffer()
		w.Buffer.SetLang("go").SetTextString(`func main() {
			b := core.NewBody("Hello")
			core.NewButton(b).SetText("Hello, World!")
			b.RunMainWindow()
		}`)
		w.Styler(func(s *styles.Style) {
			s.Min.X.Pw(50)
		})
	})

	makeBlock("COMPLETELY CUSTOMIZABLE", "Cogent Core allows developers and users to fully customize apps to fit their unique needs and preferences through a robust styling system and a powerful color system that allow developers and users to instantly customize every aspect of the appearance and behavior of an app.", func(w *core.Form) {
		w.SetStruct(core.AppearanceSettings)
		w.OnChange(func(e events.Event) {
			core.UpdateSettings(w, core.AppearanceSettings)
		})

	})

	makeBlock("POWERFUL FEATURES", "Cogent Core comes with a powerful set of advanced features that allow you to make almost anything, including fully featured text editors, video and audio players, interactive 3D graphics, customizable data plots, Markdown and HTML rendering, SVG and canvas vector graphics, and automatic views of any Go data structure for instant data binding and advanced app inspection.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.ScatterPlot)
	})

	makeBlock("OPTIMIZED EXPERIENCE", "Every part of your development experience is guided by a comprehensive set of interactive example-based documentation, in-depth video tutorials, easy-to-use command line tools specialized for Cogent Core, and active support and development from the Cogent Core developers.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.PlayCircle)
	})

	makeBlock("EXTREMELY FAST", "Cogent Core is powered by Vulkan, a modern, cross-platform, high-performance graphics framework that allows apps to run on all platforms at extremely fast speeds. All Cogent Core apps compile to machine code, allowing them to run without any overhead.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.Bolt)
	})

	makeBlock("FREE AND OPEN SOURCE", "Cogent Core is completely free and open source under the permissive BSD-3 License, allowing you to use Cogent Core for any purpose, commercially or personally. We believe that software works best when everyone can use it.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.Code)
	})

	makeBlock("USED AROUND THE WORLD", "Over six years of development, Cogent Core has been used and thoroughly tested by developers and scientists around the world for a wide variety of use cases. Cogent Core is an advanced framework actively used to power everything from end-user apps to scientific research.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.GlobeAsia)
	})

	core.NewText(home).SetType(core.TextDisplaySmall).SetText("<b>What can Cogent Core do?</b>")

	makeBlock("COGENT CODE", "Cogent Code is a fully featured Go IDE with support for syntax highlighting, code completion, symbol lookup, building and debugging, version control, keyboard shortcuts, and many other features.", func(w *core.SVG) {
		errors.Log(w.OpenFS(resources, "code-icon.svg"))
	})

	makeBlock("COGENT VECTOR", "Cogent Vector is a powerful vector graphics editor with complete support for shapes, paths, curves, text, images, gradients, groups, alignment, styling, importing, exporting, undo, redo, and various other features.", func(w *core.SVG) {
		errors.Log(w.OpenFS(resources, "vector-icon.svg"))
	})

	makeBlock("COGENT NUMBERS", "Cogent Numbers is a highly extensible math, data science, and statistics platform that combines the power of programming with the convenience of spreadsheets and graphing calculators.", func(w *core.SVG) {
		errors.Log(w.OpenFS(resources, "numbers-icon.svg"))
	})

	makeBlock("COGENT MAIL", "Cogent Mail is a customizable email client with built-in Markdown support, automatic mail filtering, and an extensive set of keyboard shortcuts for advanced mail filing.", func(w *core.SVG) {
		errors.Log(w.OpenFS(resources, "mail-icon.svg"))
	})

	makeBlock(`<a href="https://emersim.org">EMERGENT</a>`, "Emergent is a collection of biologically based 3D neural network models of the brain that power ongoing research in computational cognitive neuroscience.", func(w *core.SVG) {
		errors.Log(w.OpenFS(resources, "emergent-icon.svg"))
	})

	makeBlock(`<a href="https://github.com/WaveELD/WELDBook/blob/main/textmd/ch01_intro.md">WELD</a>`, "WELD is a set of 3D computational models of a new approach to quantum physics based on the de Broglie-Bohm pilot wave theory.", func(w *core.Image) {
		errors.Log(w.OpenFS(resources, "weld-icon.png"))
		w.Styler(func(s *styles.Style) {
			s.Min.Set(units.Dp(256))
		})
	})

	core.NewText(home).SetType(core.TextDisplaySmall).SetText("<b>Why Cogent Core instead of something else?</b>")

	makeBlock(`THE PROBLEM`, "After using other frameworks built on HTML and Qt for years to make various apps ranging from simple tools to complex scientific models, we realized that we were spending more time dealing with excessive boilerplate, browser inconsistencies, and dependency management issues than actual app development.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.Problem)
	})

	makeBlock(`THE SOLUTION`, "We decided to make a framework that would allow us to focus on app content and logic by providing a consistent and elegant API that automatically handles cross-platform support, user customization, and app packaging and deployment.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.Build)
	})

	makeBlock(`THE RESULT`, "Instead of constantly jumping through hoops to create a consistent, easy-to-use, cross-platform app, you can now take advantage of a powerful featureset on all platforms and vastly simplify your development experience.", func(w *core.Icon) {
		initIcon(w).SetIcon(icons.Check)
	})

	tree.AddChild(home, func(w *core.Button) {
		w.SetText("Get started")
		w.OnClick(func(e events.Event) {
			ctx.OpenURL("getting-started")
		})
	})

	return true
}
