Cogent Core provides switches for selecting one or more options out of a list of items, in addition to standalone switches for controlling a single bool value.

You can make a standalone switch with no label:

```Go
core.NewSwitch(b)
```

You can add a label to a standalone switch:

```Go
core.NewSwitch(b).SetText("Remember me")
```

You can make a standalone switch render as a checkbox:

```Go
core.NewSwitch(b).SetType(core.SwitchCheckbox).SetText("Remember me")
```

You can make a standalone switch render as a radio button:

```Go
core.NewSwitch(b).SetType(core.SwitchRadioButton).SetText("Remember me")
```

You can detect when the user changes whether the switch is checked:

```Go
sw := core.NewSwitch(b).SetText("Remember me")
sw.OnChange(func(e events.Event) {
    core.MessageSnackbar(sw, fmt.Sprintf("Switch is %v", sw.IsChecked()))
})
```

You can make a group of switches from a list of strings:

```Go
core.NewSwitches(b).SetStrings("Go", "Python", "C++")
```

If you need to customize the items more, you can use a list of [[core.SwitchItem]] objects:

```Go
core.NewSwitches(b).SetItems(
    core.SwitchItem{Value: "Go", Tooltip: "Elegant, fast, and easy-to-use"},
    core.SwitchItem{Value: "Python", Tooltip: "Slow and duck-typed"},
    core.SwitchItem{Value: "C++", Tooltip: "Hard to use and slow to compile"},
)
```

You can make switches mutually exclusive so that only one can be selected at a time:

```Go
core.NewSwitches(b).SetMutex(true).SetStrings("Go", "Python", "C++")
```

You can make switches render as chips:

```Go
core.NewSwitches(b).SetType(core.SwitchChip).SetStrings("Go", "Python", "C++")
```

You can make switches render as checkboxes:

```Go
core.NewSwitches(b).SetType(core.SwitchCheckbox).SetStrings("Go", "Python", "C++")
```

You can make switches render as radio buttons:

```Go
core.NewSwitches(b).SetType(core.SwitchRadioButton).SetStrings("Go", "Python", "C++")
```

You can make switches render as segmented buttons:

```Go
core.NewSwitches(b).SetType(core.SwitchSegmentedButton).SetStrings("Go", "Python", "C++")
```

You can make switches render vertically:

```Go
core.NewSwitches(b).SetStrings("Go", "Python", "C++").Styler(func(s *styles.Style) {
    s.Direction = styles.Column
})
```

You can detect when the user changes which switches are selected:

```Go
sw := core.NewSwitches(b).SetStrings("Go", "Python", "C++")
sw.OnChange(func(e events.Event) {
    core.MessageSnackbar(sw, fmt.Sprintf("Currently selected: %v", sw.SelectedItems()))
})
```
