# Images

Cogent Core provides customizable images that allow you to render any image.

You should load images by embedding them so that they work across all platforms:

```go
//go:embed image.png
var myImage embed.FS
```

Then, you can open an image from your embedded filesystem:

```Go
errors.Log(core.NewImage(parent).OpenFS(myImage, "image.png"))
```

You can change the size of an image:

```Go
img := core.NewImage(parent)
errors.Log(img.OpenFS(myImage, "image.png"))
img.Styler(func(s *styles.Style) {
    s.Min.Set(units.Dp(256))
})
```

You can set an image directly to any bounded Go image:

```Go
img := image.NewRGBA(image.Rect(0, 0, 100, 100))
draw.Draw(img, image.Rect(10, 5, 100, 90), colors.Scheme.Warn.Container, image.Point{}, draw.Src)
draw.Draw(img, image.Rect(20, 20, 60, 50), colors.Scheme.Success.Base, image.Point{}, draw.Src)
draw.Draw(img, image.Rect(60, 70, 80, 100), colors.Scheme.Error.Base, image.Point{}, draw.Src)
core.NewImage(parent).SetImage(img)
```

You can also open images directly from the system filesystem, but this is not recommended for images built into your app, since they will end up in a different location on different platforms:

```go
errors.Log(core.NewImage(parent).Open("image.png"))
```
