# Events

Cogent Core provides a robust and easy-to-use event handling system that allows you to handle various events on widgets. To handle an event, simply call the `On{EventType}` method on any widget. For example:

```Go
core.NewButton(parent).SetText("Click me!").OnClick(func(e events.Event) {
    core.MessageSnackbar(parent, "Button clicked")
})
```

The [[events.Event]] object passed to the function can be used for various things like obtaining detailed event information. For example, you can determine the exact position of a click event:

```Go
core.NewButton(parent).SetText("Click me!").OnClick(func(e events.Event) {
    core.MessageSnackbar(parent, fmt.Sprint("Button clicked at ", e.Pos()))
})
```

You can see the [advanced events page](../advanced/events) for more information if you need it.
