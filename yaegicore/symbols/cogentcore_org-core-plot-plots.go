// Code generated by 'yaegi extract cogentcore.org/core/plot/plots'. DO NOT EDIT.

package symbols

import (
	"cogentcore.org/core/plot/plots"
	"reflect"
)

func init() {
	Symbols["cogentcore.org/core/plot/plots/plots"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"AddTableLine":       reflect.ValueOf(plots.AddTableLine),
		"AddTableLinePoints": reflect.ValueOf(plots.AddTableLinePoints),
		"Box":                reflect.ValueOf(plots.Box),
		"Circle":             reflect.ValueOf(plots.Circle),
		"Cross":              reflect.ValueOf(plots.Cross),
		"DrawBox":            reflect.ValueOf(plots.DrawBox),
		"DrawCircle":         reflect.ValueOf(plots.DrawCircle),
		"DrawCross":          reflect.ValueOf(plots.DrawCross),
		"DrawPlus":           reflect.ValueOf(plots.DrawPlus),
		"DrawPyramid":        reflect.ValueOf(plots.DrawPyramid),
		"DrawRing":           reflect.ValueOf(plots.DrawRing),
		"DrawShape":          reflect.ValueOf(plots.DrawShape),
		"DrawSquare":         reflect.ValueOf(plots.DrawSquare),
		"DrawTriangle":       reflect.ValueOf(plots.DrawTriangle),
		"MidStep":            reflect.ValueOf(plots.MidStep),
		"NewBarChart":        reflect.ValueOf(plots.NewBarChart),
		"NewLabels":          reflect.ValueOf(plots.NewLabels),
		"NewLine":            reflect.ValueOf(plots.NewLine),
		"NewLinePoints":      reflect.ValueOf(plots.NewLinePoints),
		"NewScatter":         reflect.ValueOf(plots.NewScatter),
		"NewTableXYer":       reflect.ValueOf(plots.NewTableXYer),
		"NewXErrorBars":      reflect.ValueOf(plots.NewXErrorBars),
		"NewYErrorBars":      reflect.ValueOf(plots.NewYErrorBars),
		"NoStep":             reflect.ValueOf(plots.NoStep),
		"Plus":               reflect.ValueOf(plots.Plus),
		"PostStep":           reflect.ValueOf(plots.PostStep),
		"PreStep":            reflect.ValueOf(plots.PreStep),
		"Pyramid":            reflect.ValueOf(plots.Pyramid),
		"Ring":               reflect.ValueOf(plots.Ring),
		"ShapesN":            reflect.ValueOf(plots.ShapesN),
		"ShapesValues":       reflect.ValueOf(plots.ShapesValues),
		"Square":             reflect.ValueOf(plots.Square),
		"StepKindN":          reflect.ValueOf(plots.StepKindN),
		"StepKindValues":     reflect.ValueOf(plots.StepKindValues),
		"TableColumnIndex":   reflect.ValueOf(plots.TableColumnIndex),
		"Triangle":           reflect.ValueOf(plots.Triangle),

		// type definitions
		"BarChart":   reflect.ValueOf((*plots.BarChart)(nil)),
		"Errors":     reflect.ValueOf((*plots.Errors)(nil)),
		"Labels":     reflect.ValueOf((*plots.Labels)(nil)),
		"Line":       reflect.ValueOf((*plots.Line)(nil)),
		"Scatter":    reflect.ValueOf((*plots.Scatter)(nil)),
		"Shapes":     reflect.ValueOf((*plots.Shapes)(nil)),
		"StepKind":   reflect.ValueOf((*plots.StepKind)(nil)),
		"Table":      reflect.ValueOf((*plots.Table)(nil)),
		"TableXYer":  reflect.ValueOf((*plots.TableXYer)(nil)),
		"XErrorBars": reflect.ValueOf((*plots.XErrorBars)(nil)),
		"XErrorer":   reflect.ValueOf((*plots.XErrorer)(nil)),
		"XErrors":    reflect.ValueOf((*plots.XErrors)(nil)),
		"XYLabeller": reflect.ValueOf((*plots.XYLabeler)(nil)),
		"XYLabels":   reflect.ValueOf((*plots.XYLabels)(nil)),
		"YErrorBars": reflect.ValueOf((*plots.YErrorBars)(nil)),
		"YErrorer":   reflect.ValueOf((*plots.YErrorer)(nil)),
		"YErrors":    reflect.ValueOf((*plots.YErrors)(nil)),

		// interface wrapper definitions
		"_Table":      reflect.ValueOf((*_cogentcore_org_core_plot_plots_Table)(nil)),
		"_XErrorer":   reflect.ValueOf((*_cogentcore_org_core_plot_plots_XErrorer)(nil)),
		"_XYLabeller": reflect.ValueOf((*_cogentcore_org_core_plot_plots_XYLabeller)(nil)),
		"_YErrorer":   reflect.ValueOf((*_cogentcore_org_core_plot_plots_YErrorer)(nil)),
	}
}

// _cogentcore_org_core_plot_plots_Table is an interface wrapper for Table type
type _cogentcore_org_core_plot_plots_Table struct {
	IValue      interface{}
	WColumnName func(i int) string
	WNumColumns func() int
	WNumRows    func() int
	WPlotData   func(column int, row int) float32
}

func (W _cogentcore_org_core_plot_plots_Table) ColumnName(i int) string { return W.WColumnName(i) }
func (W _cogentcore_org_core_plot_plots_Table) NumColumns() int         { return W.WNumColumns() }
func (W _cogentcore_org_core_plot_plots_Table) NumRows() int            { return W.WNumRows() }
func (W _cogentcore_org_core_plot_plots_Table) PlotData(column int, row int) float32 {
	return W.WPlotData(column, row)
}

// _cogentcore_org_core_plot_plots_XErrorer is an interface wrapper for XErrorer type
type _cogentcore_org_core_plot_plots_XErrorer struct {
	IValue  interface{}
	WXError func(i int) (low float32, high float32)
}

func (W _cogentcore_org_core_plot_plots_XErrorer) XError(i int) (low float32, high float32) {
	return W.WXError(i)
}

// _cogentcore_org_core_plot_plots_XYLabeller is an interface wrapper for XYLabeller type
type _cogentcore_org_core_plot_plots_XYLabeller struct {
	IValue interface{}
	WLabel func(i int) string
	WLen   func() int
	WXY    func(i int) (x float32, y float32)
}

func (W _cogentcore_org_core_plot_plots_XYLabeller) Label(i int) string              { return W.WLabel(i) }
func (W _cogentcore_org_core_plot_plots_XYLabeller) Len() int                        { return W.WLen() }
func (W _cogentcore_org_core_plot_plots_XYLabeller) XY(i int) (x float32, y float32) { return W.WXY(i) }

// _cogentcore_org_core_plot_plots_YErrorer is an interface wrapper for YErrorer type
type _cogentcore_org_core_plot_plots_YErrorer struct {
	IValue  interface{}
	WYError func(i int) (float32, float32)
}

func (W _cogentcore_org_core_plot_plots_YErrorer) YError(i int) (float32, float32) {
	return W.WYError(i)
}
