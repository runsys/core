# Todo list

This tutorial shows how to make a todo list app using Cogent Core. You should read the [basics](../basics) section if you haven't yet before starting this.

We will represent todo list items using an `item` struct type:

```go
type item struct {
    Done bool
    Task string
}
```

We can create a slice of these items and then represent them with a [[core.Table]] widget:

```Go
type item struct {
    Done bool
    Task string
}
items := []item{{Task: "Code"}, {Task: "Eat"}}
core.NewTable(parent).SetSlice(&items)
```

We can add a button for adding a new todo list item:

```Go
type item struct {
    Done bool
    Task string
}
items := []item{{Task: "Code"}, {Task: "Eat"}}
var table *core.Table
core.NewButton(parent).SetText("Add").SetIcon(icons.Add).OnClick(func(e events.Event) {
    table.SliceNewAt(0)
})
table = core.NewTable(parent).SetSlice(&items)
```
