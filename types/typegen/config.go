// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typegen

import (
	"text/template"

	"cogentcore.org/core/base/ordmap"
)

// Config contains the configuration information
// used by typegen
type Config struct { //types:add

	// the source directory to run typegen on (can be set to multiple through paths like ./...)
	Dir string `default:"." posarg:"0" required:"-"`

	// the output file location relative to the package on which typegen is being called
	Output string `default:"typegen.go"`

	// whether to add types to typegen by default
	AddTypes bool

	// whether to add methods to typegen by default
	AddMethods bool

	// whether to add functions to typegen by default
	AddFuncs bool

	// An ordered map of configs keyed by fully qualified interface type names; if a type implements the interface, the config will be applied to it.
	// The configs are applied in sequential ascending order, which means that
	// the last config overrides the other ones, so the most specific
	// interfaces should typically be put last.
	// Note: the package typegen is run on must explicitly reference this interface at some point for this to work; adding a simple
	// `var _ MyInterface = (*MyType)(nil)` statement to check for interface implementation is an easy way to accomplish that.
	// Note: typegen will still succeed if it can not find one of the interfaces specified here in order to allow it to work generically across multiple directories; you can use the -v flag to get log warnings about this if you suspect that it is not finding interfaces when it should.
	InterfaceConfigs *ordmap.Map[string, *Config]

	// Whether to generate chaining `Set*` methods for each exported field of each type (eg: "SetText" for field "Text").
	// If this is set to true, then you can add `set:"-"` struct tags to individual fields
	// to prevent Set methods being generated for them.
	Setters bool

	// TODO: should this be called TypeTemplates and should there be a Func/Method Templates?

	// a slice of templates to execute on each type being added; the template data is of the type typegen.Type
	Templates []*template.Template
}
