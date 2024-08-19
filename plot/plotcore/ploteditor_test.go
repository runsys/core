// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotcore

import (
	"log/slog"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/tensor/table"
)

type Data struct {
	City       string
	Population float32
	Area       float32
}

func TestTablePlotEditor(t *testing.T) {
	b := core.NewBody("Plot View")

	epc := table.NewTable("epc")
	epc.OpenCSV("testdata/ra25epoch.tsv", table.Tab)

	pl := NewPlotEditor(b)
	pl.Params.Title = "RA25 Epoch Train"
	pl.Params.XAxisColumn = "Epoch"
	// pl.Params.Scale = 2
	pl.Params.Points = true
	pl.SetTable(epc)
	pl.ColumnParams("UnitErr").On = true
	b.AddAppBar(pl.MakeToolbar)
	b.AssertRender(t, "plotcore_table")
}

func TestSlicePlotEditor(t *testing.T) {
	data := []Data{
		{"Davis", 62000, 500},
		{"Boulder", 85000, 800},
	}

	b := core.NewBody("Plot View")
	dt, err := table.NewSliceTable(data)
	if err != nil {
		slog.Error(err.Error())
	}

	pl := NewPlotEditor(b)
	pl.Params.Title = "Slice Data"
	pl.Params.XAxisColumn = "City"
	pl.Params.Points = true
	pl.SetTable(dt)
	pl.ColumnParams("Population").On = true
	b.AddAppBar(pl.MakeToolbar)

	b.AssertRender(t, "plotcore_slice")
}
