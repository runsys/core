// Code generated by 'yaegi extract cogentcore.org/core/plot'. DO NOT EDIT.

package symbols

import (
	"cogentcore.org/core/plot"
	"reflect"
)

func init() {
	Symbols["cogentcore.org/core/plot/plot"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"CheckFloats":       reflect.ValueOf(plot.CheckFloats),
		"CheckNaNs":         reflect.ValueOf(plot.CheckNaNs),
		"CopyValues":        reflect.ValueOf(plot.CopyValues),
		"CopyXYZs":          reflect.ValueOf(plot.CopyXYZs),
		"CopyXYs":           reflect.ValueOf(plot.CopyXYs),
		"DefaultFontFamily": reflect.ValueOf(&plot.DefaultFontFamily).Elem(),
		"ErrInfinity":       reflect.ValueOf(&plot.ErrInfinity).Elem(),
		"ErrNoData":         reflect.ValueOf(&plot.ErrNoData).Elem(),
		"New":               reflect.ValueOf(plot.New),
		"PlotXYs":           reflect.ValueOf(plot.PlotXYs),
		"Range":             reflect.ValueOf(plot.Range),
		"UTCUnixTime":       reflect.ValueOf(&plot.UTCUnixTime).Elem(),
		"UnixTimeIn":        reflect.ValueOf(plot.UnixTimeIn),
		"XYRange":           reflect.ValueOf(plot.XYRange),

		// type definitions
		"Axis":           reflect.ValueOf((*plot.Axis)(nil)),
		"ConstantTicks":  reflect.ValueOf((*plot.ConstantTicks)(nil)),
		"DataRanger":     reflect.ValueOf((*plot.DataRanger)(nil)),
		"DefaultTicks":   reflect.ValueOf((*plot.DefaultTicks)(nil)),
		"InvertedScale":  reflect.ValueOf((*plot.InvertedScale)(nil)),
		"Labeller":       reflect.ValueOf((*plot.Labeler)(nil)),
		"Legend":         reflect.ValueOf((*plot.Legend)(nil)),
		"LegendEntry":    reflect.ValueOf((*plot.LegendEntry)(nil)),
		"LegendPosition": reflect.ValueOf((*plot.LegendPosition)(nil)),
		"LineStyle":      reflect.ValueOf((*plot.LineStyle)(nil)),
		"LinearScale":    reflect.ValueOf((*plot.LinearScale)(nil)),
		"LogScale":       reflect.ValueOf((*plot.LogScale)(nil)),
		"LogTicks":       reflect.ValueOf((*plot.LogTicks)(nil)),
		"Normalizer":     reflect.ValueOf((*plot.Normalizer)(nil)),
		"Plot":           reflect.ValueOf((*plot.Plot)(nil)),
		"Plotter":        reflect.ValueOf((*plot.Plotter)(nil)),
		"Text":           reflect.ValueOf((*plot.Text)(nil)),
		"TextStyle":      reflect.ValueOf((*plot.TextStyle)(nil)),
		"Thumbnailer":    reflect.ValueOf((*plot.Thumbnailer)(nil)),
		"Tick":           reflect.ValueOf((*plot.Tick)(nil)),
		"Ticker":         reflect.ValueOf((*plot.Ticker)(nil)),
		"TickerFunc":     reflect.ValueOf((*plot.TickerFunc)(nil)),
		"TimeTicks":      reflect.ValueOf((*plot.TimeTicks)(nil)),
		"Valuer":         reflect.ValueOf((*plot.Valuer)(nil)),
		"Values":         reflect.ValueOf((*plot.Values)(nil)),
		"XValues":        reflect.ValueOf((*plot.XValues)(nil)),
		"XYValues":       reflect.ValueOf((*plot.XYValues)(nil)),
		"XYZ":            reflect.ValueOf((*plot.XYZ)(nil)),
		"XYZer":          reflect.ValueOf((*plot.XYZer)(nil)),
		"XYZs":           reflect.ValueOf((*plot.XYZs)(nil)),
		"XYer":           reflect.ValueOf((*plot.XYer)(nil)),
		"XYs":            reflect.ValueOf((*plot.XYs)(nil)),
		"YValues":        reflect.ValueOf((*plot.YValues)(nil)),

		// interface wrapper definitions
		"_DataRanger":  reflect.ValueOf((*_cogentcore_org_core_plot_DataRanger)(nil)),
		"_Labeller":    reflect.ValueOf((*_cogentcore_org_core_plot_Labeller)(nil)),
		"_Normalizer":  reflect.ValueOf((*_cogentcore_org_core_plot_Normalizer)(nil)),
		"_Plotter":     reflect.ValueOf((*_cogentcore_org_core_plot_Plotter)(nil)),
		"_Thumbnailer": reflect.ValueOf((*_cogentcore_org_core_plot_Thumbnailer)(nil)),
		"_Ticker":      reflect.ValueOf((*_cogentcore_org_core_plot_Ticker)(nil)),
		"_Valuer":      reflect.ValueOf((*_cogentcore_org_core_plot_Valuer)(nil)),
		"_XYZer":       reflect.ValueOf((*_cogentcore_org_core_plot_XYZer)(nil)),
		"_XYer":        reflect.ValueOf((*_cogentcore_org_core_plot_XYer)(nil)),
	}
}

