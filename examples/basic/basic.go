package main

import (
	"cogentcore.org/core/core"
)

func main() {
<<<<<<< HEAD
	b := core.NewBody("Hello")
	core.NewButton(b).SetText("Hello, World!工在")
=======
	b := core.NewBody()
	core.NewButton(b).SetText("Hello, World!")
>>>>>>> upstream/main
	b.RunMainWindow()
}
