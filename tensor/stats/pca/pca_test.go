// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pca

import (
	"fmt"
	"math"
	"testing"

	"cogentcore.org/core/tensor/stats/metric"
	"cogentcore.org/core/tensor/table"
)

func TestPCAIris(t *testing.T) {
	// note: these results are verified against this example:
	// https://plot.ly/ipython-notebooks/principal-component-analysis/

	dt := table.NewTable()
	dt.AddFloat64TensorColumn("data", []int{4})
	dt.AddStringColumn("class")
	err := dt.OpenCSV("testdata/iris.data", table.Comma)
	if err != nil {
		t.Error(err)
	}
	ix := table.NewIndexView(dt)
	pc := &PCA{}
	// pc.TableCol(ix, "data", metric.Covariance64)
	// fmt.Printf("covar: %v\n", pc.Covar)
	err = pc.TableCol(ix, "data", metric.Correlation64)
	if err != nil {
		t.Error(err)
	}
	// fmt.Printf("correl: %v\n", pc.Covar)
	// fmt.Printf("correl vec: %v\n", pc.Vectors)
	// fmt.Printf("correl val: %v\n", pc.Values)

	errtol := 1.0e-9
	corvals := []float64{0.020607707235624825, 0.14735327830509573, 0.9212209307072254, 2.910818083752054}
	for i, v := range pc.Values {
		dif := math.Abs(corvals[i] - v)
		if dif > errtol {
			err = fmt.Errorf("eigenvalue: %v  differs from correct: %v  was:  %v", i, corvals[i], v)
			t.Error(err)
		}
	}

	prjt := &table.Table{}
	err = pc.ProjectColToTable(prjt, ix, "data", "class", []int{0, 1})
	if err != nil {
		t.Error(err)
	}
	// prjt.SaveCSV("test_data/projection01.csv", table.Comma, true)
}
