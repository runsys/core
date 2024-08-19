// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"strconv"
	"testing"

	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
)

const (
	nButtons = 100
	nUpdates = 100
)

// This set of benchmarks highlights the negative performance impacts of an inefficient
// fully declarative updating approach relative to a [Plan] based updating approach.

func BenchmarkUpdateDeclarative(b *testing.B) {
	for range b.N {
		for range nUpdates {
			b := NewBody()
			for i := range nButtons {
				NewButton(b).SetText(strconv.Itoa(i)).SetIcon(icons.Download).Styler(func(s *styles.Style) {
					s.Min.Set(units.Em(5))
					s.CenterAll()
				})
			}
			for i := range nButtons {
				bt := b.Child(i)
				bt.CopyFieldsFrom(bt)
			}
		}
	}
}

func BenchmarkUpdatePlan(bn *testing.B) {
	b := NewBody()
	for range bn.N {
		for range nUpdates {
			p := &tree.Plan{}
			for i := range nButtons {
				tree.AddAt(p, strconv.Itoa(i), func(w *Button) {
					w.SetText(strconv.Itoa(i)).SetIcon(icons.Download).Styler(func(s *styles.Style) {
						s.Min.Set(units.Em(5))
						s.CenterAll()
					})
				})
			}
			p.Update(b)
		}
	}
}
