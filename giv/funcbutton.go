// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giv

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"goki.dev/gi/v2/gi"
	"goki.dev/glop/sentencecase"
	"goki.dev/gti"
	"goki.dev/icons"
	"goki.dev/ki/v2"
)

// NewFuncButton adds to the given parent a new [gi.Button] that is set up to call the given
// function when pressed, using a dialog to prompt the user for any arguments. Also, it sets
// various properties like the name, label, tooltip, and icon of the button based on the
// properties of the function, using reflect and gti. The given function must be registered
// with gti; add a `//gti:add` comment directive and run `goki generate` if you get errors.
// If the given function is a method, both the method and its receiver type must be added to gti.
func NewFuncButton(par ki.Ki, fun any) *gi.Button {
	fnm := gti.FuncName(fun)
	// the "-fm" suffix indicates that it is a method
	if !strings.HasSuffix(fnm, "-fm") {
		f := gti.FuncByName(fnm)
		if f == nil {
			slog.Error("programmer error: cannot use giv.NewFuncButton with a function that has not been added to gti; see the documentation for giv.NewFuncButton", "function", fnm)
			return nil
		}
		return NewFuncButtonImpl(par, f, reflect.ValueOf(fun))
	}

	fnm = strings.TrimSuffix(fnm, "-fm")
	// the last dot separates the function name
	li := strings.LastIndex(fnm, ".")
	metnm := fnm[li+1:]
	typnm := fnm[:li]
	// get rid of any parentheses and pointer receivers
	// that may surround the type name
	typnm = strings.ReplaceAll(typnm, "(*", "")
	typnm = strings.TrimSuffix(typnm, ")")
	gtyp := gti.TypeByName(typnm)
	if gtyp == nil {
		slog.Error("programmer error: cannot use giv.NewFuncButton with a method whose receiver type has not been added to gti; see the documentation for giv.NewFuncButton", "type", typnm, "method", metnm, "fullPath", fnm)
		return nil
	}
	met := gtyp.Methods.ValByKey(metnm)
	if met == nil {
		slog.Error("programmer error: cannot use giv.NewFuncButton with a method that has not been added to gti (even though the receiver type was); see the documentation for giv.NewFuncButton", "type", typnm, "method", metnm, "fullPath", fnm)
		return nil
	}
	return NewMethodButtonImpl(par, met, reflect.ValueOf(fun))
}

// NewFuncButtonImpl is the underlying implementation of [NewFuncButton].
// It should typically not be used by end-user code.
func NewFuncButtonImpl(par ki.Ki, gfun *gti.Func, rfun reflect.Value) *gi.Button {
	// get name without package
	snm := gfun.Name
	li := strings.LastIndex(snm, ".")
	if li >= 0 {
		snm = snm[li+1:] // must also get rid of "."
	}
	bt := gi.NewButton(par, snm).SetText(sentencecase.Of(snm))
	bt.SetTooltip(gfun.Doc)
	// we default to the icon with the same name as
	// the function, if it exists
	ic := icons.Icon(strcase.ToSnake(snm))
	if ic.IsValid() {
		bt.SetIcon(ic)
	}
	return bt
}

// NewMethodButtonImpl is the underlying implementation of [NewFuncButton] for methods.
// It should typically not be used by end-user code.
func NewMethodButtonImpl(par ki.Ki, gmet *gti.Method, rmet reflect.Value) *gi.Button {
	return NewFuncButtonImpl(par, &gti.Func{
		Name:       gmet.Name,
		Doc:        gmet.Doc,
		Directives: gmet.Directives,
		Args:       gmet.Args,
		Returns:    gmet.Returns,
	}, rmet)
}
