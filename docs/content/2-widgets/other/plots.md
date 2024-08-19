# Plots

Cogent Core provides interactive and customizable data plots.

You can make an interactive plot from slice data:

```Go
type Data struct {
	Time       float32
	Population float32
}
data := []Data{
    {0, 500},
    {1, 800},
    {2, 1600},
    {3, 1400},
}
dt := errors.Log1(table.NewSliceTable(data))
pe := plotcore.NewPlotEditor(parent).SetTable(dt)
pe.Params.XAxisColumn = "Time"
pe.ColumnParams("Population").On = true
```
