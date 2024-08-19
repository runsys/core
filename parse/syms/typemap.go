// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syms

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// TypeMap is a map of types for quick looking up by name
type TypeMap map[string]*Type

// Alloc ensures that map is made
func (tm *TypeMap) Alloc() {
	if *tm == nil {
		*tm = make(TypeMap)
	}
}

// Add adds a type to the map, handling allocation for nil maps
func (tm *TypeMap) Add(ty *Type) {
	tm.Alloc()
	(*tm)[ty.Name] = ty
}

// CopyFrom copies all the types from given source map into this one.
// Types that have Kind != Unknown are retained over unknown ones.
// srcIsNewer means that the src type is newer and should be used for
// other data like source
func (tm *TypeMap) CopyFrom(src TypeMap, srcIsNewer bool) {
	tm.Alloc()
	for nm, sty := range src {
		ety, has := (*tm)[nm]
		if !has {
			(*tm)[nm] = sty
			continue
		}
		if srcIsNewer {
			ety.CopyFromSrc(sty)
		} else {
			sty.CopyFromSrc(ety)
		}
		if ety.Kind != Unknown {
			// if len(sty.Els) > 0 && len(sty.Els) != len(ety.Els) {
			// 	fmt.Printf("TypeMap dupe type: %v existing kind: %v els: %v new els: %v  kind: %v\n", nm, ety.Kind, len(ety.Els), len(sty.Els), sty.Kind)
			// }
		} else if sty.Kind != Unknown {
			// if len(ety.Els) > 0 && len(sty.Els) != len(ety.Els) {
			// 	fmt.Printf("TypeMap dupe type: %v new kind: %v els: %v existing els: %v\n", nm, sty.Kind, len(sty.Els), len(ety.Els))
			// }
			(*tm)[nm] = sty
			// } else {
			// 	fmt.Printf("TypeMap dupe type: %v both unknown: existing els: %v new els: %v\n", nm, len(ety.Els), len(sty.Els))
		}
	}
}

// Clone returns deep copy of this type map -- types are Clone() copies.
// returns nil if this map is empty
func (tm *TypeMap) Clone() TypeMap {
	sz := len(*tm)
	if sz == 0 {
		return nil
	}
	ntm := make(TypeMap, sz)
	for nm, sty := range *tm {
		ntm[nm] = sty.Clone()
	}
	return ntm
}

// Names returns a slice of the names in this map, optionally sorted
func (tm *TypeMap) Names(sorted bool) []string {
	nms := make([]string, len(*tm))
	idx := 0
	for _, ty := range *tm {
		nms[idx] = ty.Name
		idx++
	}
	if sorted {
		sort.StringSlice(nms).Sort()
	}
	return nms
}

// KindNames returns a slice of the kind:names in this map, optionally sorted
func (tm *TypeMap) KindNames(sorted bool) []string {
	nms := make([]string, len(*tm))
	idx := 0
	for _, ty := range *tm {
		nms[idx] = ty.Kind.String() + ":" + ty.Name
		idx++
	}
	if sorted {
		sort.StringSlice(nms).Sort()
	}
	return nms
}

// PrintUnknowns prints all the types that have a Kind = Unknown
// indicates an error in type resolution
func (tm *TypeMap) PrintUnknowns() {
	for _, ty := range *tm {
		if ty.Kind != Unknown {
			continue
		}
		fmt.Printf("Type named: %v has Kind = Unknown!  From file: %v:%v\n", ty.Name, ty.Filename, ty.Region)
	}
}

// WriteDoc writes basic doc info, sorted by kind and name
func (tm *TypeMap) WriteDoc(out io.Writer, depth int) {
	nms := tm.KindNames(true)
	for _, nm := range nms {
		ci := strings.Index(nm, ":")
		ty := (*tm)[nm[ci+1:]]
		ty.WriteDoc(out, depth)
	}
}

// TypeKindSize is used for initialization of builtin typemaps
type TypeKindSize struct {
	Name string
	Kind Kinds
	Size int
}
