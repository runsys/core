Cogent Core provides interactive color pickers that allow users to input colors using three sliders for the components of the HCT color system: hue, chroma, and tone. The tooltip for each slider contains more information about each component.

You can make a color picker and set its starting color to any color:

```Go
core.NewColorPicker(b).SetColor(colors.Orange)
```

You can detect when the user changes the color:

```Go
cp := core.NewColorPicker(b).SetColor(colors.Green)
cp.OnChange(func(e events.Event) {
    core.MessageSnackbar(cp, colors.AsHex(cp.Color))
})
```

You can make a button that opens a color picker dialog:

```Go
core.NewColorButton(b).SetColor(colors.Purple)
```

You can detect when the user changes the color using the dialog:

```Go
cb := core.NewColorButton(b).SetColor(colors.Gold)
cb.OnChange(func(e events.Event) {
    core.MessageSnackbar(cb, colors.AsHex(cb.Color))
})
```