// _cogentcore_org_core_plot_DataRanger is an interface wrapper for DataRanger type
type _cogentcore_org_core_plot_DataRanger struct {
	IValue     interface{}
	WDataRange func() (xmin float32, xmax float32, ymin float32, ymax float32)
}

func (W _cogentcore_org_core_plot_DataRanger) DataRange() (xmin float32, xmax float32, ymin float32, ymax float32) {
	return W.WDataRange()
}

// _cogentcore_org_core_plot_Labeller is an interface wrapper for Labeller type
type _cogentcore_org_core_plot_Labeller struct {
	IValue interface{}
	WLabel func(i int) string
}

func (W _cogentcore_org_core_plot_Labeller) Label(i int) string { return W.WLabel(i) }

// _cogentcore_org_core_plot_Normalizer is an interface wrapper for Normalizer type
type _cogentcore_org_core_plot_Normalizer struct {
	IValue     interface{}
	WNormalize func(min float32, max float32, x float32) float32
}

func (W _cogentcore_org_core_plot_Normalizer) Normalize(min float32, max float32, x float32) float32 {
	return W.WNormalize(min, max, x)
}

// _cogentcore_org_core_plot_Plotter is an interface wrapper for Plotter type
type _cogentcore_org_core_plot_Plotter struct {
	IValue  interface{}
	WPlot   func(pt *plot.Plot)
	WXYData func() (data plot.XYer, pixels plot.XYer)
}

func (W _cogentcore_org_core_plot_Plotter) Plot(pt *plot.Plot) { W.WPlot(pt) }
func (W _cogentcore_org_core_plot_Plotter) XYData() (data plot.XYer, pixels plot.XYer) {
	return W.WXYData()
}

// _cogentcore_org_core_plot_Thumbnailer is an interface wrapper for Thumbnailer type
type _cogentcore_org_core_plot_Thumbnailer struct {
	IValue     interface{}
	WThumbnail func(pt *plot.Plot)
}

func (W _cogentcore_org_core_plot_Thumbnailer) Thumbnail(pt *plot.Plot) { W.WThumbnail(pt) }

// _cogentcore_org_core_plot_Ticker is an interface wrapper for Ticker type
type _cogentcore_org_core_plot_Ticker struct {
	IValue interface{}
	WTicks func(min float32, max float32) []plot.Tick
}

func (W _cogentcore_org_core_plot_Ticker) Ticks(min float32, max float32) []plot.Tick {
	return W.WTicks(min, max)
}

// _cogentcore_org_core_plot_Valuer is an interface wrapper for Valuer type
type _cogentcore_org_core_plot_Valuer struct {
	IValue interface{}
	WLen   func() int
	WValue func(i int) float32
}

func (W _cogentcore_org_core_plot_Valuer) Len() int            { return W.WLen() }
func (W _cogentcore_org_core_plot_Valuer) Value(i int) float32 { return W.WValue(i) }

// _cogentcore_org_core_plot_XYZer is an interface wrapper for XYZer type
type _cogentcore_org_core_plot_XYZer struct {
	IValue interface{}
	WLen   func() int
	WXY    func(i int) (float32, float32)
	WXYZ   func(i int) (float32, float32, float32)
}

func (W _cogentcore_org_core_plot_XYZer) Len() int                              { return W.WLen() }
func (W _cogentcore_org_core_plot_XYZer) XY(i int) (float32, float32)           { return W.WXY(i) }
func (W _cogentcore_org_core_plot_XYZer) XYZ(i int) (float32, float32, float32) { return W.WXYZ(i) }

// _cogentcore_org_core_plot_XYer is an interface wrapper for XYer type
type _cogentcore_org_core_plot_XYer struct {
	IValue interface{}
	WLen   func() int
	WXY    func(i int) (x float32, y float32)
}

func (W _cogentcore_org_core_plot_XYer) Len() int                        { return W.WLen() }
func (W _cogentcore_org_core_plot_XYer) XY(i int) (x float32, y float32) { return W.WXY(i) }
